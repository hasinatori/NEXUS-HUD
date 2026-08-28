package main

// Zuletzt geaendert: 2026-08-17
// Event-Regeln: Cross-Module-Automatisierung durch Bus-Event-Trigger.
// Erlaubt das Reagieren auf beliebige IPC-Events mit Bedingungen und Aktionen.

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// EventRulesConfig enthaelt alle Event-Regeln.
type EventRulesConfig struct {
	Rules []EventRule `json:"rules"`
}

// EventRule definiert eine Reaktion auf ein IPC-Event.
type EventRule struct {
	Name      string     `json:"name"`
	OnEvent   string     `json:"on_event"`
	If        *Condition `json:"if,omitempty"`
	Action    RuleAction `json:"action"`
}

// RuleAction beschreibt die auszufuehrende Aktion.
type RuleAction struct {
	// Cmd sendet einen Befehl an einen bestimmten Service.
	Cmd    string         `json:"cmd,omitempty"`
	Target string         `json:"target,omitempty"`
	Params map[string]any `json:"params,omitempty"`
}

// Match prueft ob das Regel-Event auf die gegebene Methode passt.
// Unterstuetzt exakten Match und Wildcard-Suffix (z.B. "event.build.*").
func (r EventRule) Match(method string) bool {
	if r.OnEvent == method {
		return true
	}
	if strings.HasSuffix(r.OnEvent, ".*") {
		prefix := strings.TrimSuffix(r.OnEvent, ".*")
		return strings.HasPrefix(method, prefix+".")
	}
	return false
}

// loadEventRules laedt die Event-Regeln aus einer JSON-Datei.
func loadEventRules(path string) (EventRulesConfig, error) {
	var cfg EventRulesConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, nil // Keine Datei ist erlaubt (keine Regeln)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("Event-Rules parsen: %w", err)
	}
	seen := map[string]bool{}
	for _, r := range cfg.Rules {
		if r.Name == "" {
			return cfg, fmt.Errorf("Event-Rules: Regel ohne Name")
		}
		if seen[r.Name] {
			return cfg, fmt.Errorf("Event-Rules: doppelte Regel %q", r.Name)
		}
		seen[r.Name] = true
		if r.OnEvent == "" {
			return cfg, fmt.Errorf("Event-Rules: Regel %q ohne on_event", r.Name)
		}
	}
	return cfg, nil
}
