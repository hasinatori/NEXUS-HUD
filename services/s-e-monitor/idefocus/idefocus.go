// Package idefocus beobachtet eine vom IDE geschriebene Focus-Datei und
// meldet Aenderungen der aktiven Datei als event.ide.focus.
package idefocus

// Zuletzt geändert: 2026-08-28

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"sync"
	"time"
)

// Focus beschreibt die aktuell fokussierte Datei im IDE.
// Das JSON wird vom IDE geschrieben (siehe extensions/vscode-nexus).
type Focus struct {
	Project  string `json:"project"`
	Filename string `json:"filename"`
	Language string `json:"language"`
	Path     string `json:"path"`
	At       string `json:"ts"`
}

// Params liefert die Schema-Params für event.ide.focus.
func (f *Focus) Params() map[string]any {
	return map[string]any{
		"project":  f.Project,
		"filename": f.Filename,
		"language": f.Language,
		"path":     f.Path,
		"ts":       f.At,
	}
}

// Watcher liest die Focus-Datei und liefert bei Aenderung den neuen Fokus.
type Watcher struct {
	Path string

	mu   sync.Mutex
	last *Focus
}

// New erstellt einen Watcher für die angegebene Focus-Datei.
func New(path string) *Watcher {
	return &Watcher{Path: path}
}

// Poll liest die Focus-Datei und liefert höchstens eine Aenderung zurück.
// Existiert die Datei nicht (mehr) oder ist sie leer, ist das Ergebnis nil
// ohne Fehler. Unlesbares JSON wird als Fehler gemeldet, ohne den letzten
// Stand zu verwerfen.
func (w *Watcher) Poll() (*Focus, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := os.ReadFile(w.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}

	var f Focus
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	if f.At == "" {
		f.At = time.Now().UTC().Format(time.RFC3339)
	}
	if reflect.DeepEqual(w.last, &f) {
		return nil, nil
	}
	cp := f
	w.last = &cp
	return &cp, nil
}
