// Package buildlog beobachtet ein Build-Log und erkennt Build-Enden
// (Erfolg/Fehler) als Zustandswechsel.
package buildlog

// Zuletzt geändert: 2026-08-14

import (
	"os"
	"regexp"
	"strings"
	"sync"
)

// Result beschreibt ein erkanntes Build-Ende.
type Result struct {
	Ok     bool
	Output string
}

// Watcher liest ein Log-File inkrementell und meldet Build-Zustandswechsel.
type Watcher struct {
	Path    string
	Project string

	mu        sync.Mutex
	offset    int64
	lastState *bool
	success   *regexp.Regexp
	failure   *regexp.Regexp
}

// New erstellt einen Watcher für das angegebene Log-File.
func New(path, project string) *Watcher {
	return &Watcher{
		Path:    path,
		Project: project,
		success: regexp.MustCompile(`(?i)(build succeeded|build ok|build successful|build passed|build complete|erfolgreich|\bsuccess\b)`),
		failure: regexp.MustCompile(`(?i)(\berror\b|\bfail(ed|ure)?\b|\bfehler\b)`),
	}
}

// Poll liest neue Zeilen und liefert höchstens einen Zustandswechsel zurück.
// Existiert das File (noch) nicht, ist das Ergebnis nil ohne Fehler.
func (w *Watcher) Poll() (*Result, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	f, err := os.Open(w.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	if fi, err := f.Stat(); err == nil && fi.Size() < w.offset {
		w.offset = 0 // Log wurde rotiert/truncatet
	}
	if _, err := f.Seek(w.offset, 0); err != nil {
		return nil, err
	}

	buf := make([]byte, 64*1024)
	n, err := f.Read(buf)
	if n == 0 {
		return nil, err
	}
	w.offset += int64(n)

	var result *Result
	for _, line := range strings.Split(string(buf[:n]), "\n") {
		state := classify(w, line)
		if state == nil {
			continue
		}
		if w.lastState == nil || *w.lastState != *state {
			w.lastState = state
			result = &Result{Ok: *state, Output: strings.TrimSpace(line)}
			break
		}
	}
	return result, err
}

func classify(w *Watcher, line string) *bool {
	if w.success.MatchString(line) {
		ok := true
		return &ok
	}
	if w.failure.MatchString(line) {
		ok := false
		return &ok
	}
	return nil
}
