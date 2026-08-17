package main

import (
	"poetrade47/src/config"
	"poetrade47/src/helpers"

	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-vgo/robotgo"
)

type ActionType string
type CommandManager struct {
	mu     sync.RWMutex
	macros map[string]FlexMacro
}

/*
Name      string            `json:"name"`
Modifiers []string          `json:"modifiers"`
Key       string            `json:"key"`
Type      config.ActionType `json:"type"`
Payload   string            `json:"payload"`
DelayMS   int               `json:"delay_ms"`
Enabled   bool              `json:"enabled"`
*/
type FlexMacro struct {
	cfg     config.FlexMacroSettings
	Type    ActionType `json:"type"`
	DelayMS int        `json:"delay_ms"`
}

var GlobalCommandManager = &CommandManager{
	macros: make(map[string]FlexMacro),
}

var (
	modMu              sync.RWMutex
	activeModsMap      = make(map[string]bool)
	SupportedModifiers = []string{helpers.ModCtrl, helpers.ModShift, helpers.ModAlt}
	isSupportedModMap  = map[string]bool{"ctrl": true, "shift": true, "alt": true}
)

const (
	ActionTextCommand ActionType = "TextCommand" // '/hideout'
	ActionKeySequence ActionType = "KeySequence" // 1, 2, 3
)

// func (cm *CommandManager) _() is saying that there is a CommandManager that
// 		will call this function AND use it's pointer (only 1)

func (cm *CommandManager) Remove(macro FlexMacro) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Saftey net
	if cm.macros == nil {
		return
	}

	lookupKey := macro.ToLookupKey()
	delete(cm.macros, macro.ToLookupKey())
	helpers.DebugLog("Removed macro under lookupkey: %s", lookupKey)
}

// Checks if the command exists and executes it. Returns false if it doesn't exist
func (cm *CommandManager) ExecuteIfMapped(lookupKey string) bool {
	// helpers.DebugLog("Handling input for key '%s'", lookupKey)
	cm.mu.RLock()
	cmd, exists := cm.macros[lookupKey]
	cm.mu.RUnlock()

	if !exists || !cmd.cfg.Enabled {
		return false
	}

	cmd.ExecuteMacro()

	return true
}

func (cm *CommandManager) SyncMacros(entries []config.FlexMacroSettings) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	helpers.DebugLog("Updating CommandManager data...")
	// entries is the UI entry that is in the custom list
	// for each entry data,
	// grab the keybind button's data
	// grab the text label data
	// build FlexMacro
	// Store flexMacro

	// Tear down listeners
	//cm.clearActiveHooks()

	// clear out existing macros if it exists, make otherwise
	if cm.macros == nil {
		cm.macros = make(map[string]FlexMacro)
	} else {
		clear(cm.macros)
	}

	for _, entry := range entries {
		// build modifiers + key string for the map key
		mapKey := ""
		if len(entry.Modifiers) > 0 {
			mapKey = strings.Join(entry.Modifiers, "+") + "+" + entry.Key
		} else {
			mapKey = entry.Key
		}

		if mapKey == "" {
			continue
		}

		cm.macros[mapKey] = FlexMacro{
			cfg:     entry,
			Type:    ActionTextCommand,
			DelayMS: 1,
		}
	}

}

func (m *FlexMacro) ExecuteMacro() {
	if !m.cfg.Enabled {
		return
	}

	switch m.Type {
	case ActionTextCommand:
		// ie. /hideout
		robotgo.KeyTap("enter")
		time.Sleep(20 * time.Millisecond)
		robotgo.Type(m.cfg.Payload)
		time.Sleep(20 * time.Millisecond)
		robotgo.KeyTap("enter")
	case ActionKeySequence:
		// ie. 1,2,3 for flasks
		keys := helpers.ParseKeys(m.cfg.Payload)

		for _, k := range keys {
			robotgo.KeyTap(k)
			if m.DelayMS > 0 {
				robotgo.MilliSleep(m.DelayMS)
			}
		}
	}
	helpers.DebugLog("Attempting to execute macro: %+v", m)
}

func (m FlexMacro) ToLookupKey() string {
	if len(m.cfg.Modifiers) == 0 {
		return m.cfg.Key // e.g., "F2"
	}

	// Copy and sort to guarantee consistent ordering regardless of how it's stored
	mods := make([]string, len(m.cfg.Modifiers))
	copy(mods, m.cfg.Modifiers)
	sort.Strings(mods)

	return strings.Join(mods, "+") + "+" + m.cfg.Key // e.g., "Ctrl+Shift+F2"
}
