package helpers

/*
input_helpers.go contains runtime functions to deal with hardware input from the user,
	as well as return information that is more cleanly useable in the app.
*/

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
)

const (
	ModCtrl  = "ctrl"
	ModShift = "shift"
	ModAlt   = "alt"
)

var (
	SupportedModifiers = []string{ModCtrl, ModShift, ModAlt}
	isSupportedModMap  = map[string]bool{"ctrl": true, "shift": true, "alt": true}
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
		strings.Contains(k, "control"),
		strings.Contains(k, "alt"):
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

func ParseKeys(s string) []string {
	var keys []string

	// Trim keys to their raw form
	rawKeys := strings.Split(s, ",")
	for _, k := range rawKeys {
		// Trim away spaces
		cleaned := strings.TrimSpace(k)
		if cleaned != "" {
			keys = append(keys, cleaned)
		}
	}

	return keys
}
