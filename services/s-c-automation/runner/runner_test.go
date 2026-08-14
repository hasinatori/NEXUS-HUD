package runner

// Zuletzt geändert: 2026-08-14

import (
	"context"
	"sync"
	"testing"
)

type fakeSink struct {
	mu       sync.Mutex
	started  []string
	finished []struct {
		name string
		code int
	}
}

func (f *fakeSink) AutomationStarted(_ context.Context, id, name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.started = append(f.started, name)
	_ = id
}

func (f *fakeSink) AutomationFinished(_ context.Context, id, name string, exitCode int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.finished = append(f.finished, struct {
		name string
		code int
	}{name, exitCode})
	_ = id
}

func (f *fakeSink) len() (int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.started), len(f.finished)
}

func TestRunEmitsStartedAndFinished(t *testing.T) {
	sink := &fakeSink{}
	r := New(sink)
	seq := 0
	r.NewID = func() string {
		seq++
		return "id" + string(rune('0'+seq))
	}

	r.Run(context.Background(), Task{Name: "test", Command: []string{"true"}})

	started, finished := sink.len()
	if started != 1 || finished != 1 {
		t.Fatalf("started=%d finished=%d, want 1/1", started, finished)
	}
	if sink.finished[0].code != 0 {
		t.Fatalf("exit_code=%d, want 0", sink.finished[0].code)
	}
}

func TestRunExitCodeFailure(t *testing.T) {
	sink := &fakeSink{}
	r := New(sink)
	r.Run(context.Background(), Task{Name: "fehlschlag", Command: []string{"sh", "-c", "exit 3"}})
	if sink.finished[0].code != 3 {
		t.Fatalf("exit_code=%d, want 3", sink.finished[0].code)
	}
}

func TestRunTimeout(t *testing.T) {
	sink := &fakeSink{}
	r := New(sink)
	// "sleep 5" wird per Timeout abgebrochen -> Exit-Code != 0, aber finished wird gemeldet.
	r.Run(context.Background(), Task{Name: "sleep", Command: []string{"sleep", "5"}, TimeoutMS: 50})
	started, finished := sink.len()
	if started != 1 || finished != 1 {
		t.Fatalf("started=%d finished=%d, want 1/1", started, finished)
	}
	if sink.finished[0].code == 0 {
		t.Fatal("Timeout-Abbruch sollte Exit-Code != 0 liefern")
	}
}
