package ui

/*
settings.go is in charge of the layout and building of the UI for the settings
	page within the app.

Also has callback functions to notify the logic handlers that the ui on this page
	has been updated.
*/

import (
	"fmt"
	"poetrade47/src/config"
	"poetrade47/src/helpers"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type SettingsView struct {
	cfg                  *config.Manager
	scrollClickerFeature bool
	entries              []*CommandEntry
	list                 *widget.List

	//Callbacks
	OnCommandChanged        func(entries []CommandEntry)
	OnClickerFeatureChanged func(enabled bool)
}

func (s *SettingsView) BuildSettingsLayout() *fyne.Container {
	currentCfg := s.cfg.GetSettings()
	clicker_checkBox := widget.NewCheck("Enable Scroll Wheel -> Left click (stash)", func(isChecked bool) {
		s.cfg.Update(func(cfg *config.AppConfig) {
			cfg.ScrollClickerEnabled = isChecked
		})
		s.scrollClickerFeature = isChecked
		s.notifyScrollClickerChanged(isChecked)
	})

	clicker_checkBox.SetChecked(currentCfg.ScrollClickerEnabled)

	page := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Inventory Utilities", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			clicker_checkBox,
		), // top
		container.NewHBox(widget.NewLabelWithStyle(
			"Test Subject",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true})), // bottom
		nil,                   // left
		nil,                   // right
		s.setupMacroSection(), // center
	)

	return page
}

func NewSettingsView(cfg *config.Manager) *SettingsView {
	return &SettingsView{
		cfg:                  cfg,
		scrollClickerFeature: false, // Default state
		entries:              []*CommandEntry{},
	}
}

func (s *SettingsView) setupMacroSection() fyne.CanvasObject {
	macroList := s.buildMacroList()

	leftText := widget.NewLabelWithStyle("Macros", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	addBtn := widget.NewButton("Add Cmd", func() {
		s.entries = append(s.entries, newCommandEntry())
		s.list.Refresh()
	})
	header := container.NewBorder(nil, nil, leftText, addBtn, nil)

	return container.NewBorder(header, nil, nil, nil, macroList)
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
	helpers.DebugLog("Creating new CommandEntry...")

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
	// GlobalCommandManager.AddOrUpdate(macro)
	return cmd
}

func (s *SettingsView) buildMacroList() fyne.CanvasObject {
	helpers.DebugLog("Building custom list for macros...")
	s.entries = append(s.entries, newCommandEntry())

	list := widget.NewList(
		func() int {
			return len(s.entries)
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

			opt := s.entries[id]

			// Sync bind btn
			bindBtn.OnTapped = func() {
				bindBtn.startListening()
				if canvas := getActiveCanvas(bindBtn); canvas != nil {
					canvas.Focus(bindBtn)
				}
			}
			bindBtn.onBound = func(modifiers []string, key string) {
				bindBtn.updateActiveLabel()
				s.list.Refresh()
			}
			bindBtn.updateActiveLabel()

			// Sync entry
			entry.SetText(opt.TextEntry.Text)
			entry.OnChanged = opt.TextEntry.OnChanged
			entry.OnSubmitted = opt.TextEntry.OnSubmitted

			// Sync rem btn
			remBtn.OnTapped = func() {
				helpers.DebugLog("Attempting to remove command %v", s.entries[id].BindBttn.Text)
				win := fyne.CurrentApp().Driver().AllWindows()[0]

				dialog.ShowConfirm("Delete Hotkey?", "Remove this hotkey and command?", func(confirmed bool) {
					if confirmed {
						if id < 0 || id >= len(s.entries) {
							fmt.Printf("[ERROR] Attempted to remove a non-existant command @ id %v", id)
							return
						}
						// Remove the item id from the backing slice
						s.entries = append(s.entries[:id], s.entries[id+1:]...)
						// refresh the renderer
						s.list.Refresh()

						// Fire Update Events
						s.notifyCommandChanges()
					} else {
						helpers.DebugLog("User hit cancel on removing command %v", s.entries[id].TextEntry.Text)
					}
				}, win)
			}
		})
	s.list = list
	return list
}

func (s *SettingsView) notifyScrollClickerChanged(isChecked bool) {
	helpers.DebugLog("Notifying that scroll clicker has changed state...")
	if s.OnClickerFeatureChanged != nil {
		s.OnClickerFeatureChanged(isChecked)
	}
}

func (s *SettingsView) notifyCommandChanges() {
	helpers.DebugLog("Notifying that commands list has changed state...")
	if s.OnCommandChanged != nil {
		s.OnCommandChanged(s.GetCommandEntries())
	}
}

func (s *SettingsView) GetCommandEntries() []CommandEntry {
	helpers.DebugLog("Collecting valid command entries...")
	validEntries := make([]CommandEntry, 0, len(s.entries))

	for _, entry := range s.entries {
		if entry == nil {
			continue
		}

		cleanKey := strings.TrimSpace(entry.BindBttn.key)
		cleanText := strings.TrimSpace(entry.TextEntry.Text)

		if cleanKey != "" && cleanText != "" {
			validEntries = append(validEntries, CommandEntry{})
		}
	}

	return validEntries
}
