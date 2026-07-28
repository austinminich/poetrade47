package main

import (
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
			debugLog("Focus tracker stopped. App or channel closed")
			return
		case <-ticker.C:
			title := robotgo.GetTitle()
			isPoe := strings.Contains(title, "Path of Exile")

			if isPoe != poeFocused.Load() {
				debugLog("Focus state changed: PoE active = %v", isPoe)
				poeFocused.Store(isPoe)
			}
		}
	}
}

func isPoEActive() bool {
	return poeFocused.Load()
}

// Manages single global input listener loop
func StartInputManager(stopChan <-chan struct{}) {
	debugLog("Starting global OS input manager...")

	evChan := hook.Start()
	defer func() {
		debugLog("Global OS input manager stopped and hook is unbound.")
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

			// 1. Scroll-clicking handler
			case hook.MouseWheel:
				debugLog("Wheel detected")
				GlobalScrollClicker.TriggerClick()

			// 2. Hotkey handler
			case hook.KeyDown:
				k := strings.ToLower(hook.RawcodetoKeychar(ev.Rawcode))
				if k == "" {
					k = strings.ToLower(string(ev.Keychar))
				}

				if isModifier(k) {
					continue
				}

				m := FlexMacro{
					Modifiers: ExtractModifiers(ev.Mask),
					Key:       k,
				}
				lookupKey := m.ToLookupKey()
				if GlobalCommandManager.ExecuteIfMapped(lookupKey) {
					debugLog("Handled macro keycode: %v", ev.Rawcode)
				}
			}
		}
	}
}
