package ui

/*
settings.go is in charge of the layout and building of the UI for the settings
	page within the app.

Also has callback functions to notify the logic handlers that the ui on this page
	has been updated.
*/

import (
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
	logoutBinding        config.HotkeyBinding
	logoutBindButton     *KeybindButton

	//Callbacks
	OnCommandChanged        func(entries []config.FlexMacroSettings)
	OnClickerFeatureChanged func(enabled bool)
	OnStatusChanged         func(message string)
	OnLogoutBindingChanged  func(binding config.HotkeyBinding)
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
		), //top
		nil,                   // bottom
		nil,                   // left
		nil,                   // right
		s.setupMacroSection(), // center
	)

	return page
}

func NewSettingsView(cfg *config.Manager) *SettingsView {
	savedBinding := cfg.GetSettings().LogoutBinding
	s := &SettingsView{
		cfg:                  cfg,
		scrollClickerFeature: false, // Default state
		entries:              []*ListEntry{},
		logoutBinding:        savedBinding,
	}

	s.logoutBindButton = NewKeybindButton("Click to bind", s.onLogoutBound)
	s.logoutBindButton.key = savedBinding.Key
	for _, modifier := range savedBinding.Modifiers {
		if normalized, ok := helpers.NormalizeModifier(modifier); ok {
			s.logoutBindButton.modifiers[normalized] = true
		}
	}
	if s.logoutBindButton.key != "" {
		s.logoutBindButton.updateActiveLabel()
	}

	s.page = s.buildPage()
	return s
}

