package ui

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type AppShell struct {
	root                 *fyne.Container
	content              *fyne.Container
	status               *widget.Label
	statusMu             sync.Mutex
	poeFocused           bool
	focusStateKnown      bool
	scrollClickerEnabled bool
}

func NewAppShell(init fyne.CanvasObject) *AppShell {
	shell := &AppShell{
		content: container.NewMax(init),
		status:  widget.NewLabel("Ready"),
	}
	shell.root = container.NewBorder(nil, shell.status, nil, nil, shell.content)
	return shell
}

func (s *AppShell) Canvas() fyne.CanvasObject {
	return s.root
}

func (s *AppShell) SetContent(content fyne.CanvasObject) {
	s.content.Objects = []fyne.CanvasObject{content}
	s.content.Refresh()
}

func (s *AppShell) SetStatus(msg string) {
	fyne.Do(func() {
		s.status.SetText(msg)
	})
}

func (s *AppShell) SetPoEFocused(focused bool) {
	s.statusMu.Lock()
	s.poeFocused = focused
	s.focusStateKnown = true
	message := s.statusMessageLocked()
	s.statusMu.Unlock()
	s.SetStatus(message)
}

func (s *AppShell) SetScrollClickerEnabled(enabled bool) {
	s.statusMu.Lock()
	s.scrollClickerEnabled = enabled
	message := s.statusMessageLocked()
	s.statusMu.Unlock()
	s.SetStatus(message)
}

func (s *AppShell) statusMessageLocked() string {
	focusMessage := "Checking game focus"
	if s.focusStateKnown {
		if s.poeFocused {
			focusMessage = "Path of Exile focused"
		} else {
			focusMessage = "Path of Exile not focused; hotkeys paused"
		}
	}

	clickerMessage := "Scroll clicker disabled"
	if s.scrollClickerEnabled {
		clickerMessage = "Scroll clicker enabled"
	}

	return focusMessage + " · " + clickerMessage
}
