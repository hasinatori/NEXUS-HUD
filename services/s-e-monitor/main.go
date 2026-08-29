// Command s-e-monitor ist der Coding & Build Monitor (S-E).
package main

// Zuletzt geändert: 2026-08-29

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hasinatori/NEXUS-HUD/services/s-e-monitor/buildlog"
	"github.com/hasinatori/NEXUS-HUD/services/s-e-monitor/gitstatus"
	"github.com/hasinatori/NEXUS-HUD/services/s-e-monitor/idefocus"
	"github.com/hasinatori/NEXUS-HUD/services/s-e-monitor/metrics"
	"github.com/hasinatori/NEXUS-HUD/shared/version"
	"github.com/hasinatori/NEXUS-HUD/shared/wsclient"
)

const (
	source    = "S-E"
	serviceID = "s-e-monitor"

	defaultGitInterval = 15 * time.Second
)

func main() {
	port := flag.Int("port", wsclient.PortFromEnv(), "Port des lokalen WebSocket-Bus")
	showVersion := flag.Bool("version", false, "Version ausgeben und beenden")
	metricsInterval := flag.Duration("metrics-interval", 5*time.Second, "Intervall für Systemmetriken (0 = aus)")
	gitDir := flag.String("git-dir", "", "Zu überwachendes Git-Repo (leer = aus)")
	gitInterval := flag.Duration("git-interval", defaultGitInterval, "Intervall für den Git-Status")
	buildLog := flag.String("build-log", "", "Zu überwachendes Build-Log (leer = aus)")
	buildProject := flag.String("build-project", "build", "Projektname für Build-Events")
	ideFocus := flag.String("ide-focus", "", "IDE-Focus-Datei, die der IDE-Plugin schreibt (leer = aus)")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.SEMonitor)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("%s (%s) %s startet, Bus-Port %d", serviceID, source, version.SEMonitor, *port)

	c := wsclient.New(*port, source, serviceID, version.SEMonitor)

	metricsChanges := make(chan time.Duration, 4)
	gitWatchDirs := make(chan string, 4)
	c.OnMessage = func(raw json.RawMessage) {
		handleMessage(raw, metricsChanges, gitWatchDirs)
	}

	if err := c.Connect(ctx); err != nil {
		log.Fatalf("%s: %v", serviceID, err)
	}
	go c.RunHelloLoop(ctx)

	go runMetrics(ctx, c, *metricsInterval, metricsChanges)
	go runGitWatcher(ctx, c, *gitDir, *gitInterval, gitWatchDirs)
	if *buildLog != "" {
		go runBuildLog(ctx, c, *buildLog, *buildProject)
	}
	if *ideFocus != "" {
		go runIdeFocus(ctx, c, *ideFocus)
	}

	<-ctx.Done()
	c.Close()
}

// handleMessage verarbeitet eingehende Commands (Empfängt: cmd.metrics.set_interval,
// cmd.git.watch). Andere Nachrichten werden ignoriert.
func handleMessage(raw json.RawMessage, metricsChanges chan<- time.Duration, gitWatchDirs chan<- string) {
	var m struct {
		Method string         `json:"method"`
		Params map[string]any `json:"params"`
	}
	if err := json.Unmarshal(raw, &m); err != nil || m.Params == nil {
		return
	}
	switch m.Method {
	case "cmd.metrics.set_interval":
		if ms, ok := metricsIntervalMs(m.Params["interval_ms"]); ok {
			log.Printf("[%s] cmd.metrics.set_interval -> %s", serviceID, ms)
			select {
			case metricsChanges <- ms:
			default:
			}
		}
	case "cmd.git.watch":
		if dir, ok := m.Params["path"].(string); ok && dir != "" {
			log.Printf("[%s] cmd.git.watch -> %s", serviceID, dir)
			select {
			case gitWatchDirs <- dir:
			default:
			}
		}
	}
}

// metricsIntervalMs wandelt einen interval_ms-Parameter in eine Duration um
// (0 = aus; negative/ungültige Werte werden verworfen).
func metricsIntervalMs(v any) (time.Duration, bool) {
	f, ok := v.(float64)
	if !ok || f < 0 {
		return 0, false
	}
	return time.Duration(f) * time.Millisecond, true
}

