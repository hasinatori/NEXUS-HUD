package main

// Zuletzt geaendert: 2026-08-17
// Bedingungen (IF) fuer Automation-Regeln — Tests.

import (
	"testing"
)

func TestConditionEmptyAlwaysMatches(t *testing.T) {
	c := Condition{}
	ctx := EvalContext{Profile: "dev"}
	if !c.Evaluate(ctx) {
		t.Fatal("leere Condition sollte immer matchen")
	}
}

func TestConditionProfile(t *testing.T) {
	c := Condition{Profile: "dev"}
	if !c.Evaluate(EvalContext{Profile: "dev"}) {
		t.Fatal("profile=dev sollte matchen")
	}
	if c.Evaluate(EvalContext{Profile: "gaming"}) {
		t.Fatal("profile=gaming sollte nicht matchen")
	}
}

func TestConditionWildcardProfile(t *testing.T) {
	c := Condition{Profile: "*"}
	for _, p := range []string{"dev", "gaming", "afk", ""} {
		if !c.Evaluate(EvalContext{Profile: p}) {
			t.Fatalf("wildcard sollte fuer %q matchen", p)
		}
	}
}

func TestConditionEventField(t *testing.T) {
	c := Condition{EventField: "build.ok", EventValue: "false"}
	state := map[string]string{"build.ok": "false"}
	if !c.Evaluate(EvalContext{EventState: state}) {
		t.Fatal("event field sollte matchen")
	}
	state["build.ok"] = "true"
	if c.Evaluate(EvalContext{EventState: state}) {
		t.Fatal("event value mismatch sollte nicht matchen")
	}
}

func TestConditionEventFieldMissing(t *testing.T) {
	c := Condition{EventField: "build.ok", EventValue: "false"}
	if c.Evaluate(EvalContext{EventState: nil}) {
		t.Fatal("nil state sollte nicht matchen")
	}
	if c.Evaluate(EvalContext{EventState: map[string]string{}}) {
		t.Fatal("leerer state sollte nicht matchen")
	}
}

func TestConditionMaxRuns(t *testing.T) {
	c := Condition{MaxRuns: 2, RunWindowMS: 60_000}
	now := int64(1000000)

	// Kein Limit bei leerem Verlauf
	if !c.Evaluate(EvalContext{RunHistory: nil}) {
		t.Fatal("leerer Verlauf sollte matchen")
	}

	// Ein Lauf im Fenster
	history := []RunRecord{{TaskName: "t", RunAt: now}}
	if !c.Evaluate(EvalContext{RunHistory: history}) {
		t.Fatal("1 Lauf bei max 2 sollte matchen")
	}

	// Zwei Laeufe im Fenster -> Limit erreicht
	history = []RunRecord{
		{TaskName: "t", RunAt: now},
		{TaskName: "t", RunAt: now + 100},
	}
	if c.Evaluate(EvalContext{RunHistory: history}) {
		t.Fatal("2 Laeufe bei max 2 sollte nicht matchen")
	}

	// Alter Lauf ausserhalb des Fensters
	history = []RunRecord{
		{TaskName: "t", RunAt: now - 70_000},
		{TaskName: "t", RunAt: now},
	}
	if !c.Evaluate(EvalContext{RunHistory: history}) {
		t.Fatal("1 aktueller Lauf (alter ausserhalb Fenster) sollte matchen")
	}
}

func TestConditionCombinedProfileAndEvent(t *testing.T) {
	c := Condition{Profile: "dev", EventField: "build.ok", EventValue: "false"}
	ctx := EvalContext{
		Profile:    "dev",
		EventState: map[string]string{"build.ok": "false"},
	}
	if !c.Evaluate(ctx) {
		t.Fatal("beide Bedingungen erfuellt sollte matchen")
	}
	ctx.Profile = "gaming"
	if c.Evaluate(ctx) {
		t.Fatal("falsches Profil sollte nicht matchen")
	}
	ctx.Profile = "dev"
	ctx.EventState["build.ok"] = "true"
	if c.Evaluate(ctx) {
		t.Fatal("falscher Event-Wert sollte nicht matchen")
	}
}
