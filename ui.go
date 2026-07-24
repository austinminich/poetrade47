package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func setupWindow() {
	trade47 := app.New()
	myWindow := trade47.NewWindow("PoE Trade47")
	myWindow.Resize(fyne.NewSize(720, 1080))

	content := setupHotkeyUI(myWindow)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func setupHotkeyUI(window fyne.Window) fyne.CanvasObject {
	// Create label headers and boxes
	cmdInput := widget.NewEntry()
	cmdInput.SetPlaceHolder("eg. /logout")

	bindButton := widget.NewButton("Click to Bind Key", nil)
	statusLabel := widget.NewLabel("Satus: Idle")

	isListeningForBind := false
	bindButton.OnTapped = func() {
		isListeningForBind = true
		bindButton.SetText("Press Any Key Combo...")
		statusLabel.SetText("Listening for hotkey sequence...")
	}

	go StartKeyboardEngine(func(pressedKey string) {
		// UI waiting
		if isListeningForBind {
			fmt.Println("Entering keyboard engine")
			ActiveProfile.TriggerKey = pressedKey
			ActiveProfile.CmdString = strings.TrimSpace(cmdInput.Text)
			isListeningForBind = false

			bindButton.SetText(fmt.Sprintf("Bound to: %s", strings.ToUpper(pressedKey)))
			statusLabel.SetText("Status: Macro cfg locked and loaded.")
			return
		}

		// Verify if keystroke is part of profile
		if ActiveProfile.TriggerKey != "" && pressedKey == ActiveProfile.TriggerKey {
			statusLabel.SetText(fmt.Sprintf("Status: Executing %s", ActiveProfile.CmdString))
			ExecuteGameHotkey(ActiveProfile.CmdString)
		}
	})

	return container.NewVBox(
		widget.NewLabel("In-Game Command: "),
		cmdInput,
		widget.NewLabel("Trigger hotkey: "),
		bindButton,
		widget.NewSeparator(),
		statusLabel,
	)
}