func (s *SettingsView) setupMacroSection() fyne.CanvasObject {
	macroList := s.buildMacroList()
	logoutRow := container.NewHBox(
		s.logoutBindButton,
		widget.NewLabel("Logout (built-in)"),
	)

	leftText := widget.NewLabelWithStyle("Macros", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	addBtn := widget.NewButton("Add Cmd", func() {
		s.entries = append(s.entries, newListEntry())
		helpers.DebugLog("Creating new ListEntry...")
		s.list.Refresh()
		s.notifyCommandChanges()
	})
	header := container.NewBorder(nil, nil, leftText, addBtn, nil)

	macroArea := container.NewBorder(
		logoutRow, // fixed-height top row
		nil,
		nil,
		nil,
		macroList, // fills remaining space
	)

	return container.NewBorder(header, nil, nil, nil, macroArea)
}

func (s *SettingsView) buildMacroList() fyne.CanvasObject {
	helpers.DebugLog("Building custom list for macros...")
	savedMacros := s.cfg.GetSettings().SavedMacros

	if len(savedMacros) > 0 {
		helpers.DebugLog("Settings detected saved macros. Initalizing list...")
		for _, entry := range savedMacros {
			s.entries = append(s.entries, newListEntry(entry))
		}
	}

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
			if len(borderContainer.Objects) < 3 {
				return
			}
			bindBtn, ok := borderContainer.Objects[1].(*KeybindButton)
			if !ok {
				return
			}
			entry, ok := borderContainer.Objects[0].(*widget.Entry)
			if !ok {
				return
			}
			remBtn, ok := borderContainer.Objects[2].(*widget.Button)
			if !ok {
				return
			}
			if id < 0 || id >= len(s.entries) {
				return
			}

			opt := s.entries[id]
			if opt == nil || opt.BindBttn == nil || opt.TextEntry == nil {
				return
			}

			if opt.BindBttn.key == "" && len(opt.BindBttn.modifiers) == 0 {
				bindBtn.key = ""
				bindBtn.modifiers = map[string]bool{}
				bindBtn.SetText("Click to bind")
			} else {
				bindBtn.key = opt.BindBttn.key
				bindBtn.modifiers = opt.BindBttn.modifiers
				bindBtn.updateActiveLabel()
			}

			// Sync bind btn
			bindBtn.OnTapped = func() {
				bindBtn.startListening()
				if canvas := getActiveCanvas(bindBtn); canvas != nil {
					canvas.Focus(bindBtn)
				}
			}
			bindBtn.onBound = func(modifiers map[string]bool, key string) {
				candidateModifiers := make([]string, 0, len(modifiers))
				for modifier, pressed := range modifiers {
					if pressed {
						candidateModifiers = append(candidateModifiers, modifier)
					}
				}
				candidateKey := helpers.MakeLookupKey(candidateModifiers, key)

				for _, other := range s.entries {
					if other == nil || other == opt || other.BindBttn == nil || other.BindBttn.key == "" {
						continue
					}

					existingKey := helpers.MakeLookupKey(
						other.BindBttn.GetModifiers(),
						other.BindBttn.key,
					)
					if existingKey != candidateKey {
						continue
					}

					// Restore this row's previous binding; the attempted one is already on bindBtn.
					bindBtn.key = opt.BindBttn.key
					bindBtn.modifiers = make(map[string]bool, len(opt.BindBttn.modifiers))
					for modifier, pressed := range opt.BindBttn.modifiers {
						bindBtn.modifiers[modifier] = pressed
					}
					if bindBtn.key == "" {
						bindBtn.SetText("Click to bind")
					} else {
						bindBtn.updateActiveLabel()
					}

					windows := fyne.CurrentApp().Driver().AllWindows()
					if len(windows) > 0 {
						dialog.ShowInformation(
							"Keybind already assigned",
							"That key combination is already used by another command.",
							windows[0],
						)
						if s.OnStatusChanged != nil {
							s.OnStatusChanged("Keybind " + candidateKey + " already assigned")
						}
					}
					return
				}

				if opt.BindBttn != nil {
					opt.BindBttn.key = key
					opt.BindBttn.modifiers = modifiers
					opt.BindBttn.Text = bindBtn.Text
				}
				bindBtn.key = key
				bindBtn.modifiers = modifiers

				if opt.IsUpdated() {
					s.notifyCommandChanges()
				}
			}

			// Sync entry
			entry.OnChanged = nil
			entry.SetText(opt.TextEntry.Text)
			entry.OnChanged = func(value string) {
				opt.TextEntry.SetText(value)
				s.notifyCommandChanges()
			}

			// Sync rem btn
			remBtn.OnTapped = func() {
				windows := fyne.CurrentApp().Driver().AllWindows()
				if len(windows) == 0 {
					return
				}

				dialog.ShowConfirm("Delete Hotkey?", "Remove this hotkey and command?", func(confirmed bool) {
					if !confirmed {
						return
					}
					for index, current := range s.entries {
						if current == opt {
							s.entries = append(s.entries[:index], s.entries[index+1:]...)
							s.list.Refresh()
							s.notifyCommandChanges()
							return
						}
					}
				}, windows[0])
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
		// Convert list of entries to a list of commands
		cmds := []config.FlexMacroSettings{}

		for _, entry := range s.GetCommandEntries() {
			macro := config.FlexMacroSettings{
				Modifiers: entry.BindBttn.GetModifiers(),
				Key:       entry.BindBttn.key,
				Payload:   entry.TextEntry.Text,
				Enabled:   true, // default on if it's in the UI
			}

			cmds = append(cmds, macro)
		}

		// Update the save state
		s.cfg.Update(func(c *config.AppConfig) {
			c.SavedMacros = cmds
		})

		if s.OnCommandChanged != nil {
			s.OnCommandChanged(cmds)
		}
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

func (s *SettingsView) onLogoutBound(modifiers map[string]bool, key string) {
	modifierNames := make([]string, 0, len(modifiers))
	for modifier, pressed := range modifiers {
		if pressed {
			modifierNames = append(modifierNames, modifier)
		}
	}

	candidateKey := helpers.MakeLookupKey(modifierNames, key)
	for _, entry := range s.entries {
		if entry == nil || entry.BindBttn == nil || entry.BindBttn.key == "" {
			continue
		}
		existingKey := helpers.MakeLookupKey(entry.BindBttn.GetModifiers(), entry.BindBttn.key)
		if candidateKey != "" && candidateKey == existingKey {
			s.logoutBindButton.key = s.logoutBinding.Key
			s.logoutBindButton.modifiers = make(map[string]bool, len(s.logoutBinding.Modifiers))
			for _, modifier := range s.logoutBinding.Modifiers {
				s.logoutBindButton.modifiers[modifier] = true
			}
			s.logoutBindButton.updateActiveLabel()

			if s.OnStatusChanged != nil {
				s.OnStatusChanged("That key combination is already assigned")
			}
			return
		}
	}

	binding := config.HotkeyBinding{
		Modifiers: modifierNames,
		Key:       key,
	}
	s.logoutBinding = binding
	s.cfg.Update(func(cfg *config.AppConfig) {
		cfg.LogoutBinding = binding
	})
	if s.OnLogoutBindingChanged != nil {
		s.OnLogoutBindingChanged(binding)
	}
}