// runMetrics pollt die Systemmetriken. Intervalle können über den
// metricsChanges-Kanal dynamisch gesetzt werden (0 = Pause).
func runMetrics(ctx context.Context, c *wsclient.Client, interval time.Duration, changes <-chan time.Duration) {
	r := metrics.NewReader()
	interval = normalizeInterval(interval)
	var ticker *time.Ticker
	newTicker := func(d time.Duration) *time.Ticker {
		if ticker != nil {
			ticker.Stop()
		}
		if d <= 0 {
			return nil
		}
		return time.NewTicker(d)
	}
	ticker = newTicker(interval)
	defer func() {
		if ticker != nil {
			ticker.Stop()
		}
	}()

	var tickC <-chan time.Time
	if ticker != nil {
		tickC = ticker.C
	}
	for {
		select {
		case <-ctx.Done():
			return
		case d := <-changes:
			interval = normalizeInterval(d)
			ticker = newTicker(interval)
			tickC = nil
			if ticker != nil {
				tickC = ticker.C
			}
		case <-tickC:
			snap, err := r.Read()
			if err != nil {
				log.Printf("[%s] Metriken: %v", serviceID, err)
				continue
			}
			if err := c.Notify(ctx, "event.system.metrics", snap.Params()); err != nil && ctx.Err() == nil {
				log.Printf("[%s] Metriken senden: %v", serviceID, err)
			}
		}
	}
}

func normalizeInterval(d time.Duration) time.Duration {
	if d < 0 {
		return 0
	}
	return d
}

// runGitWatcher verwaltet alle überwachten Git-Repos (Initial aus -git-dir,
// weitere via cmd.git.watch) und meldet den Status je Repo im Intervall.
func runGitWatcher(ctx context.Context, c *wsclient.Client, dir string, interval time.Duration, add <-chan string) {
	watched := map[string]bool{}
	if dir != "" {
		watched[dir] = true
		go runGit(ctx, c, dir, interval)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case d := <-add:
			if d == "" || watched[d] {
				continue
			}
			watched[d] = true
			log.Printf("[%s] Git-Repo zusätzlich überwacht: %s", serviceID, d)
			go runGit(ctx, c, d, interval)
		}
	}
}

func runGit(ctx context.Context, c *wsclient.Client, dir string, interval time.Duration) {
	g := gitstatus.New(dir)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status, err := g.Read()
			if err != nil {
				log.Printf("[%s] Git-Status: %v", serviceID, err)
				continue
			}
			if err := c.Notify(ctx, "event.git.status", status.Params()); err != nil && ctx.Err() == nil {
				log.Printf("[%s] Git-Status senden: %v", serviceID, err)
			}
		}
	}
}

func runBuildLog(ctx context.Context, c *wsclient.Client, path, project string) {
	w := buildlog.New(path, project)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			res, err := w.Poll()
			if err != nil {
				log.Printf("[%s] Build-Log: %v", serviceID, err)
				continue
			}
			if res == nil {
				continue
			}
			method := "event.build.succeeded"
			if !res.Ok {
				method = "event.build.failed"
			}
			params := map[string]any{
				"project": project,
				"ok":      res.Ok,
				"output":  res.Output,
			}
			if err := c.Notify(ctx, method, params); err != nil && ctx.Err() == nil {
				log.Printf("[%s] Build-Event senden: %v", serviceID, err)
			}
		}
	}
}

func runIdeFocus(ctx context.Context, c *wsclient.Client, path string) {
	w := idefocus.New(path)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			focus, err := w.Poll()
			if err != nil {
				log.Printf("[%s] IDE-Fokus: %v", serviceID, err)
				continue
			}
			if focus == nil {
				continue
			}
			if err := c.Notify(ctx, "event.ide.focus", focus.Params()); err != nil && ctx.Err() == nil {
				log.Printf("[%s] IDE-Fokus senden: %v", serviceID, err)
			}
		}
	}
}
