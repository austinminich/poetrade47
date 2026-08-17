package ui

/*
keybindButton.go is the UI keybind button that contains logic dealing with the UI
	button element wrapped in a custom struct so that i can tap into the button's
	attributes for ui changes
*/

import (
	"maps"
	"poetrade47/src/helpers"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type KeybindButton struct {
	widget.Button
	isListening bool
	modifiers   map[string]bool
	key         string
	onBound     func(modifiers map[string]bool, key string)
	window      fyne.Window
}

func NewKeybindButton(initialLabel string, onBound func(mods map[string]bool, key string)) *KeybindButton {
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

func (b *KeybindButton) GetKey() string {
	return b.key
}

func (b *KeybindButton) GetModifiers() []string {
	ret := slices.Collect(maps.Keys(b.modifiers))
	return ret
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
	helpers.DebugLog("Binding button...")
	k.isListening = true
	k.modifiers = make(map[string]bool)
	k.key = ""

	canvas := getActiveCanvas(k)
	if canvas != nil {
		fyne.Do(func() {
			k.SetText("Press key combo...")
			k.window.Canvas().Focus(k)
		})
	} else {
		helpers.DebugLog("[ERROR] Failed to find a canvas!")
	}
}

func (k *KeybindButton) stopListening() {
	k.isListening = false
	if k.key == "" {
		fyne.Do(func() {
			k.SetText("Unbound")
		})
		helpers.DebugLog("Finished bind. Bind is empty.")
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

	var activeMods = make(map[string]bool)
	for _, mod := range helpers.SupportedModifiers {
		if k.modifiers[string(mod)] {
			activeMods[string(mod)] = true
		}
	}

	if k.onBound != nil {
		helpers.DebugLog("KeybindButton finished bind. Calling onBound func.")
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
