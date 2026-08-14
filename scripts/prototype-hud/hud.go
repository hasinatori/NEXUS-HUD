package main

// Zuletzt geändert: 2026-08-14

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hasinatori/NEXUS-HUD/shared/bus"
	"github.com/hasinatori/NEXUS-HUD/shared/wsclient"
)

func jsonUnmarshal(raw []byte, v any) error { return json.Unmarshal(raw, v) }

type hudState struct {
	mu      sync.Mutex
	hellos  map[string]int // Quelle -> Anzahl
	metrics *metricsView
	git     *gitView
	build   *buildView
	others  map[string]int // Method -> Anzahl
	busUp   bool
}

type metricsView struct {
	cpuPct  float64
	usedMB  float64
	totalMB float64
}
type gitView struct {
	branch      string
	staged      int
	uncommitted int
	ahead       int
	behind      int
}
type buildView struct {
	ok     bool
	output string
}

func newHUD() *hudState {
	return &hudState{hellos: map[string]int{}, others: map[string]int{}}
}

func (h *hudState) apply(m bus.Message) {
	params := m.Params
	h.mu.Lock()
	defer h.mu.Unlock()
	source := fmt.Sprint(params["source"])
	switch m.Method {
	case "event.system.hello":
		h.hellos[source]++
	case "event.system.metrics":
		h.metrics = &metricsView{cpuPct: num(params["cpu"]), usedMB: num(params["used_mb"]), totalMB: num(params["total_mb"])}
	case "event.git.status":
		h.git = &gitView{
			branch:      fmt.Sprint(params["branch"]),
			staged:      int(num(params["staged"])),
			uncommitted: int(num(params["uncommitted"])),
			ahead:       int(num(params["ahead"])),
			behind:      int(num(params["behind"])),
		}
	case "event.build.succeeded":
		h.build = &buildView{ok: true, output: fmt.Sprint(params["output"])}
	case "event.build.failed":
		h.build = &buildView{ok: false, output: fmt.Sprint(params["output"])}
	default:
		h.others[m.Method]++
	}
}

func num(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	}
	return 0
}

func (h *hudState) render() {
	h.mu.Lock()
	defer h.mu.Unlock()
	var b strings.Builder
	fmt.Fprintf(&b, "\x1b[2J\x1b[H\x1b[36mNEXUS HUD — Prototyp (Phase 2)\x1b[0m\n\n")
	fmt.Fprintf(&b, " \x1b[1mHello der Dienste\x1b[0m  ")
	if len(h.hellos) == 0 {
		b.WriteString("—")
	} else {
		var keys []string
		for k := range h.hellos {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%s (%d)", k, h.hellos[k])
		}
	}
	fmt.Fprintf(&b, "\n\n \x1b[1mSystem-Metriken (S-E)\x1b[0m\n")
	if h.metrics == nil {
		b.WriteString("   noch keine Daten\n")
	} else {
		fmt.Fprintf(&b, "   CPU %5.1f%%   RAM %6.0f / %6.0f MB\n", h.metrics.cpuPct, h.metrics.usedMB, h.metrics.totalMB)
	}
	fmt.Fprintf(&b, "\n \x1b[1mGit-Status (S-E)\x1b[0m\n")
	if h.git == nil {
		b.WriteString("   noch keine Daten\n")
	} else {
		fmt.Fprintf(&b, "   %s  staged %d  uncommitted %d  ↑%d ↓%d\n", h.git.branch, h.git.staged, h.git.uncommitted, h.git.ahead, h.git.behind)
	}
	fmt.Fprintf(&b, "\n \x1b[1mBuild (S-E)\x1b[0m\n")
	if h.build == nil {
		b.WriteString("   noch keine Daten\n")
	} else {
		col, word := "\x1b[32m", "OK"
		if !h.build.ok {
			col, word = "\x1b[31m", "FEHLER"
		}
		fmt.Fprintf(&b, "   %s%s\x1b[0m  %s\n", col, word, h.build.output)
	}
	fmt.Fprintf(&b, "\n \x1b[1mWeitere Events\x1b[0m  ")
	if len(h.others) == 0 {
		b.WriteString("—")
	} else {
		var keys []string
		for k := range h.others {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%s (%d)", k, h.others[k])
		}
	}
	fmt.Fprint(&b, "\n\n (Strg+C zum Beenden)\n")
	os.Stdout.WriteString(b.String())
}

func runHUD(ctx context.Context, c *wsclient.Client, events <-chan bus.Message) {
	state := newHUD()
	state.render()
	go func() {
		for m := range events {
			state.apply(m)
			state.render()
		}
	}()
	<-ctx.Done()
	c.Close()
}

func runTest(ctx context.Context, c *wsclient.Client, events <-chan bus.Message, window time.Duration, expects []expect) int {
	seen := map[string]map[string]bool{}
	var mu sync.Mutex

	collect := func(m bus.Message) {
		mu.Lock()
		defer mu.Unlock()
		if seen[m.Method] == nil {
			seen[m.Method] = map[string]bool{}
		}
		seen[m.Method][fmt.Sprint(m.Params["source"])] = true
	}

	done := time.After(window)
	for {
		select {
		case <-ctx.Done():
			c.Close()
			fmt.Println("abgebrochen")
			return 1
		case <-done:
			c.Close()
			missing := checkExpects(seen, expects)
			fmt.Printf("Testfenster (%s) beendet. Empfangene Events:\n", window)
			var methods []string
			for m := range seen {
				methods = append(methods, m)
			}
			sort.Strings(methods)
			for _, m := range methods {
				var sources []string
				for s := range seen[m] {
					sources = append(sources, s)
				}
				sort.Strings(sources)
				fmt.Printf("  %s von %s\n", m, strings.Join(sources, ", "))
			}
			if len(missing) > 0 {
				fmt.Fprintf(os.Stderr, "FEHLER — fehlende Events: %s\n", strings.Join(missing, ", "))
				return 1
			}
			fmt.Println("OK — alle erwarteten Events empfangen.")
			return 0
		case m := <-events:
			collect(m)
		}
	}
}
