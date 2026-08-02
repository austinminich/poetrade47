package main

import (
	"fmt"
	"poetrade47/src/helpers"
	"poetrade47/src/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	hook "github.com/robotn/gohook"
)

var trade47 fyne.App

func main() {
	// fmt.Println("Hello world in Go!") // xD
	helpers.DebugLog("App started")

	// Tracker for making sure the program only runs in the desired game (poe1 or 2)
	stopTracker := make(chan struct{})
	stopInput := make(chan struct{})

	go StartFocusTracker(stopTracker)
	go StartInputManager(stopInput)

	defer func() {
		close(stopTracker)
		close(stopInput)
	}()

	trade47 := app.NewWithID("cheesecake47.poetrade")
	myWindow := trade47.NewWindow("PoE Trade47")
	myWindow.SetMaster()
	myWindow.Resize(fyne.NewSize(720, 720))

	// Load resources
	iconRes, err := fyne.LoadResourceFromPath("../resources/cheesecake47.jpg")
	if err != nil || iconRes == nil {
		fmt.Println("[WARN] Failed to load app icon.")
	} else {
		trade47.SetIcon(iconRes)
		myWindow.SetIcon(iconRes)
	}

	// Build settings page
	settingsPage := ui.NewSettingsView(
		GlobalScrollClicker.SetEnabled,
		GlobalCommandManager.SyncMacros,
	)
	myWindow.SetContent(settingsPage.BuildSettingsLayout())

	// Load Settings into memory
	// subscribe inputmanager to ui updates

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
	helpers.DebugLog("Registering System Tray Menu...")

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
