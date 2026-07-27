package main

import (
	"strings"

	"fyne.io/fyne/v2"
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

// #region UI Elements
func newCommandOption(win fyne.Window) *fyne.Container {
	testMacro := FlexMacro{
		Name:    "Go to hideout",
		Type:    ActionTextCommand,
		Payload: "/hideout",
		Enabled: true,
	}

	textEntry := widget.NewEntry()
	textEntry.SetPlaceHolder("Text (ie. /hideout)")

	bindButton := NewKeybindButton(win, "Click to Bind Key", func(mods []string, key string) {
		testMacro.Modifiers = mods
		testMacro.Key = key

		//statusLabel.SetText("Bound Keybind" + testMacro.DisplayTrigger())
		//fmt.Printf("[TEST SUCCESS] Bound %s -> Mods: %v | Key: %s\n", testMacro.Name, testMacro.Modifiers, testMacro.Key)

	})

	// Pass directly to command manager
	//GlobalCommandManager.AddOrUpdate(macro)
	return container.NewBorder(nil, nil, bindButton, nil, textEntry)
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

func CreateScrollClickerTile() *fyne.Container {
	check := widget.NewCheck("Enable Scroll Wheel -> Left click (stash)", func(checked bool) {
		GlobalScrollClicker.SetEnabled(checked)
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("Inventory Utilities", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		check,
	)
}

// #endregion

// #region UI Sections

func setupWindow(myWindow fyne.Window) {
	mainLayout := container.NewVBox(
		CreateScrollClickerTile(), // checkboxes
		widget.NewSeparator(),
		setupMacroSection(myWindow),
	)

	//content := setupHotkeyUI(myWindow)
	myWindow.SetContent(mainLayout)
	myWindow.ShowAndRun()
}

func setupMacroSection(win fyne.Window) *fyne.Container {

	macros := container.NewVBox(
		widget.NewLabelWithStyle("Macro Setup", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewAdaptiveGrid(2,
			newCommandOption(win),
			newCommandOption(win),
			newCommandOption(win),
			newCommandOption(win),
			newCommandOption(win),
		),
	)

	return macros
}

// #endregion

// #region Keybind button
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

// #endregion

var _ desktop.Keyable = (*KeybindButton)(nil)
var _ fyne.Focusable = (*KeybindButton)(nil)
