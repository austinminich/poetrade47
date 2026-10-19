package ui

import (
	"poetrade47/src/config"
	"poetrade47/src/helpers"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ListEntry struct {
	DrawBox    *fyne.Container
	TextEntry  *widget.Entry
	BindBttn   *KeybindButton
	removeBttn *widget.Button

	isListening bool
}

func newListEntry(opts ...config.FlexMacroSettings) *ListEntry {
	entry := &ListEntry{
		BindBttn: &KeybindButton{},
	}

	entry.TextEntry = widget.NewEntry()
	entry.TextEntry.SetPlaceHolder("Text (ie. /hideout)")

	entry.TextEntry.OnSubmitted = func(val string) {
		if strings.TrimSpace(val) == "" && val != "" {
			fyne.Do(func() {
				helpers.DebugLog("Text entry had empty spaces. Set to have nothing")
				entry.TextEntry.SetText("")
			})
		}
		if val != "" || strings.TrimSpace(val) != "" {
			helpers.DebugLog("TextEntry submitted with text: %v", val)
		}
	}

	entry.BindBttn = NewKeybindButton("Click to bind", func(mods map[string]bool, key string) {
		// Update UI element of command entry if you want to change any visuals
	})

	entry.removeBttn = widget.NewButton("Remove", func() {
		// Removing it, so no need to update the ui for this entry.
	})

	if len(opts) > 0 {
		cmd := opts[0]

		if entry.IsEmpty() {
			entry.TextEntry.SetText(cmd.Payload)
			entry.BindBttn.key = cmd.Key
			entry.BindBttn.modifiers = make(map[string]bool)
			for _, mod := range cmd.Modifiers {
				if mod != "" {
					entry.BindBttn.modifiers[mod] = true
				}
			}
			entry.BindBttn.updateActiveLabel()
		}
	}

	entry.DrawBox = container.NewBorder(nil, nil, entry.BindBttn, entry.removeBttn, entry.TextEntry)
	return entry
}

func (e *ListEntry) IsUpdated() bool {
	// Check if textbox is not nil and keybindbttn has something
	if e.TextEntry.Text != "" && e.BindBttn.key != "" {
		helpers.DebugLog("Custom ListEntry updated and contains something useful...")
		return true
	}
	return false
}

func (e *ListEntry) IsEmpty() bool {
	if e == nil || e.TextEntry == nil || e.BindBttn == nil {
		return false
	}
	return true
}
