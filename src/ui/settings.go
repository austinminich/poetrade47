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
	entries              []*ListEntry
	list                 *widget.List
	page                 fyne.CanvasObject

	//Callbacks
	OnCommandChanged        func(entries []ListEntry)
	OnClickerFeatureChanged func(enabled bool)
}

func (s *SettingsView) Canvas() fyne.CanvasObject {
	return s.page
}

func (s *SettingsView) buildPage() fyne.CanvasObject {
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
	s := &SettingsView{
		cfg:                  cfg,
		scrollClickerFeature: false, // Default state
		entries:              []*ListEntry{},
	}

	s.page = s.buildPage()
	return s
}

func (s *SettingsView) setupMacroSection() fyne.CanvasObject {
	macroList := s.buildMacroList()

	leftText := widget.NewLabelWithStyle("Macros", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	addBtn := widget.NewButton("Add Cmd", func() {
		s.entries = append(s.entries, newListEntry())
		helpers.DebugLog("Creating new ListEntry...")
		s.list.Refresh()
		s.notifyCommandChanges()
	})
	header := container.NewBorder(nil, nil, leftText, addBtn, nil)

	return container.NewBorder(header, nil, nil, nil, macroList)
}

func (s *SettingsView) buildMacroList() fyne.CanvasObject {
	helpers.DebugLog("Building custom list for macros...")

	if len(s.entries) == 0 {
		s.entries = append(s.entries, newListEntry())
	}

	list := widget.NewList(
		func() int {
			return len(s.entries)
		},
		func() fyne.CanvasObject {
			opt := newListEntry()
			return opt.DrawBox
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			borderContainer := item.(*fyne.Container)

			// Extract our button box and entry from border container
			bindBtn := borderContainer.Objects[1].(*KeybindButton)
			entry := borderContainer.Objects[0].(*widget.Entry)
			remBtn := borderContainer.Objects[2].(*widget.Button)

			opt := s.entries[id]

			if opt.BindBttn != nil {
				bindBtn.key = opt.BindBttn.key
				bindBtn.modifiers = opt.BindBttn.modifiers
				bindBtn.SetText(opt.BindBttn.Text)
			}

			// Sync bind btn
			bindBtn.OnTapped = func() {
				bindBtn.startListening()
				if canvas := getActiveCanvas(bindBtn); canvas != nil {
					canvas.Focus(bindBtn)
				}
			}
			bindBtn.onBound = func(modifiers map[string]bool, key string) {
				if opt.BindBttn != nil {
					opt.BindBttn.key = key
					opt.BindBttn.modifiers = modifiers
					opt.BindBttn.Text = "dummy test"
				}
				bindBtn.key = key
				bindBtn.modifiers = modifiers
				bindBtn.updateActiveLabel()

				if opt.IsUpdated() {
					s.notifyCommandChanges()
				}
			}
			//bindBtn.updateActiveLabel()

			// Sync entry
			entry.OnChanged = nil
			if opt.TextEntry != nil {
				entry.SetText(opt.TextEntry.Text)
			} else {
				entry.SetText("")
			}
			entry.OnSubmitted = func(val string) {
				opt.TextEntry.Text = entry.Text
				if opt.TextEntry != nil {
					opt.TextEntry.Text = val
					if opt.IsUpdated() {
						s.notifyCommandChanges()
					}
				}
			}

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

func (s *SettingsView) GetCommandEntries() []ListEntry {
	helpers.DebugLog("Collecting valid command entries...")
	validEntries := make([]ListEntry, 0, len(s.entries))

	for _, entry := range s.entries {
		if entry == nil {
			continue
		}

		cleanKey := strings.TrimSpace(entry.BindBttn.key)
		cleanText := strings.TrimSpace(entry.TextEntry.Text)

		if cleanKey != "" && cleanText != "" {
			validEntries = append(validEntries, *entry)
		}
	}

	return validEntries
}
