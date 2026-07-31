package main

import (
	"sync"
)

type CommandManager struct {
	mu     sync.RWMutex
	macros map[string]FlexMacro
}

var GlobalCommandManager = &CommandManager{
	macros: make(map[string]FlexMacro),
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
	debugLog("Registered command '%s' [key: %s]", cmd.Name, lookupKey)
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
	debugLog("Removed macro under lookupkey: %s", lookupKey)
}

// Checks if the command exists and executes it. Returns false if it doesn't exist
func (cm *CommandManager) ExecuteIfMapped(lookupKey string) bool {
	debugLog("Handling input for key '%s'", lookupKey)
	cm.mu.RLock()
	cmd, exists := cm.macros[lookupKey]
	cm.mu.RUnlock()

	if !exists || !cmd.Enabled {
		return false
	}

	cmd.ExecuteMacro()

	return true
}
