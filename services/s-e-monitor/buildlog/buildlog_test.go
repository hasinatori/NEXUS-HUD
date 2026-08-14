package buildlog

// Zuletzt geändert: 2026-08-14

import (
	"os"
	"path/filepath"
	"testing"
)

func writeLog(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write log: %v", err)
	}
}

func TestBuildlogStateChange(t *testing.T) {
	log := filepath.Join(t.TempDir(), "build.log")
	writeLog(t, log, "Kompiliere...\n")
	w := New(log, "testprojekt")

	res, err := w.Poll()
	if err != nil {
		t.Fatalf("Poll: %v", err)
	}
	if res != nil {
		t.Fatalf("keine Build-Zeile, aber Ergebnis: %+v", res)
	}

	writeLog(t, log, "Kompiliere...\nBUILD SUCCESSFUL\n")
	res, err = w.Poll()
	if err != nil {
		t.Fatalf("Poll 2: %v", err)
	}
	if res == nil || !res.Ok {
		t.Fatalf("erwartet Erfolg, bekam %+v", res)
	}

	// Gleicher Zustand erneut -> kein Ergebnis (kein Wechsel).
	writeLog(t, log, "BUILD SUCCESSFUL\n")
	res, err = w.Poll()
	if err != nil {
		t.Fatalf("Poll 3: %v", err)
	}
	if res != nil {
		t.Fatalf("kein Wechsel erwartet, bekam %+v", res)
	}

	writeLog(t, log, "src/main.go:12: error: undefined: foo\nBUILD FAILED\n")
	res, err = w.Poll()
	if err != nil {
		t.Fatalf("Poll 4: %v", err)
	}
	if res == nil || res.Ok {
		t.Fatalf("erwartet Fehlschlag, bekam %+v", res)
	}
}

func TestBuildlogMissingFile(t *testing.T) {
	w := New(filepath.Join(t.TempDir(), "gibt-es-nicht.log"), "x")
	res, err := w.Poll()
	if err != nil {
		t.Fatalf("Poll bei fehlendem File: %v", err)
	}
	if res != nil {
		t.Fatalf("kein Ergebnis erwartet, bekam %+v", res)
	}
}

func TestClassify(t *testing.T) {
	w := New("", "")
	cases := []struct {
		line string
		want *bool
	}{
		{"BUILD SUCCESSFUL", boolPtr(true)},
		{"build passed in 1.2s", boolPtr(true)},
		{"src/main.go:12: error: undefined", boolPtr(false)},
		{"Fehler: Datei nicht gefunden", boolPtr(false)},
		{"normale Log-Zeile", nil},
	}
	for _, tc := range cases {
		got := classify(w, tc.line)
		if (got == nil) != (tc.want == nil) {
			t.Fatalf("Zeile %q: klassifiziert als %v, want %v", tc.line, got, tc.want)
		}
		if got != nil && *got != *tc.want {
			t.Fatalf("Zeile %q: Ergebnis %v, want %v", tc.line, *got, *tc.want)
		}
	}
}

func boolPtr(b bool) *bool { return &b }
