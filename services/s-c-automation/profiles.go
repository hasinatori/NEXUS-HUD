package main

// Zuletzt geaendert: 2026-08-17
// Profile-System: Definition von Context-Profilen (dev/gaming/afk) mit
// Hotkeys, Watcher- und Media-Einstellungen.

import (
	"encoding/json"
	"fmt"
	"os"
)

// ProfilesConfig enthaelt alle Profile und deren Einstellungen.
type ProfilesConfig struct {
	Profiles []ProfileDef `json:"profiles"`
}

// ProfileDef definiert ein einzelnes Context-Profil.
type ProfileDef struct {
	Name    string           `json:"name"`
	Hotkeys []HotkeyBinding  `json:"hotkeys,omitempty"`
	Watchers []WatcherOverride `json:"watchers,omitempty"`
	Media   *MediaSettings   `json:"media,omitempty"`
}

// HotkeyBinding verknüpft einen Hotkey mit einer Aktion fuer dieses Profil.
type HotkeyBinding struct {
	Modifiers []string `json:"modifiers"`
	Key       string   `json:"key"`
	Action    string   `json:"action"`
}

// WatcherOverride ermoeglicht, Watcher pro Profil zu aktivieren/deaktivieren.
type WatcherOverride struct {
	Path    string `json:"path"`
	Enabled bool   `json:"enabled"`
}

// MediaSettings definiert Media-Voreinstellungen fuer ein Profil.
type MediaSettings struct {
	Volume       int    `json:"volume,omitempty"`
	ActivityType string `json:"activity_type,omitempty"`
	StatusText   string `json:"status_text,omitempty"`
}

// ProfileManager verwaltet den aktuellen Profil-Zustand.
type ProfileManager struct {
	Current  string
	Profiles map[string]ProfileDef
}

// NewProfileManager erzeugt einen neuen ProfileManager aus der Konfiguration.
func NewProfileManager(cfg ProfilesConfig) *ProfileManager {
	pm := &ProfileManager{
		Current:  "dev",
		Profiles: make(map[string]ProfileDef),
	}
	for _, p := range cfg.Profiles {
		pm.Profiles[p.Name] = p
	}
	return pm
}

// SwitchTo wechselt das Profil. Gibt false zurueck wenn Profil unbekannt.
func (pm *ProfileManager) SwitchTo(name string) bool {
	if _, ok := pm.Profiles[name]; !ok {
		return false
	}
	pm.Current = name
	return true
}

// GetMediaSettings gibt die Media-Einstellungen des aktuellen Profils zurueck.
func (pm *ProfileManager) GetMediaSettings() *MediaSettings {
	if p, ok := pm.Profiles[pm.Current]; ok {
		return p.Media
	}
	return nil
}

// GetHotkeys gibt die Hotkey-Bindings des aktuellen Profils zurueck.
func (pm *ProfileManager) GetHotkeys() []HotkeyBinding {
	if p, ok := pm.Profiles[pm.Current]; ok {
		return p.Hotkeys
	}
	return nil
}

// IsWatcherEnabled prueft ob ein Watcher fuer den aktuellen Pfad im Profil aktiv ist.
// Gibt true zurueck wenn kein Override vorhanden ist (Standard: aktiv).
func (pm *ProfileManager) IsWatcherEnabled(path string) bool {
	if p, ok := pm.Profiles[pm.Current]; ok {
		for _, w := range p.Watchers {
			if w.Path == path {
				return w.Enabled
			}
		}
	}
	return true
}

// loadProfiles laedt die Profile-Konfiguration aus einer JSON-Datei.
func loadProfiles(path string) (ProfilesConfig, error) {
	var cfg ProfilesConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("Profiles lesen: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("Profiles parsen: %w", err)
	}
	if len(cfg.Profiles) == 0 {
		return cfg, fmt.Errorf("Profiles: mindestens ein Profil erforderlich")
	}
	seen := map[string]bool{}
	for _, p := range cfg.Profiles {
		if p.Name == "" {
			return cfg, fmt.Errorf("Profiles: Profil ohne Name")
		}
		if seen[p.Name] {
			return cfg, fmt.Errorf("Profiles: doppeltes Profil %q", p.Name)
		}
		seen[p.Name] = true
	}
	return cfg, nil
}
