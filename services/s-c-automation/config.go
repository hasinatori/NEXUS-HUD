package main

// Zuletzt geändert: 2026-08-14

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hasinatori/NEXUS-HUD/services/s-c-automation/runner"
)

// Config ist die Automation-Konfiguration (IF-THIS-THEN-THAT, v1 als JSON).
type Config struct {
	Tasks    map[string]TaskDef `json:"tasks"`
	Watchers []WatcherDef       `json:"watchers"`
}

// TaskDef definiert einen ausführbaren Task.
type TaskDef struct {
	Command   []string `json:"command"`
	TimeoutMS int      `json:"timeout_ms"`
}

// WatcherDef verknüpft einen Ordner (Trigger) mit einem Task (Aktion).
type WatcherDef struct {
	Path     string   `json:"path"`
	Triggers []string `json:"triggers"`
	Then     string   `json:"then"`
}

func (c Config) runnerTask(name string, def TaskDef) runner.Task {
	return runner.Task{Name: name, Command: def.Command, TimeoutMS: def.TimeoutMS}
}

// watcherTrigger normalisiert Trigger-Angaben auf das Schema-Enum.
var watcherTrigger = func(triggers []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range triggers {
		norm := normalizeChange(t)
		if norm == "" || seen[norm] {
			continue
		}
		seen[norm] = true
		out = append(out, norm)
	}
	return out
}

func normalizeChange(c string) string {
	switch c {
	case "create", "write", "remove", "rename", "chmod":
		return c
	case "created", "changed":
		return "write"
	case "deleted":
		return "remove"
	}
	return ""
}

// loadConfig liest die Konfigurationsdatei.
func loadConfig(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("Config lesen: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("Config parsen: %w", err)
	}
	for _, w := range cfg.Watchers {
		if w.Path == "" {
			return cfg, fmt.Errorf("Config: Watcher ohne path")
		}
		if w.Then == "" {
			return cfg, fmt.Errorf("Config: Watcher %s ohne then-Task", w.Path)
		}
		if _, ok := cfg.Tasks[w.Then]; !ok {
			return cfg, fmt.Errorf("Config: then-Task %q existiert nicht", w.Then)
		}
	}
	return cfg, nil
}
