package ui

import (
	"fmt"
	"poetrade47/src/helpers"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type SettingsConfig struct {
	ScrollClicker bool
	Commands      []CommandEntry
}

type SettingsViewState struct {
	scrollClickerFeature bool
	entries              []*CommandEntry
	list                 *widget.List
}

func (s *SettingsViewState) BuildSettingsLayout() *fyne.Container {
	page := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Inventory Utilities", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewCheck("Enable Scroll Wheel -> Left click (stash)", func(isChecked bool) {
				s.scrollClickerFeature = isChecked
			}),
		), // top
		container.NewHBox(widget.NewLabelWithStyle(
			"Test Subject",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true})), // bottom
		nil,                 // left
		nil,                 // right
		setupMacroSection(), // center
	)

	return page
}

func NewSettingsView() *SettingsViewState {
	return &SettingsViewState{
		scrollClickerFeature: false, // Default state
		entries:              []*CommandEntry{},
	}
}

func (s *SettingsViewState) GetSettingsConfig() SettingsConfig {
	var validEntries []CommandEntry
	for _, entry := range s.entries {
		if entry.BindBttn.key != "" && entry.TextEntry.Text != "" {
			validEntries = append(validEntries, *entry)
		}
	}

	return SettingsConfig{
		ScrollClicker: s.scrollClickerFeature,
		Commands:      validEntries,
	}
}

func setupMacroSection() fyne.CanvasObject {
	macroListView, macros := buildMacroListUI()

	leftText := widget.NewLabelWithStyle("Macros", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	addBtn := widget.NewButton("Add Cmd", func() {
		macros.Options = append(macros.Options, newCommandEntry())
		macros.List.Refresh()
	})
	header := container.NewBorder(nil, nil, leftText, addBtn, nil)

	return container.NewBorder(header, nil, nil, nil, macroListView)
}

type CommandEntry struct {
	TextEntry  *widget.Entry
	BindBttn   *KeybindButton
	removeBttn *widget.Button
}

func newCommandEntry() *CommandEntry {
	cmd := &CommandEntry{
		BindBttn: &KeybindButton{},
	}

	cmd.TextEntry = widget.NewEntry()
	cmd.TextEntry.SetPlaceHolder("Text (ie. /hideout)")

	cmd.TextEntry.OnSubmitted = func(val string) {
		if strings.TrimSpace(val) == "" && val != "" {
			fyne.Do(func() {
				helpers.DebugLog("Text entry had empty spaces. Set to have nothing")
				cmd.TextEntry.SetText("")
			})
		}
		if val != "" || strings.TrimSpace(val) != "" {
			helpers.DebugLog("TextEntry submitted with text")
		}
	}

	cmd.BindBttn = NewKeybindButton("Click to bind", func(mods []string, key string) {

	})

	cmd.removeBttn = widget.NewButton("Remove", func() {
		helpers.DebugLog("Remove button: tapped")
	})

	// Pass directly to command manager
	//GlobalCommandManager.AddOrUpdate(macro)
	return cmd
}

type MacroState struct {
	Options []*CommandEntry
	List    *widget.List
}

func buildMacroListUI() (fyne.CanvasObject, *MacroState) {
	state := &MacroState{
		Options: []*CommandEntry{},
	}
	state.Options = append(state.Options, newCommandEntry())

	list := widget.NewList(
		func() int {
			return len(state.Options)
		},
		func() fyne.CanvasObject {
			opt := newCommandEntry()

			rightPart := container.NewHBox(
				opt.TextEntry,
				opt.removeBttn,
			)
			return container.NewBorder(nil, nil, opt.BindBttn, nil, rightPart)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			borderContainer := item.(*fyne.Container)

			// Extract our button box and entry from border container
			bindBtn := borderContainer.Objects[1].(*KeybindButton)
			rightPart := borderContainer.Objects[0].(*fyne.Container)

			entry := rightPart.Objects[0].(*widget.Entry)
			remBtn := rightPart.Objects[1].(*widget.Button)

			opt := state.Options[id]

			// Sync bind btn
			bindBtn.OnTapped = func() {
				bindBtn.startListening()
				if canvas := getActiveCanvas(bindBtn); canvas != nil {
					canvas.Focus(bindBtn)
				}
			}
			bindBtn.onBound = func(modifiers []string, key string) {
				bindBtn.updateActiveLabel()
				state.List.Refresh()
			}
			bindBtn.updateActiveLabel()

			// Sync entry
			entry.SetText(opt.TextEntry.Text)
			entry.OnChanged = opt.TextEntry.OnChanged
			entry.OnSubmitted = opt.TextEntry.OnSubmitted

			// Sync rem btn
			remBtn.OnTapped = func() {
				helpers.DebugLog("Attempting to remove command %v", state.Options[id].BindBttn.Text)
				win := fyne.CurrentApp().Driver().AllWindows()[0]

				dialog.ShowConfirm("Delete Hotkey?", "Remove this hotkey and command?", func(confirmed bool) {
					if confirmed {
						if id < 0 || id >= len(state.Options) {
							fmt.Printf("[ERROR] Attempted to remove a non-existant command @ id %v", id)
							return
						}
						// Remove the item id from the backing slice
						state.Options = append(state.Options[:id], state.Options[id+1:]...)
						// refresh the renderer
						state.List.Refresh()
					}

				}, win)
			}
		})
	state.List = list
	return list, state
}
