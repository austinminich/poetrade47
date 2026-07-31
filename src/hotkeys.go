package main

import (
	"fmt"
	"poetrade47/src/helpers"
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
	modMu              sync.RWMutex
	activeModsMap      = make(map[string]bool)
	SupportedModifiers = []string{ModCtrl, ModShift, ModAlt}
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
		keys := helpers.ParseKeys(m.Payload)

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
