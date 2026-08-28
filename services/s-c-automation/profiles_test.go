package main

// Zuletzt geaendert: 2026-08-17

import (
	"os"
	"path/filepath"
	"testing"
)

func writeProfiles(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "profiles.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

const goodProfiles = `{
  "profiles": [
    {
      "name": "dev",
      "hotkeys": [
        {"modifiers": ["ctrl", "shift"], "key": "b", "action": "build"},
        {"modifiers": ["ctrl"], "key": "1", "action": "profile:dev"}
      ],
      "watchers": [
        {"path": "/tmp/inbox", "enabled": true}
      ],
      "media": {"volume": 70, "activity_type": "playing", "status_text": "Coding"}
    },
    {
      "name": "gaming",
      "hotkeys": [
        {"modifiers": ["ctrl", "shift"], "key": "m", "action": "mute"}
      ],
      "media": {"volume": 30, "activity_type": "listening", "status_text": "Gaming"}
    },
    {
      "name": "afk"
    }
  ]
}`

func TestLoadProfilesValid(t *testing.T) {
	cfg, err := loadProfiles(writeProfiles(t, goodProfiles))
	if err != nil {
		t.Fatalf("loadProfiles: %v", err)
	}
	if len(cfg.Profiles) != 3 {
		t.Fatalf("profiles=%d, want 3", len(cfg.Profiles))
	}
}

func TestLoadProfilesErrors(t *testing.T) {
	cases := map[string]string{
		"ungültiges JSON":    `{`,
		"kein Profil":        `{"profiles": []}`,
		"Name fehlt":         `{"profiles": [{"hotkeys": []}]}`,
		"Doppeltes Profil":   `{"profiles": [{"name": "dev"}, {"name": "dev"}]}`,
	}
	for name, content := range cases {
		if _, err := loadProfiles(writeProfiles(t, content)); err == nil {
			t.Errorf("%s: Fehler erwartet", name)
		}
	}
}

func TestLoadProfilesFileNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gibt-es-nicht.json")
	_, err := loadProfiles(path)
	if err == nil {
		t.Fatal("Fehler erwartet bei fehlender Datei")
	}
}

func TestProfileManagerSwitch(t *testing.T) {
	cfg, _ := loadProfiles(writeProfiles(t, goodProfiles))
	pm := NewProfileManager(cfg)

	if pm.Current != "dev" {
		t.Fatalf("Current=%q, want dev", pm.Current)
	}
	if !pm.SwitchTo("gaming") {
		t.Fatal("SwitchTo gaming sollte funktionieren")
	}
	if pm.Current != "gaming" {
		t.Fatalf("Current=%q, want gaming", pm.Current)
	}
	if pm.SwitchTo("unbekannt") {
		t.Fatal("SwitchTo unbekannt sollte fehlschlagen")
	}
	if pm.Current != "gaming" {
		t.Fatal("Profil sollte nicht gewechselt sein")
	}
}

func TestProfileManagerMediaSettings(t *testing.T) {
	cfg, _ := loadProfiles(writeProfiles(t, goodProfiles))
	pm := NewProfileManager(cfg)

	ms := pm.GetMediaSettings()
	if ms == nil || ms.Volume != 70 {
		t.Fatalf("dev media volume=%v, want 70", ms)
	}

	pm.SwitchTo("gaming")
	ms = pm.GetMediaSettings()
	if ms == nil || ms.Volume != 30 {
		t.Fatalf("gaming media volume=%v, want 30", ms)
	}

	pm.SwitchTo("afk")
	ms = pm.GetMediaSettings()
	if ms != nil {
		t.Fatalf("afk media sollte nil sein, got %+v", ms)
	}
}

func TestProfileManagerHotkeys(t *testing.T) {
	cfg, _ := loadProfiles(writeProfiles(t, goodProfiles))
	pm := NewProfileManager(cfg)

	hk := pm.GetHotkeys()
	if len(hk) != 2 {
		t.Fatalf("dev hotkeys=%d, want 2", len(hk))
	}
	if hk[0].Action != "build" {
		t.Fatalf("hotkey action=%q, want build", hk[0].Action)
	}

	pm.SwitchTo("afk")
	hk = pm.GetHotkeys()
	if len(hk) != 0 {
		t.Fatalf("afk hotkeys=%d, want 0", len(hk))
	}
}

func TestProfileManagerWatcherEnabled(t *testing.T) {
	cfg, _ := loadProfiles(writeProfiles(t, goodProfiles))
	pm := NewProfileManager(cfg)

	if !pm.IsWatcherEnabled("/tmp/inbox") {
		t.Fatal("/tmp/inbox sollte in dev aktiv sein")
	}
	if !pm.IsWatcherEnabled("/tmp/unknown") {
		t.Fatal("unbekannter Watcher sollte standardmaessig aktiv sein")
	}
}
