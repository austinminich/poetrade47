package ui

/*
keybindButton.go is the UI keybind button that contains logic dealing with the UI
	button element wrapped in a custom struct so that i can tap into the button's
	attributes for ui changes
*/

import (
	"poetrade47/src/helpers"
	"strings"

	"fyne.io/fyne/v2"
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

func NewKeybindButton(initialLabel string, onBound func(mods []string, key string)) *KeybindButton {
	k := &KeybindButton{
		modifiers: make(map[string]bool),
		onBound:   onBound,
	}
	k.Text = initialLabel
	k.OnTapped = k.startListening
	windows := fyne.CurrentApp().Driver().AllWindows()
	if len(windows) > 0 {
		k.window = windows[0]
	}
	k.ExtendBaseWidget(k)
	return k
}

func formatKeyLabel(mods []string, key string) string {
	if len(mods) == 0 && key == "" {
		return "Click to Bind Key"
	}
	return strings.ToUpper(strings.Join(mods, "+") + key)
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

	canvas := getActiveCanvas(k)
	if canvas != nil {
		fyne.Do(func() {
			k.SetText("Press key combo...")
			k.window.Canvas().Focus(k)
		})
	}
}

func (k *KeybindButton) stopListening() {
	k.isListening = false
	if k.key == "" {
		fyne.Do(func() {
			k.SetText("Unbound")
		})
	}
}

func (k *KeybindButton) KeyUp(key *fyne.KeyEvent) {
	if !k.isListening {
		return
	}
	if helpers.IsFyneModifier(key.Name) {
		delete(k.modifiers, helpers.CleanFyneModifierName(key.Name))
		k.updateActiveLabel()
	}
}

func (k *KeybindButton) KeyDown(key *fyne.KeyEvent) {
	if !k.isListening {
		return
	}
	if helpers.IsFyneModifier(key.Name) {
		k.modifiers[helpers.CleanFyneModifierName(key.Name)] = true
		k.updateActiveLabel()
		return
	}
	k.key = strings.ToLower(string(key.Name))
	k.updateActiveLabel()

	var activeMods []string
	for _, mod := range helpers.SupportedModifiers {
		if k.modifiers[string(mod)] {
			activeMods = append(activeMods, string(mod))
		}
	}

	if k.onBound != nil {
		k.onBound(activeMods, k.key)
	}

	k.isListening = false
	canvas := getActiveCanvas(k)
	if canvas != nil {
		k.window.Canvas().Unfocus()
	}
}

func (k *KeybindButton) updateActiveLabel() {
	var parts []string
	for _, mod := range helpers.SupportedModifiers {
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
	fyne.Do(func() {
		k.SetText(strings.Join(parts, " + "))
	})

}

// #endregion

func getActiveCanvas(obj fyne.CanvasObject) fyne.Canvas {
	// 1. Try resolving via object hierarchy
	if canvas := fyne.CurrentApp().Driver().CanvasForObject(obj); canvas != nil {
		return canvas
	}

	// 2. Fallback: Get the main application window's canvas directly
	windows := fyne.CurrentApp().Driver().AllWindows()
	if len(windows) > 0 {
		return windows[0].Canvas()
	}

	return nil
}
