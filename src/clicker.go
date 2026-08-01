package main

import (
	"poetrade47/src/helpers"
	"sync/atomic"

	"github.com/go-vgo/robotgo"
)

type ScrollClicker struct {
	enabled atomic.Bool
}

var GlobalScrollClicker = &ScrollClicker{}

func (sc *ScrollClicker) SetEnabled(enable bool) {
	sc.enabled.Store(enable)
	helpers.DebugLog("ScrollClicker state changed: enabled = %v", enable)
}

func (sc *ScrollClicker) IsEnabled() bool {
	return sc.enabled.Load()
}

func (sc *ScrollClicker) TriggerClick() {
	if sc.IsEnabled() {
		robotgo.Click("left")
	}
}
