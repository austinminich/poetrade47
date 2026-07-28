package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

const (
	ModCtrl  = "ctrl"
	ModShift = "shift"
	ModAlt   = "alt"
)

var (
	modMu         sync.RWMutex
	activeModsMap = make(map[string]bool)
)
var SupportedModifiers = []string{ModCtrl, ModShift, ModAlt}
var isSupportedModMap = map[string]bool{"ctrl": true, "shift": true, "alt": true}

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

func (m FlexMacro) ToLookupKey() string {
	if len(m.Modifiers) == 0 {
		return m.Key // e.g., "F2"
	}

	// Copy and sort to guarantee consistent ordering regardless of how it's stored
	mods := make([]string, len(m.Modifiers))
	copy(mods, m.Modifiers)
	sort.Strings(mods)

	return strings.Join(mods, "+") + "+" + m.Key // e.g., "Ctrl+Shift+F2"
}

func (m *FlexMacro) ExecuteMacro() {
	if !m.Enabled {
		return
	}

	switch m.Type {
	case ActionTextCommand:
		// ie. /hideout
		robotgo.KeyTap("enter")
		time.Sleep(20 * time.Millisecond)
		robotgo.Type(m.Payload)
		time.Sleep(20 * time.Millisecond)
		robotgo.KeyTap("enter")
	case ActionKeySequence:
		// ie. 1,2,3 for flasks
		keys := parseKeys(m.Payload)

		for _, k := range keys {
			robotgo.KeyTap(k)
			if m.DelayMS > 0 {
				robotgo.MilliSleep(m.DelayMS)
			}
		}
	}
	debugLog("Attempting to execute macro: %+v", m)
}

func (m *FlexMacro) DisplayTrigger() string {
	if len(m.Modifiers) == 0 {
		return strings.ToUpper(m.Key)
	}
	var formattedMods []string
	for _, mod := range m.Modifiers {
		formattedMods = append(formattedMods, mod)
	}
	return strings.Join(formattedMods, " + ") + " + " + strings.ToUpper(m.Key)
}

func StartKeyboardEngine(onKeyPress func(string)) {
	evChan := hook.Start()
	defer hook.End()

	for ev := range evChan {
		fmt.Println("Received event: ", ev, " ")
	}
}

// Lower case function name dictates PRIVATE
// Upper case function name dictates PUBLIC and is exported
func parseKeys(s string) []string {
	var keys []string

	// Trim keys to their raw form
	rawKeys := strings.Split(s, ",")
	for _, k := range rawKeys {
		// Trim away spaces
		cleaned := strings.TrimSpace(k)
		if cleaned != "" {
			keys = append(keys, cleaned)
		}
	}

	return keys
}

func cleanModifierName(key string) string {
	k := strings.ToLower(key)
	switch {
	case strings.Contains(k, "shift"):
		return "shift"
	case strings.Contains(k, "control"):
		return "ctrl"
	case strings.Contains(k, "alt"):
		return "alt"
	default:
		return k
	}
}

func isModifier(key string) bool {
	return isSupportedModMap[cleanModifierName(key)]
}

func ExtractModifiers(mask uint16) []string {
	var activeMods []string

	// Standard gohook bit flags: 1 = Shift, 2 = Ctrl, 4 = Alt
	if mask&2 != 0 {
		activeMods = append(activeMods, ModCtrl)
	}
	if mask&1 != 0 {
		activeMods = append(activeMods, ModShift)
	}
	if mask&4 != 0 {
		activeMods = append(activeMods, ModAlt)
	}

	return activeMods
}
