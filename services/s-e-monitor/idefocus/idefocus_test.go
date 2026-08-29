package idefocus

// Zuletzt geändert: 2026-08-28

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFocus(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write focus: %v", err)
	}
}

const validJSON = `{
  "project": "NEXUS-HUD",
  "filename": "main.go",
  "language": "go",
  "path": "/home/sam/NEXUS/services/s-e-monitor/main.go",
  "ts": "2026-08-28T10:00:00Z"
}`

func TestFocusStateChange(t *testing.T) {
	file := filepath.Join(t.TempDir(), "ide-focus.json")
	writeFocus(t, file, validJSON)
	w := New(file)

	f, err := w.Poll()
	if err != nil {
		t.Fatalf("Poll 1: %v", err)
	}
	if f == nil {
		t.Fatal("erwartet Fokus, bekam nil")
	}
	if f.Filename != "main.go" || f.Project != "NEXUS-HUD" || f.Language != "go" {
		t.Fatalf("Fokus falsch geparst: %+v", f)
	}
	if f.Path == "" || f.At == "" {
		t.Fatalf("Pfad/ts fehlt: %+v", f)
	}

	// Gleicher Inhalt erneut -> kein Ergebnis (kein Wechsel).
	f, err = w.Poll()
	if err != nil {
		t.Fatalf("Poll 2: %v", err)
	}
	if f != nil {
		t.Fatalf("kein Wechsel erwartet, bekam %+v", f)
	}

	// Aenderung -> neuer Fokus.
	writeFocus(t, file, `{"project":"NEXUS-HUD","filename":"idefocus.go","language":"go","path":"/home/sam/NEXUS/services/s-e-monitor/idefocus/idefocus.go","ts":"2026-08-28T10:01:00Z"}`)
	f, err = w.Poll()
	if err != nil {
		t.Fatalf("Poll 3: %v", err)
	}
	if f == nil || f.Filename != "idefocus.go" {
		t.Fatalf("erwartet neuen Fokus, bekam %+v", f)
	}
}

func TestFocusMissingFile(t *testing.T) {
	w := New(filepath.Join(t.TempDir(), "gibt-es-nicht.json"))
	f, err := w.Poll()
	if err != nil {
		t.Fatalf("Poll bei fehlendem File: %v", err)
	}
	if f != nil {
		t.Fatalf("kein Ergebnis erwartet, bekam %+v", f)
	}
}

func TestFocusEmptyFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "ide-focus.json")
	writeFocus(t, file, "\n  \n")
	w := New(file)
	f, err := w.Poll()
	if err != nil {
		t.Fatalf("Poll bei leerem File: %v", err)
	}
	if f != nil {
		t.Fatalf("kein Ergebnis erwartet, bekam %+v", f)
	}
}

func TestFocusInvalidJSON(t *testing.T) {
	file := filepath.Join(t.TempDir(), "ide-focus.json")
	writeFocus(t, file, "{kaputt")
	w := New(file)
	if _, err := w.Poll(); err == nil {
		t.Fatal("erwartet Fehler bei kaputtem JSON")
	}

	// Nach Fehler wird der naechste gueltige Stand trotzdem gemeldet.
	writeFocus(t, file, validJSON)
	f, err := w.Poll()
	if err != nil {
		t.Fatalf("Poll nach Fehler: %v", err)
	}
	if f == nil || f.Filename != "main.go" {
		t.Fatalf("erwartet Fokus nach Erholung, bekam %+v", f)
	}
}

func TestFocusMissingTSIsStamped(t *testing.T) {
	file := filepath.Join(t.TempDir(), "ide-focus.json")
	writeFocus(t, file, `{"project":"X","filename":"a.go","language":"go","path":"/tmp/a.go"}`)
	w := New(file)
	f, err := w.Poll()
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if f == nil || f.At == "" {
		t.Fatalf("ts sollte gesetzt werden, bekam %+v", f)
	}
}

func TestFocusParams(t *testing.T) {
	f := &Focus{Project: "P", Filename: "f.go", Language: "go", Path: "/tmp/f.go", At: "2026-08-28T10:00:00Z"}
	params := f.Params()
	want := map[string]any{
		"project":  "P",
		"filename": "f.go",
		"language": "go",
		"path":     "/tmp/f.go",
		"ts":       "2026-08-28T10:00:00Z",
	}
	for k, v := range want {
		if params[k] != v {
			t.Fatalf("Params[%q] = %v, want %v", k, params[k], v)
		}
	}
	if len(params) != len(want) {
		t.Fatalf("Params hat %d Eintraege, want %d: %+v", len(params), len(want), params)
	}
}
