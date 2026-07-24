package main

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"github.com/go-vgo/robotgo"
	"github.com/robotn/gohook"
)

const (
	ModCtrl  = "ctrl"
	ModShift = "shift"
	ModAlt   = "alt"
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

func (m *FlexMacro) ExecuteMacro(a fyne.App) {
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
			// TODO: robotgo.KeyTap(k)
			fmt.Printf(string(k))
			if m.DelayMS > 0 {
				robotgo.MilliSleep(m.DelayMS)
			}
		}
	}
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
