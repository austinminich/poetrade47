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

	activeMods := make(map[string]bool)
	currentModifiers := func() []string {
		modifiers := make([]string, 0, len(activeMods))
		for mod, pressed := range activeMods {
			if pressed {
				modifiers = append(modifiers, mod)
			}
		}
		return modifiers
	}

	for {
		select {
		case <-stopChan:
			return
		case ev, ok := <-evChan:
			if !ok {
				return
			}

			poeInFocus := isPoEActive()
			if !poeInFocus && ev.Kind != hook.KeyUp {
				continue
			}
			if poeInFocus {
				helpers.DebugLog("PoE in focus.")
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
					if modifier, ok := helpers.NormalizeModifier(key); ok {
						activeMods[modifier] = true
						helpers.DebugLog("MODIFIER pressed: %s", modifier)
						GlobalScrollClicker.SetModifierPressed(activeMods[helpers.ModCtrl])
					}
					continue
				}

				lookupKey := helpers.MakeLookupKey(currentModifiers(), key)
				if lookupKey != "" {
					go GlobalCommandManager.ExecuteIfMapped(lookupKey)
				}
			// To reset mod is released
			case hook.KeyUp:
				isMod, key := helpers.SolveKeycode(ev.Rawcode)
				if isMod {
					if modifier, ok := helpers.NormalizeModifier(key); ok {
						helpers.DebugLog("Modifier released: %s", modifier)
						delete(activeMods, modifier)
						GlobalScrollClicker.SetModifierPressed(activeMods[helpers.ModCtrl])
					}
				}
			case hook.MouseWheel:
				if GlobalScrollClicker.canTrigger() {
					GlobalScrollClicker.TriggerClick()
				}
			}
		}
	}
}
