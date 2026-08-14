package main

// Zuletzt geändert: 2026-08-14

import (
	"testing"

	"github.com/hasinatori/NEXUS-HUD/shared/bus"
)

func TestParseExpects(t *testing.T) {
	e := parseExpects("event.system.hello:bus, event.system.metrics:S-E,event.build.failed")
	if len(e) != 3 {
		t.Fatalf("len=%d, want 3", len(e))
	}
	if e[0].method != "event.system.hello" || e[0].source != "bus" {
		t.Fatalf("e0=%+v", e[0])
	}
	if e[1].source != "S-E" {
		t.Fatalf("e1=%+v", e[1])
	}
	if e[2].source != "" {
		t.Fatalf("e2=%+v", e[2])
	}
	if len(parseExpects("")) != 0 || len(parseExpects(" , ,")) != 0 {
		t.Fatal("leere Liste soll 0 Elemente ergeben")
	}
}

func TestCheckExpects(t *testing.T) {
	seen := map[string]map[string]bool{
		"event.system.hello":   {"bus": true},
		"event.system.metrics": {"S-E": true, "X": true},
	}
	cases := []struct {
		name    string
		expects string
		missing int
	}{
		{"alle vorhanden", "event.system.hello:bus,event.system.metrics:S-E", 0},
		{"ohne Quelle reicht jede", "event.system.metrics", 0},
		{"Quelle fehlt", "event.system.hello:S-E", 1},
		{"Methode fehlt", "event.build.failed", 1},
	}
	for _, tc := range cases {
		got := checkExpects(seen, parseExpects(tc.expects))
		if len(got) != tc.missing {
			t.Errorf("%s: missing=%v, want %d", tc.name, got, tc.missing)
		}
	}
}

func TestRender(t *testing.T) {
	h := newHUD()
	h.apply(helloMsg("bus"))
	h.apply(helloMsg("S-E"))
	h.apply(metricsMsg())
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.hellos) != 2 {
		t.Fatalf("hellos=%v", h.hellos)
	}
	if h.metrics == nil || h.metrics.cpuPct != 12.5 {
		t.Fatalf("metrics=%+v", h.metrics)
	}
}

func helloMsg(source string) bus.Message {
	return bus.Message{JSONRPC: "2.0", Method: "event.system.hello", Params: map[string]any{"source": source}}
}

func metricsMsg() bus.Message {
	return bus.Message{JSONRPC: "2.0", Method: "event.system.metrics", Params: map[string]any{"source": "S-E", "cpu": 12.5, "used_mb": 2048.0, "total_mb": 8192.0}}
}
