package main

import (
	"poetrade47/src/config"
	"poetrade47/src/helpers"

	"sync"
	"time"

	"github.com/go-vgo/robotgo"
)

type ActionType string
type CommandManager struct {
	mu     sync.RWMutex
	macros map[string]FlexMacro
}

type FlexMacro struct {
	cfg     config.FlexMacroSettings
	Type    ActionType `json:"type"`
	DelayMS int        `json:"delay_ms"`
}

var GlobalCommandManager = &CommandManager{
	macros: make(map[string]FlexMacro),
}

const (
	ActionTextCommand ActionType = "TextCommand" // '/hideout'
	ActionKeySequence ActionType = "KeySequence" // 1, 2, 3
)

func (cm *CommandManager) Remove(macro FlexMacro) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Saftey net
	if cm.macros == nil {
		return
	}

	lookupKey := macro.ToLookupKey()
	delete(cm.macros, lookupKey)
	helpers.DebugLog("Removed macro under lookupkey: %s", lookupKey)
}

// Checks if the command exists and executes it. Returns false if it doesn't exist
func (cm *CommandManager) Lookup(lookupKey string) (FlexMacro, bool) {
	// helpers.DebugLog("Handling input for key '%s'", lookupKey)
	cm.mu.RLock()
	cmd, exists := cm.macros[lookupKey]
	cm.mu.RUnlock()

	if !exists || !cmd.cfg.Enabled {
		return FlexMacro{}, false
	}

	return cmd, true
}

func (cm *CommandManager) SyncMacros(entries []config.FlexMacroSettings) {
	helpers.DebugLog("Updating CommandManager data...")
	// entries is the UI entry that is in the custom list
	// for each entry data,
	// grab the keybind button's data
	// grab the text label data
	// build FlexMacro
	// Store flexMacro

	// Tear down listeners
	//cm.clearActiveHooks()
	newMacros := make(map[string]FlexMacro, len(entries))

	for _, entry := range entries {
		// build modifiers + key string for the map key
		mapKey := helpers.MakeLookupKey(entry.Modifiers, entry.Key)

		if mapKey == "" {
			continue
		}

		newMacros[mapKey] = FlexMacro{
			cfg:     entry,
			Type:    ActionTextCommand,
			DelayMS: 1,
		}
	}

	cm.mu.Lock()
	cm.macros = newMacros
	cm.mu.Unlock()
}

func (m *FlexMacro) ExecuteMacro() {
	if !m.cfg.Enabled {
		return
	}

	switch m.Type {
	case ActionTextCommand:
		// ie. /hideout
		if err := robotgo.KeyTap("enter"); err != nil {
			helpers.DebugLog("Could not open game chat: %v", err)
			return
		}

		// Clear any text already in the chat input.
		if err := robotgo.KeyTap("a", "ctrl"); err != nil {
			helpers.DebugLog("Could not select game chat text: %v", err)
			return
		}
		if err := robotgo.KeyTap("backspace"); err != nil {
			helpers.DebugLog("Could not clear game chat text: %v", err)
			return
		}

		time.Sleep(5 * time.Millisecond)
		robotgo.Type(m.cfg.Payload)
		time.Sleep(5 * time.Millisecond)
		if err := robotgo.KeyTap("enter"); err != nil {
			helpers.DebugLog("Could not submit game command: %v", err)
		}
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
	return helpers.MakeLookupKey(m.cfg.Modifiers, m.cfg.Key)
}
