package main

import (
	"poetrade47/src/helpers"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

type pendingMacro struct {
	macro     FlexMacro
	modifiers []string
}

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
	downKeys := make(map[uint16]bool)
	pending := make([]pendingMacro, 0)

	currentModifiers := func() []string {
		modifiers := make([]string, 0, len(activeMods))
		for modifier := range activeMods {
			modifiers = append(modifiers, modifier)
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

			poeActive := isPoEActive()

			if !poeActive {
				// Clear all state when PoE is not focused
				clear(activeMods)
				clear(downKeys)
				pending = pending[:0]
				GlobalScrollClicker.SetModifierPressed(false)
				continue
			}

			// Event dispatcher
			switch ev.Kind {

			// Handler
			case hook.KeyDown:
				isMod, key := helpers.SolveKeycode(ev.Rawcode)
				if isMod {
					if modifier, ok := helpers.NormalizeModifier(key); ok {
						activeMods[modifier] = true
						helpers.DebugLog("MODIFIER pressed: %s", modifier)
						GlobalScrollClicker.SetModifierPressed(activeMods[helpers.ModCtrl])
					}
					continue
				}

				// Ignore key-repeat events while this physical key remains held.
				if downKeys[ev.Rawcode] {
					continue
				}
				downKeys[ev.Rawcode] = true

				if key == "" {
					key = strings.ToLower(string(ev.Keychar))
				}

				modifiers := currentModifiers()
				lookupKey := helpers.MakeLookupKey(modifiers, key)
				macro, found := GlobalCommandManager.Lookup(lookupKey)
				if !found {
					continue
				}

				if len(modifiers) == 0 {
					go macro.ExecuteMacro()
					continue
				}

				pending = append(pending, pendingMacro{
					macro:     macro,
					modifiers: append([]string(nil), modifiers...),
				})

			case hook.KeyUp:
				delete(downKeys, ev.Rawcode)

				isMod, key := helpers.SolveKeycode(ev.Rawcode)
				if !isMod {
					continue
				}

				modifier, ok := helpers.NormalizeModifier(key)
				if !ok {
					continue
				}
				delete(activeMods, modifier)
				GlobalScrollClicker.SetModifierPressed(activeMods[helpers.ModCtrl])

				remaining := pending[:0]
				for _, item := range pending {
					stillHeld := false
					for _, required := range item.modifiers {
						if activeMods[required] {
							stillHeld = true
							break
						}
					}

					if stillHeld {
						remaining = append(remaining, item)
					} else {
						macro := item.macro
						go macro.ExecuteMacro()
					}
				}
				pending = remaining

			case hook.MouseWheel:
				if GlobalScrollClicker.canTrigger() {
					GlobalScrollClicker.TriggerClick()
				}
			}
		}
	}
}
