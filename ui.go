package main

import (
	//"fmt"
	//"strings"

	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type KeybindButton struct {
	widget.Button
	isListening bool
	modifiers   map[string]bool
	key         string
	onBound     func(modifiers []string, key string)
	window      fyne.Window
}

func setupWindow() {
	trade47 := app.New()
	myWindow := trade47.NewWindow("PoE Trade47")
	myWindow.Resize(fyne.NewSize(720, 720))

	testMacro := FlexMacro{
		Name:    "Go to hideout",
		Type:    ActionTextCommand,
		Payload: "/hideout",
		Enabled: true,
	}
	statusLabel := widget.NewLabel("Bound Keybind: " + testMacro.DisplayTrigger())
	bindBttn := NewKeybindButton(myWindow, "Click to Bind Key", func(mods []string, key string) {
		testMacro.Modifiers = mods
		testMacro.Key = key

		statusLabel.SetText("Bound Keybind" + testMacro.DisplayTrigger())
		fmt.Printf("[TEST SUCCESS] Bound %s -> Mods: %v | Key: %s\n", testMacro.Name, testMacro.Modifiers, testMacro.Key)

	})
	content := container.NewVBox(
		widget.NewLabelWithStyle("Macro Setup", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Action: "+testMacro.Name+" ("+testMacro.Payload+")"),
		bindBttn,
		widget.NewSeparator(),
		statusLabel,
	)

	//content := setupHotkeyUI(myWindow)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func setupHotkeyUI(window fyne.Window) fyne.CanvasObject {
	// Create label headers and boxes
	cmdInput := widget.NewEntry()
	cmdInput.SetPlaceHolder("eg. /logout")

	bindButton := widget.NewButton("Click to Bind Key", nil)
	statusLabel := widget.NewLabel("Satus: Idle")

	//isListeningForBind := false
	bindButton.OnTapped = func() {
		//isListeningForBind = true
		fmt.Println("pressed")
		bindButton.SetText("Press Any Key Combo...")
		statusLabel.SetText("Listening for hotkey sequence...")
	}

	return container.NewVBox(
		widget.NewLabel("In-Game Command: "),
		cmdInput,
		widget.NewLabel("Trigger hotkey: "),
		bindButton,
		widget.NewSeparator(),
		statusLabel,
	)
}

func NewKeybindButton(win fyne.Window, initialLabel string, onBound func(mods []string, key string)) *KeybindButton {
	k := &KeybindButton{
		modifiers: make(map[string]bool),
		onBound:   onBound,
		window:    win,
	}
	k.Text = initialLabel
	k.OnTapped = k.startListening
	k.ExtendBaseWidget(k)
	return k
}

func (k *KeybindButton) FocusGained() {}
func (k *KeybindButton) FocusLost() {
	if k.isListening {
		k.stopListening()
	}
}

func (k *KeybindButton) TypedRune(rune)          {}
func (k *KeybindButton) TypedKey(*fyne.KeyEvent) {}

func (k *KeybindButton) startListening() {
	k.isListening = true
	k.modifiers = make(map[string]bool)
	k.key = ""
	k.SetText("Press key combo...")
	k.window.Canvas().Focus(k)
}

func (k *KeybindButton) stopListening() {
	k.isListening = false
	if k.key == "" {
		k.SetText("Unbound")
	}
}

func (k *KeybindButton) KeyDown(key *fyne.KeyEvent) {
	if !k.isListening {
		return
	}
	keyName := string(key.Name)
	if isModifier(keyName) {
		k.modifiers[cleanModifierName(keyName)] = true
		k.updateActiveLabel()
		return
	}
	k.key = strings.ToLower(keyName)
	k.updateActiveLabel()

	var activeMods []string
	for _, mod := range SupportedModifiers {
		if k.modifiers[string(mod)] {
			activeMods = append(activeMods, string(mod))
		}
	}

	if k.onBound != nil {
		k.onBound(activeMods, k.key)
	}

	k.isListening = false
	k.window.Canvas().Unfocus()
}

func (k *KeybindButton) KeyUp(key *fyne.KeyEvent) {
	if !k.isListening {
		return
	}

	keyName := string(key.Name)
	if isModifier(keyName) {
		delete(k.modifiers, cleanModifierName(keyName))
		k.updateActiveLabel()
	}
}

func (k *KeybindButton) updateActiveLabel() {
	var parts []string
	for _, mod := range SupportedModifiers {
		if k.modifiers[string(mod)] {
			parts = append(parts, strings.ToTitle(string(mod)))
		}
	}

	if k.key != "" {
		parts = append(parts, strings.ToUpper(k.key))
	} else if len(parts) > 0 {
		parts = append(parts, "...")
	} else {
		parts = append(parts, "Press key...")
	}
	k.SetText(strings.Join(parts, " + "))
}

var _ desktop.Keyable = (*KeybindButton)(nil)
var _ fyne.Focusable = (*KeybindButton)(nil)
