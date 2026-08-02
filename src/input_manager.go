package main

import (
	"poetrade47/src/helpers"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

// Global flag for lock-free focus checks
var poeFocused atomic.Bool

func StartFocusTracker(stopChan <-chan struct{}) {
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			helpers.DebugLog("Focus tracker stopped. App or channel closed")
			return
		case <-ticker.C:
			title := robotgo.GetTitle()
			isPoe := strings.Contains(title, "Path of Exile")

			if isPoe != poeFocused.Load() {
				helpers.DebugLog("Focus state changed: PoE active = %v", isPoe)
				poeFocused.Store(isPoe)
				// GlobalScrollClicker.SetModifierPressed(false)
			}
		}
	}
}

func isPoEActive() bool {
	return poeFocused.Load()
}

// Manages single global input listener loop
func StartInputManager(stopChan <-chan struct{}) {
	helpers.DebugLog("Starting global OS input manager...")

	evChan := hook.Start()
	defer func() {
		helpers.DebugLog("Global OS input manager stopped and hook is unbound.")
		hook.End()
	}()

	for {
		select {
		case <-stopChan:
			return
		case ev, ok := <-evChan:
			if !ok {
				return
			}
			if !isPoEActive() {
				continue
			}

			// Event dispatcher
			switch ev.Kind {

			// Handler
			case hook.KeyDown:
				isMod, key := helpers.SolveKeycode(ev.Rawcode)
				helpers.DebugLog("Keydown detected: %s", key)
				if key == "" {
					helpers.DebugLog("Key '%s' was found to be empty?", key)
					key = strings.ToLower(string(ev.Keychar))
				}
				if isMod {
					helpers.DebugLog("MODIFIER keydown: %s", key)
					GlobalScrollClicker.SetModifierPressed(true)
					continue
				}
				go GlobalCommandManager.ExecuteIfMapped(key)
			// To reset mod is released
			case hook.KeyUp:
				isMod, _ := helpers.SolveKeycode(ev.Rawcode)
				if isMod {
					helpers.DebugLog("MODIFIER released")
					GlobalScrollClicker.SetModifierPressed(false)
				}
			case hook.MouseWheel:
				if GlobalScrollClicker.canTrigger() {
					GlobalScrollClicker.TriggerClick()
				}
			}
		}
	}
}

/*

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

*/
