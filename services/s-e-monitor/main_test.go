package main

// Zuletzt geändert: 2026-08-29

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMetricsIntervalMs(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want time.Duration
		ok   bool
	}{
		{"gueltig", float64(2000), 2 * time.Second, true},
		{"null = aus", float64(0), 0, true},
		{"negativ", float64(-5), 0, false},
		{"falscher Typ", "2000", 0, false},
		{"fehlt", nil, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := metricsIntervalMs(tc.in)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("metricsIntervalMs(%v) = %v, %v; want %v, %v", tc.in, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestNormalizeInterval(t *testing.T) {
	if normalizeInterval(-1) != 0 {
		t.Fatal("negative Intervalle müssen auf 0 (aus) fallen")
	}
	if normalizeInterval(1000) != 1000 {
		t.Fatal("positive Intervalle bleiben erhalten")
	}
}

func TestHandleMessageMetricsInterval(t *testing.T) {
	ch := make(chan time.Duration, 4)
	gitCh := make(chan string, 4)
	raw, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "cmd.metrics.set_interval",
		"params":  map[string]any{"interval_ms": 1500},
	})
	handleMessage(raw, ch, gitCh)

	select {
	case d := <-ch:
		if d != 1500*time.Millisecond {
			t.Fatalf("Intervall = %v, want 1.5s", d)
		}
	default:
		t.Fatal("kein Intervall auf dem Kanal")
	}
	if len(gitCh) != 0 {
		t.Fatal("cmd.metrics muss nichts auf den Git-Kanal legen")
	}
}

func TestHandleMessageGitWatch(t *testing.T) {
	metricsCh := make(chan time.Duration, 4)
	gitCh := make(chan string, 4)
	raw, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "cmd.git.watch",
		"params":  map[string]any{"path": "/tmp/repo"},
	})
	handleMessage(raw, metricsCh, gitCh)

	select {
	case p := <-gitCh:
		if p != "/tmp/repo" {
			t.Fatalf("Pfad = %q, want /tmp/repo", p)
		}
	default:
		t.Fatal("kein Pfad auf dem Git-Kanal")
	}
	if len(metricsCh) != 0 {
		t.Fatal("cmd.git.watch muss nichts auf den Metrics-Kanal legen")
	}
}

func TestHandleMessageIgnoresUnknown(t *testing.T) {
	metricsCh := make(chan time.Duration, 4)
	gitCh := make(chan string, 4)
	raw, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "event.media.state",
		"params":  map[string]any{"playing": true},
	})
	handleMessage(raw, metricsCh, gitCh)
	if len(metricsCh) != 0 || len(gitCh) != 0 {
		t.Fatal("unbekannte Methoden müssen ignoriert werden")
	}
}

func TestHandleMessageRejectsBadInterval(t *testing.T) {
	metricsCh := make(chan time.Duration, 4)
	gitCh := make(chan string, 4)
	raw, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "cmd.metrics.set_interval",
		"params":  map[string]any{"interval_ms": -5},
	})
	handleMessage(raw, metricsCh, gitCh)
	if len(metricsCh) != 0 {
		t.Fatalf("negativ/ungültige Intervalle dürfen nicht durchgereicht werden: %d", len(metricsCh))
	}
}
