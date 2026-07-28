package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
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

	//SetupSystemTray(trade47, myWindow)
	setupWindow(myWindow)

	myWindow.Show()
}
