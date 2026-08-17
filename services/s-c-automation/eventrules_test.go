package main

// Zuletzt geaendert: 2026-08-17

import (
	"os"
	"path/filepath"
	"testing"
)

func writeEventRules(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "event_rules.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

func TestLoadEventRulesEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no-file.json")
	cfg, err := loadEventRules(path)
	if err != nil {
		t.Fatalf("keine Datei sollte erlaubt sein: %v", err)
	}
	if len(cfg.Rules) != 0 {
		t.Fatalf("rules=%d, want 0", len(cfg.Rules))
	}
}

func TestLoadEventRulesValid(t *testing.T) {
	content := `{
  "rules": [
    {
      "name": "build-failed-alert",
      "on_event": "event.build.failed",
      "action": {"cmd": "cmd.media.toggle", "target": "S-D", "params": {"toggle": true}}
    },
    {
      "name": "any-build-event",
      "on_event": "event.build.*",
      "if": {"profile": "dev"},
      "action": {"cmd": "cmd.automation.run", "target": "S-C", "params": {"task": "backup"}}
    }
  ]
}`
	cfg, err := loadEventRules(writeEventRules(t, content))
	if err != nil {
		t.Fatalf("loadEventRules: %v", err)
	}
	if len(cfg.Rules) != 2 {
		t.Fatalf("rules=%d, want 2", len(cfg.Rules))
	}
}

func TestLoadEventRulesErrors(t *testing.T) {
	cases := map[string]string{
		"Name fehlt":       `{"rules": [{"on_event": "x", "action": {"cmd": "y"}}]}`,
		"on_event fehlt":   `{"rules": [{"name": "x", "action": {"cmd": "y"}}]}`,
		"Doppelte Regel":   `{"rules": [{"name": "x", "on_event": "a"}, {"name": "x", "on_event": "b"}]}`,
		"ungültiges JSON":  `{`,
	}
	for name, content := range cases {
		if _, err := loadEventRules(writeEventRules(t, content)); err == nil {
			t.Errorf("%s: Fehler erwartet", name)
		}
	}
}

func TestEventRuleMatchExact(t *testing.T) {
	r := EventRule{OnEvent: "event.build.failed"}
	if !r.Match("event.build.failed") {
		t.Fatal("exakter Match sollte funktionieren")
	}
	if r.Match("event.build.succeeded") {
		t.Fatal("anderer Name sollte nicht matchen")
	}
}

func TestEventRuleMatchWildcard(t *testing.T) {
	r := EventRule{OnEvent: "event.build.*"}
	if !r.Match("event.build.failed") {
		t.Fatal("event.build.failed sollte matchen")
	}
	if !r.Match("event.build.succeeded") {
		t.Fatal("event.build.succeeded sollte matchen")
	}
	if r.Match("event.media.state") {
		t.Fatal("event.media.state sollte nicht matchen")
	}
	if r.Match("event.build") {
		t.Fatal("event.build ohne Suffix sollte nicht matchen")
	}
}

func TestEventRuleMatchEmpty(t *testing.T) {
	r := EventRule{OnEvent: ""}
	if r.Match("event.any") {
		t.Fatal("leere Regel sollte nichts matchen")
	}
}
