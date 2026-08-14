package main

// Zuletzt geändert: 2026-08-14

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "automations.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

const goodConfig = `{
  "tasks": {
    "backup": {"command": ["sh", "-c", "cp -r src bak"], "timeout_ms": 30000},
    "cleanup": {"command": ["rm", "-rf", "tmp"]}
  },
  "watchers": [
    {"path": "/tmp/inbox", "triggers": ["created"], "then": "backup"},
    {"path": "/tmp/trash", "triggers": [], "then": "cleanup"}
  ]
}`

func TestLoadConfigValid(t *testing.T) {
	cfg, err := loadConfig(writeConfig(t, goodConfig))
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if len(cfg.Tasks) != 2 || len(cfg.Watchers) != 2 {
		t.Fatalf("tasks=%d watchers=%d, want 2/2", len(cfg.Tasks), len(cfg.Watchers))
	}
	if cfg.Tasks["backup"].TimeoutMS != 30000 {
		t.Fatalf("timeout=%d", cfg.Tasks["backup"].TimeoutMS)
	}
}

func TestLoadConfigErrors(t *testing.T) {
	cases := map[string]string{
		"ungültiges JSON": `{`,
		"then fehlt":      `{"tasks": {"a": {"command": ["true"]}}, "watchers": [{"path": "/tmp/x"}]}`,
		"then unbekannt":  `{"tasks": {"a": {"command": ["true"]}}, "watchers": [{"path": "/tmp/x", "then": "b"}]}`,
		"path fehlt":      `{"tasks": {"a": {"command": ["true"]}}, "watchers": [{"then": "a"}]}`,
		"Datei fehlt":     `{"tasks": {"a": {"command": ["true"]}}}`,
	}
	for name, content := range cases {
		path := writeConfig(t, content)
		if name == "Datei fehlt" {
			path = filepath.Join(t.TempDir(), "gibt-es-nicht.json")
		}
		if _, err := loadConfig(path); err == nil {
			t.Errorf("%s: Fehler erwartet", name)
		}
	}
}

func TestNormalizeChange(t *testing.T) {
	got := watcherTrigger([]string{"created", "changed", "created", "kaputt"})
	want := []string{"write"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("watcherTrigger=%v, want %v", got, want)
	}
}

func TestRunnerTaskMapping(t *testing.T) {
	var cfg Config
	cfg.Tasks = map[string]TaskDef{"a": {Command: []string{"true"}, TimeoutMS: 100}}
	task := cfg.runnerTask("a", cfg.Tasks["a"])
	if task.Name != "a" || len(task.Command) != 1 || task.TimeoutMS != 100 {
		t.Fatalf("task=%+v", task)
	}
}
