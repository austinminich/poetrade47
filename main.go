package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

var trade47 fyne.App

func main() {
	// fmt.Println("Hello world in Go!") // xD
	debugLog("App started")
	// Tracker for making sure the program only runs in the desired game (poe1 or 2)
	stopTracker := make(chan struct{})
	stopInput := make(chan struct{})

	go StartFocusTracker(stopTracker)
	go StartInputManager(stopInput)

	defer func() {
		close(stopTracker)
		close(stopInput)
	}()

	trade47 := app.New()
	myWindow := trade47.NewWindow("PoE Trade47")
	myWindow.Resize(fyne.NewSize(720, 720))

	if desktop, ok := trade47.(desktop.App); ok {
		icon := theme.FyneLogo()
		myWindow.SetIcon(icon)

		// Sys tray menu
		trayMenu := fyne.NewMenu("Utility",
			fyne.NewMenuItem("Show Window", func() {
				myWindow.Show()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Close", func() {
				trade47.Quit()
			}),
		)

		desktop.SetSystemTrayMenu(trayMenu)
		desktop.SetSystemTrayIcon(icon)

		// Intecept the window 'X' close button to hide to sys tray
		myWindow.SetCloseIntercept(func() {
			myWindow.Hide()
			debugLog("Window hidden to system tray.")
		})
	}

	setupWindow(myWindow)
}
