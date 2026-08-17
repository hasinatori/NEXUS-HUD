package main

// Zuletzt geaendert: 2026-08-17
// Bedingungen (IF) fuer Automation-Regeln. Ermoeglicht profil- und
// zustandsbasierte Filterung vor Task-Ausfuehrung.

// Condition repraesentiert eine IF-Bedingung, die vor dem THEN erfuellt
// sein muss. Alle Felder sind optional (leeres Feld = kein Filter).
type Condition struct {
	// Profile schraenkt die Regel auf einen bestimmten Context-Modus ein.
	// Gueltige Werte: "dev", "gaming", "afk", oder "*" (immer).
	Profile string `json:"profile,omitempty"`

	// EventField ermoeglicht Bedingungen auf Basis vergangener Events.
	// Z.B. {"field": "build.ok", "value": "false"} triggert nur bei Build-Fehler.
	EventField string `json:"event_field,omitempty"`
	EventValue string `json:"event_value,omitempty"`

	// MaxRuns begrenzt, wie oft die Regel innerhalb von RunWindowMS ausfuehrt werden darf.
	// 0 = unbegrenzt.
	MaxRuns    int   `json:"max_runs,omitempty"`
	RunWindowMS int64 `json:"run_window_ms,omitempty"`
}

// EvalContext enthaelt den aktuellen Systemzustand zur Auswertung von Bedingungen.
type EvalContext struct {
	Profile    string
	EventState map[string]string
	RunHistory []RunRecord
}

// RunRecord protokolliert einen vorherigen Task-Lauf fuer Rate-Limiting.
type RunRecord struct {
	TaskName string
	RunAt    int64 // Unix-Millis
}

// Evaluate prueft, ob die Bedingung im gegebenen Kontext erfuellt ist.
func (c Condition) Evaluate(ctx EvalContext) bool {
	if !c.matchProfile(ctx.Profile) {
		return false
	}
	if !c.matchEventField(ctx.EventState) {
		return false
	}
	if !c.matchRunLimit(ctx.RunHistory) {
		return false
	}
	return true
}

func (c Condition) matchProfile(current string) bool {
	if c.Profile == "" || c.Profile == "*" {
		return true
	}
	return c.Profile == current
}

func (c Condition) matchEventField(state map[string]string) bool {
	if c.EventField == "" {
		return true
	}
	if state == nil {
		return false
	}
	actual, ok := state[c.EventField]
	if !ok {
		return false
	}
	return actual == c.EventValue
}

func (c Condition) matchRunLimit(history []RunRecord) bool {
	if c.MaxRuns <= 0 || c.RunWindowMS <= 0 {
		return true
	}
	windowStart := int64(0)
	if len(history) > 0 {
		windowStart = history[len(history)-1].RunAt - c.RunWindowMS
	}
	count := 0
	for _, r := range history {
		if r.RunAt >= windowStart {
			count++
		}
	}
	return count < c.MaxRuns
}
