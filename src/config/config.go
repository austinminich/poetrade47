package config

/*
Config.go is responsible for the logic revolving around the serialization of
	settings from the app the computer and vice-versa
*/

import (
	"encoding/json"
	"sync"

	"fyne.io/fyne/v2"
)

type FlexMacroSettings struct {
	Name      string   `json:"name"`
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
	cfg  AppConfig
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

	m.cfg = defaultConfig()
	rawJSON := m.pref.StringWithFallback("app_config", "")
	if rawJSON != "" {
		_ = json.Unmarshal([]byte(rawJSON), &m.cfg)
	}
}

func (m *Manager) GetSettings() AppConfig {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.cfg
}

func (m *Manager) Update(fn func(cfg *AppConfig)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fn(&m.cfg)

	data, err := json.Marshal(m.cfg)
	if err == nil {
		m.pref.SetString("app_config", string(data))
	}
}
