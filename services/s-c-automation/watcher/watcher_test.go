package watcher

// Zuletzt geändert: 2026-08-14

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func fsnotifyEvent(op string) fsnotify.Event {
	var o fsnotify.Op
	switch op {
	case "CREATE":
		o = fsnotify.Create
	case "WRITE":
		o = fsnotify.Write
	case "REMOVE":
		o = fsnotify.Remove
	case "RENAME":
		o = fsnotify.Rename
	case "CHMOD":
		o = fsnotify.Chmod
	}
	return fsnotify.Event{Name: "/tmp/x", Op: o}
}

func TestWatchSeesCreate(t *testing.T) {
	dir := t.TempDir()
	events := make(chan [2]string, 4)
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = Watch(t.Context(), dir, func(path, change string) {
			events <- [2]string{path, change}
		})
		wg.Done()
	}()

	// Warten, bis der Watcher aktiv ist, dann Datei anlegen.
	time.Sleep(300 * time.Millisecond)
	if err := os.WriteFile(filepath.Join(dir, "neu.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	select {
	case ev := <-events:
		if ev[1] != "create" {
			t.Fatalf("change=%q, want create", ev[1])
		}
	case <-time.After(3 * time.Second):
		t.Fatal("kein create-Event empfangen")
	}
	_ = done
}

func TestClassify(t *testing.T) {
	cases := []struct {
		op   string
		want string
	}{
		{"CREATE", "create"},
		{"WRITE", "write"},
		{"REMOVE", "remove"},
		{"RENAME", "rename"},
		{"CHMOD", "chmod"},
	}
	for _, tc := range cases {
		got := classify(fsnotifyEvent(tc.op))
		if got != tc.want {
			t.Fatalf("op %s -> %q, want %q", tc.op, got, tc.want)
		}
	}
}

func TestMatch(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "unter")
	os.Mkdir(sub, 0o755)

	if !Match(dir, "write", filepath.Join(dir, "a.txt"), nil) {
		t.Error("direkte Datei soll matchen")
	}
	if Match(dir, "remove", filepath.Join(dir, "a.txt"), []string{"write"}) {
		t.Error("Trigger remove soll bei write nicht matchen")
	}
	if Match(dir, "write", filepath.Join(sub, "a.txt"), nil) {
		t.Error("Unterordner soll (noch) nicht matchen")
	}
}
