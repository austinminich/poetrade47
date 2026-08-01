package main

import (
	"poetrade47/src/helpers"
	"poetrade47/src/ui"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

// Global flag for lock-free focus checks
var poeFocused atomic.Bool

func ApplySettingsConfig(cfg ui.SettingsConfig) {
	GlobalScrollClicker.SetEnabled(cfg.ScrollClicker)
	GlobalCommandManager.SyncMacros(cfg.Commands)
}

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

			// 1. Scroll-clicking handler
			case hook.MouseWheel:
				GlobalScrollClicker.TriggerClick()

			// 2. Hotkey handler
			case hook.KeyDown:
				isMod, key := helpers.SolveKeycode(ev.Rawcode)
				helpers.DebugLog("Keydown detected: %s", key)
				if key == "" {
					helpers.DebugLog("Key '%s' was found to be empty?", key)
					key = strings.ToLower(string(ev.Keychar))
				}
				if isMod {
					helpers.DebugLog("MODIFIER keydown: %s", key)
					continue
				}
				go GlobalCommandManager.ExecuteIfMapped(key)
			}
		}
	}
}
