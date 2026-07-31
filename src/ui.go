package main

import (
	"fmt"
	"poetrade47/src/helpers"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	hook "github.com/robotn/gohook"
)

func setupWindow(myWindow fyne.Window) {
	footer := container.NewHBox(widget.NewLabelWithStyle(
		"Test Subject",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true}))

	mainLayout := container.NewBorder(
		CreateScrollClickerTile(), // top
		footer,                    // bottom
		nil,                       // left
		nil,                       // right
		setupMacroSection(),       // center
	)

	myWindow.SetContent(mainLayout)
	myWindow.ShowAndRun()
}

func SetupSystemTray(app fyne.App, win fyne.Window) { // This function causes the
	// Intecept the window 'X' close button to hide to sys tray
	/*
		win.SetCloseIntercept(func() {
			debugLog("Window hidden to system tray.")
			//win.Hide() // this breaks the functionality for wheel input.
		})
	*/

	desktop, ok := app.(desktop.App)
	if !ok {
		fmt.Println("[ERROR] Desktop driver not supported or type assertion failed!")
		return
	}
	debugLog("Registering System Tray Menu...")

	// Sys tray menu
	trayMenu := fyne.NewMenu("Trade47",
		fyne.NewMenuItem("Show Window", func() {
			win.Show()
			win.RequestFocus()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Close", func() {
			//stopHooks()
			trade47.Quit()
		}),
	)

	desktop.SetSystemTrayWindow(win)
	desktop.SetSystemTrayMenu(trayMenu)
}

func stopHooks() {
	defer func() {
		_ = recover()
	}()
	hook.End()
}

// #region UI sub Elements

type CommandOption struct {
	Macro      *FlexMacro
	TextEntry  *widget.Entry
	BindBttn   *KeybindButton
	removeBttn *widget.Button
}

func newCommandOption() *CommandOption {
	cmd := &CommandOption{
		Macro:    &FlexMacro{},
		BindBttn: &KeybindButton{},
	}

	cmd.TextEntry = widget.NewEntry()
	cmd.TextEntry.SetPlaceHolder("Text (ie. /hideout)")

	cmd.TextEntry.OnSubmitted = func(val string) {
		if strings.TrimSpace(val) == "" && val != "" {
			fyne.Do(func() {
				debugLog("Text entry had empty spaces. Set to have nothing")
				cmd.TextEntry.SetText("")
			})
		}
		if val != "" || strings.TrimSpace(val) != "" {
			debugLog("TextEntry submitted with text")
		}
	}

	cmd.BindBttn = NewKeybindButton("Click to bind", func(mods []string, key string) {
		cmd.Macro.Modifiers = mods
		cmd.Macro.Key = key
	})

	cmd.removeBttn = widget.NewButton("Remove", func() {
		debugLog("Remove button: tapped")
	})

	// Pass directly to command manager
	//GlobalCommandManager.AddOrUpdate(macro)
	return cmd
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

// #region Macro Section

type MacroState struct {
	Options []*CommandOption
	List    *widget.List
}

func setupMacroSection() fyne.CanvasObject {
	macroListView, macros := buildMacroListUI()

	leftText := widget.NewLabelWithStyle("Macros", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	addBtn := widget.NewButton("Add Cmd", func() {
		macros.Options = append(macros.Options, newCommandOption())
		macros.List.Refresh()
	})
	header := container.NewBorder(nil, nil, leftText, addBtn, nil)

	return container.NewBorder(header, nil, nil, nil, macroListView)
}

func buildMacroListUI() (fyne.CanvasObject, *MacroState) {
	state := &MacroState{
		Options: []*CommandOption{},
	}
	state.Options = append(state.Options, newCommandOption())

	list := widget.NewList(
		func() int {
			return len(state.Options)
		},
		func() fyne.CanvasObject {
			opt := newCommandOption()

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
				opt.Macro.Modifiers = modifiers
				opt.Macro.Key = key
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
				debugLog("Attempting to remove command %v", state.Options[id].Macro.Name)
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

// #endregion

// #region Keybind button

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
	for _, mod := range SupportedModifiers {
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
