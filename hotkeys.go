package main

import (
	"fmt"
)
import "github.com/go-vgo/robotgo"
import "github.com/robotn/gohook"

type ActiveMacro struct {
	TriggerKey string
	CmdString  string
}

var ActiveProfile = &ActiveMacro{}

func StartKeyboardEngine(onKeyPress func(string)) {
	evChan := hook.Start()
	defer hook.End()

	for ev := range evChan {
		fmt.Println("Received event: ", ev, " ")

	}
}

func ExecuteGameHotkey(command string) {
	robotgo.KeyTap("enter")
	robotgo.Sleep(10)
	robotgo.Type(command)
	robotgo.KeyTap("enter")
}
