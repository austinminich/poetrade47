package main

import (
	"poetrade47/src/helpers"
	"poetrade47/src/ui"
	"sync"
)

type CommandManager struct {
	mu     sync.RWMutex
	macros map[string]FlexMacro
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

type ActionType string

const (
	ActionTextCommand ActionType = "TextCommand" // '/hideout'
	ActionKeySequence ActionType = "KeySequence" // 1, 2, 3
)

type FlexMacro struct {
	Name      string     `json:"name"`
	Modifiers []string   `json:"modifiers"`
	Key       string     `json:"key"`
	Type      ActionType `json:"type"`
	Payload   string     `json:"payload"`
	DelayMS   int        `json:"delay_ms"`
	Enabled   bool       `json:"enabled"`
}

// func (cm *CommandManager) _() is saying that there is a CommandManager that
// 		will call this function AND use it's pointer (only 1)

// Add or update commands
func (cm *CommandManager) AddOrUpdate(cmd FlexMacro) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Saftey net
	if cm.macros == nil {
		cm.macros = make(map[string]FlexMacro)
	}

	lookupKey := cmd.ToLookupKey()
	cm.macros[lookupKey] = cmd
	helpers.DebugLog("Registered command '%s' [key: %s]", cmd.Name, lookupKey)
}

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
	helpers.DebugLog("Handling input for key '%s'", lookupKey)
	cm.mu.RLock()
	cmd, exists := cm.macros[lookupKey]
	cm.mu.RUnlock()

	if !exists || !cmd.Enabled {
		return false
	}

	cmd.ExecuteMacro()

	return true
}

func (cm *CommandManager) SyncMacros(cmds []ui.CommandEntry) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Tear down listeners
	//cm.clearActiveHooks()

	// clear out existing macros
	//cm.macros = nil
	// convert each ui entry into macro and re-register
	/*
		for i, entry := range cmds {
			if strings.Trimspace(entry.BindBttn.Text) == "" ||
				strings.TrimSpace(entry.TextEntry.Text) == "" {
				continue
			}
			macro := FlexMacro{
				//Name      string     `json:"name"`
				//Modifiers []string   `json:"modifiers"`
				//Key       string     `json:"key"`
				//Type      ActionType `json:"type"`
				//Payload   string     `json:"payload"`
				Name:      entry.TextEntry.Text,
				Modifiers: entry.BindBttn,
			}
		}
	*/
}
