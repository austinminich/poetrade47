package main

import (
	_ "embed"
	"poetrade47/src/config"
	"poetrade47/src/helpers"
	"poetrade47/src/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

//go:embed resources/cheesecake47.jpg
var icon []byte
var trade47 fyne.App

func main() {
	// fmt.Println("Hello world in Go!") // xD
	helpers.DebugLog("App started")

	// Tracker for making sure the program only runs in the desired game (poe1 or 2)
	stopTracker := make(chan struct{})
	stopInput := make(chan struct{})

	defer func() {
		close(stopTracker)
		close(stopInput)
	}()

	trade47 = app.NewWithID("cheesecake47.poetrade")
	myWindow := trade47.NewWindow("PoE Trade47")
	myWindow.SetMaster()
	myWindow.Resize(fyne.NewSize(720, 720))

	// Load resources
	iconRes := fyne.NewStaticResource("cheesecake47.jpg", icon)
	trade47.SetIcon(iconRes)
	myWindow.SetIcon(iconRes)

	// Load settings saved from user
	cfg := config.NewManager(trade47.Preferences())
	initCfg := cfg.GetSettings()

	GlobalScrollClicker.SetEnabled(initCfg.ScrollClickerEnabled)
	GlobalCommandManager.SyncMacros(initCfg.SavedMacros)
	GlobalCommandManager.SetLogoutBinding(initCfg.LogoutBinding)
	// Set hotkeys by reading from preferences

	// Build settings page
	settingsPage := ui.NewSettingsView(cfg)

	// Build app shell
	appShell := ui.NewAppShell(settingsPage.Canvas())
	appShell.SetScrollClickerEnabled(initCfg.ScrollClickerEnabled)

	// Link callbacks
	settingsPage.OnClickerFeatureChanged = func(enabled bool) {
		GlobalScrollClicker.SetEnabled(enabled)
		appShell.SetScrollClickerEnabled(enabled)
	}
	settingsPage.OnCommandChanged = GlobalCommandManager.SyncMacros
	settingsPage.OnLogoutBindingChanged = GlobalCommandManager.SetLogoutBinding
	settingsPage.OnStatusChanged = appShell.SetStatus

	// Build app window
	appShell.SetStatus("Checking game focus...")
	myWindow.SetContent(appShell.Canvas())

	go StartFocusTracker(stopTracker, func(isPoeFocused bool) {
		appShell.SetPoEFocused(isPoeFocused)
	})
	go StartInputManager(stopInput)

	myWindow.ShowAndRun()
}
