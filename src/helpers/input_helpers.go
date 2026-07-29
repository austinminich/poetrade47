package helpers

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
)

func SolveKeycode(code uint16) (bool, string) {
	switch code {
	// Shift, Ctrl, Alt
	case 65505:
		return true, "shift"
	case 65507:
		return true, "ctrl"
	case 65513:
		return true, "alt"
	default:
		return false, fmt.Sprintf("%c", code)
	}
}

func IsFyneModifier(key fyne.KeyName) bool {
	// key should be a string that you can just check against
	k := strings.ToLower(string(key))
	switch {
	case strings.Contains(k, "shift"),
		strings.Contains(k, "shift"),
		strings.Contains(k, "shift"):
		return true
	default:
		return false
	}
}

func CleanFyneModifierName(key fyne.KeyName) string {
	k := strings.ToLower(string(key))
	switch {
	case strings.Contains(k, "shift"):
		return "shift"
	case strings.Contains(k, "control"):
		return "ctrl"
	case strings.Contains(k, "alt"):
		return "alt"
	default:
		return k
	}
}
