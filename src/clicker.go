package main

import (
	"poetrade47/src/helpers"
	"sync/atomic"

	"github.com/go-vgo/robotgo"
)

// Always needs to have a modifier active so we can keep track of it here
type ScrollClicker struct {
	enabled      atomic.Bool
	isModPressed atomic.Bool
}

var GlobalScrollClicker = &ScrollClicker{}

func (sc *ScrollClicker) SetEnabled(enable bool) {
	sc.enabled.Store(enable)
	helpers.DebugLog("ScrollClicker state changed: enabled = %v", enable)
}

func (sc *ScrollClicker) SetModifierPressed(pressed bool) {
	sc.isModPressed.Store(pressed)
	helpers.DebugLog("ScrollClicker modifier pressed changed: pressed = %v", pressed)
}

func (sc *ScrollClicker) canTrigger() bool {
	enabled := sc.enabled.Load()
	mod := sc.isModPressed.Load()
	//helpers.DebugLog("[CHECK] Enabled: %v, ModActive: %v", enabled, mod)
	return enabled && mod
}

func (sc *ScrollClicker) TriggerClick() {
	if sc.canTrigger() {
		//helpers.DebugLog("[CLICKER] Triggering robotgo.Click NOW...")
		robotgo.Click("left")
	}
}
