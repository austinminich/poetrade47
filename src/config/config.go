package config

/*
Config.go is responsible for the logic revolving around the serialization of
	settings from the app the computer and vice-versa
*/

import (
	"encoding/json"
	"poetrade47/src/helpers"
	"sync"

	"fyne.io/fyne/v2"
)

type FlexMacroSettings struct {
	//Name      string   `json:"name"`
	Modifiers []string `json:"modifiers"`
	Key       string   `json:"key"`
	Payload   string   `json:"payload"`
	Enabled   bool     `json:"enabled"`
}

type AppConfig struct {
	ScrollClickerEnabled bool
	SelectedModifier     string
	OverlayOpacity       float64
	SavedMacros          []FlexMacroSettings
}

func defaultConfig() AppConfig {
	return AppConfig{
		ScrollClickerEnabled: false,
		SelectedModifier:     "",
		OverlayOpacity:       1.0,
		SavedMacros:          []FlexMacroSettings{},
	}
}

type Manager struct {
	mu   sync.RWMutex
	Cfg  AppConfig
	pref fyne.Preferences
}

func NewManager(pref fyne.Preferences) *Manager {
	m := &Manager{pref: pref}
	m.Load()
	return m
}

func (m *Manager) Load() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Cfg = defaultConfig()
	rawJSON := m.pref.StringWithFallback("app_config", "")
	if rawJSON != "" {
		helpers.DebugLog("Found saved settings. Loading preferences...")
		_ = json.Unmarshal([]byte(rawJSON), &m.Cfg)
	}
}

func (m *Manager) GetSettings() AppConfig {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.Cfg
}

func (m *Manager) Update(fn func(cfg *AppConfig)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fn(&m.Cfg)

	data, err := json.Marshal(m.Cfg)
	if err == nil {
		m.pref.SetString("app_config", string(data))
	} else {
		helpers.DebugLog("Failed to update preferences: ", err)
	}
}
