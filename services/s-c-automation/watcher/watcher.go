// Package watcher überwacht Ordner per fsnotify und meldet Datei-Ereignisse
// (event.file.changed) an den Bus.
package watcher

// Zuletzt geändert: 2026-08-14

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// Handler wird bei jeder Dateiänderung mit dem absoluten Pfad und dem
// change-Typ (Schema-Enum) aufgerufen.
type Handler func(path, change string)

// Watch überwacht das angegebene Verzeichnis, bis ctx beendet wird.
func Watch(ctx context.Context, dir string, h Handler) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("fsnotify: %w", err)
	}
	defer w.Close()
	if err := w.Add(dir); err != nil {
		return fmt.Errorf("watch %s: %w", dir, err)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-w.Events:
			if !ok {
				return errors.New("watcher geschlossen")
			}
			if change := classify(event); change != "" {
				h(event.Name, change)
			}
		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}
			return err
		}
	}
}

func classify(e fsnotify.Event) string {
	switch {
	case e.Op&fsnotify.Create != 0:
		return "create"
	case e.Op&fsnotify.Write != 0:
		return "write"
	case e.Op&fsnotify.Remove != 0:
		return "remove"
	case e.Op&fsnotify.Rename != 0:
		return "rename"
	case e.Op&fsnotify.Chmod != 0:
		return "chmod"
	}
	return ""
}

// Match prüft, ob eine Dateiänderung zu einer Watcher-Regel passt.
func Match(dir, change string, path string, triggers []string) bool {
	if len(triggers) > 0 {
		ok := false
		for _, t := range triggers {
			if t == change {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	base := filepath.Clean(path)
	if filepath.Dir(base) == filepath.Clean(dir) || base == filepath.Clean(dir) {
		return true
	}
	return false
}
