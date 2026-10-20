package helpers

/*
input_helpers.go contains runtime functions to deal with hardware input from the user,
	as well as return information that is more cleanly useable in the app.
*/

import (
	"fmt"
	"sort"
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
)

func SolveKeycode(code uint16) (bool, string) {
	if code >= 0xFFBE && code <= 0xFFC9 {
		return false, fmt.Sprintf("f%d", code-0xFFBD)
	}

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
	_, ok := NormalizeModifier(string(key))
	return ok
}

func NormalizeModifier(mod string) (string, bool) {
	mod = strings.ToLower(strings.TrimSpace(mod))

	switch {
	case mod == ModCtrl, strings.Contains(mod, "control"): // Checks for ctrl mod first then if not, checks if the string contains "control" (ie. Control_L, Control_R)
		return ModCtrl, true
	case mod == ModShift || strings.Contains(mod, "shift"):
		return ModShift, true
	case mod == ModAlt || strings.Contains(mod, "alt"):
		return ModAlt, true
	default:
		return "", false
	}
}

func MakeLookupKey(mods []string, key string) string {
	// Normalize the key to lowercase
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" {
		return ""
	}

	normalized := make([]string, 0, len(mods))
	seen := make(map[string]bool, len(mods))
	for _, mod := range mods {
		mod, ok := NormalizeModifier(mod)
		if !ok {
			return ""
		}
		if !seen[mod] {
			seen[mod] = true
			normalized = append(normalized, mod)
		}
	}

	// Sort the normalized modifiers to ensure consistent ordering
	sort.Strings(normalized)
	if len(normalized) == 0 {
		return key
	}
	return strings.Join(normalized, "+") + "+" + key
}

func CleanFyneModifierName(key fyne.KeyName) string {
	mod, ok := NormalizeModifier(string(key))
	if !ok {
		return ""
	}
	return mod
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
