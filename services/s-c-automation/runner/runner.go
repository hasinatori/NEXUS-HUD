// Package runner führt S-C-Automation-Tasks aus und meldet den Ablauf als
// event.automation.started / event.automation.finished an den Bus.
package runner

// Zuletzt geändert: 2026-08-14

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// Task ist eine definierte Automation (aus der S-C-Konfiguration).
type Task struct {
	Name      string        `json:"name"`
	Command   []string      `json:"command"`
	Timeout   time.Duration `json:"-"`
	TimeoutMS int           `json:"timeout_ms"`
}

// EventSink meldet den Task-Zustand an den Bus.
type EventSink interface {
	AutomationStarted(ctx context.Context, id, name string)
	AutomationFinished(ctx context.Context, id, name string, exitCode int)
}

// ExecFunc führt eine Kommandozeile aus und liefert Exit-Code und Fehler.
// In Tests injizierbar, produktiv exec.CommandContext.
type ExecFunc func(ctx context.Context, command []string) (exitCode int, err error)

// Runner führt Tasks mit Zeitlimit aus.
type Runner struct {
	Sink  EventSink
	Exec  ExecFunc
	Now   func() time.Time
	NewID func() string
}

// New erstellt einen Runner mit Standard-Exec (echtes Kommando).
func New(sink EventSink) *Runner {
	return &Runner{
		Sink: sink,
		Exec: execCommand,
		Now:  time.Now,
		NewID: func() string {
			b := make([]byte, 4)
			if _, err := rand.Read(b); err != nil {
				return fmt.Sprintf("run-%d", time.Now().UnixNano())
			}
			return hex.EncodeToString(b)
		},
	}
}

// Run führt den Task aus: started-Event, Ausführung (Timeout), finished-Event.
func (r *Runner) Run(ctx context.Context, t Task) {
	if t.Timeout == 0 && t.TimeoutMS > 0 {
		t.Timeout = time.Duration(t.TimeoutMS) * time.Millisecond
	}
	id := r.NewID()
	runCtx := ctx
	cancel := func() {}
	if t.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, t.Timeout)
	}
	defer cancel()

	r.Sink.AutomationStarted(runCtx, id, t.Name)
	exitCode, _ := r.Exec(runCtx, t.Command)
	r.Sink.AutomationFinished(ctx, id, t.Name, exitCode)
}

func execCommand(ctx context.Context, command []string) (int, error) {
	if len(command) == 0 {
		return -1, fmt.Errorf("leeres Kommando")
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	err := cmd.Run()
	if err == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() >= 0 {
		return exitErr.ExitCode(), nil
	}
	return -1, err
}
