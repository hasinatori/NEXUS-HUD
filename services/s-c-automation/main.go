// Command s-c-automation ist die Automation Engine (S-C): File-Watcher,
// Task-Runner und einfache IF-THIS-THEN-THAT-Regeln (JSON-Konfiguration).
package main

// Zuletzt geändert: 2026-08-14

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/hasinatori/NEXUS-HUD/services/s-c-automation/runner"
	"github.com/hasinatori/NEXUS-HUD/services/s-c-automation/watcher"
	"github.com/hasinatori/NEXUS-HUD/shared/bus"
	"github.com/hasinatori/NEXUS-HUD/shared/version"
	"github.com/hasinatori/NEXUS-HUD/shared/wsclient"
)

const (
	source    = "S-C"
	serviceID = "s-c-automation"
)

type sink struct{ c *wsclient.Client }

func (s sink) AutomationStarted(ctx context.Context, id, name string) {
	_ = s.c.Notify(ctx, "event.automation.started", map[string]any{"id": id, "name": name})
}

func (s sink) AutomationFinished(ctx context.Context, id, name string, exitCode int) {
	_ = s.c.Notify(ctx, "event.automation.finished", map[string]any{"id": id, "name": name, "exit_code": exitCode})
}

func main() {
	port := flag.Int("port", wsclient.PortFromEnv(), "Port des lokalen WebSocket-Bus")
	showVersion := flag.Bool("version", false, "Version ausgeben und beenden")
	configPath := flag.String("config", "services/s-c-automation/automations.json", "Pfad zur Automations-Konfiguration")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.SCAutomation)
		return
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("%s: %v", serviceID, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("%s (%s) %s startet, Bus-Port %d, %d Tasks, %d Watcher", serviceID, source, version.SCAutomation, *port, len(cfg.Tasks), len(cfg.Watchers))

	c := wsclient.New(*port, source, serviceID, version.SCAutomation)
	c.OnMessage = func(raw json.RawMessage) { handleMessage(ctx, c, cfg, raw) }
	if err := c.Connect(ctx); err != nil {
		log.Fatalf("%s: %v", serviceID, err)
	}
	go c.RunHelloLoop(ctx)

	r := runner.New(sink{c: c})
	startWatchers(ctx, c, cfg, r)

	<-ctx.Done()
	c.Close()
}

func handleMessage(ctx context.Context, c *wsclient.Client, cfg Config, raw json.RawMessage) {
	var m bus.Message
	if err := json.Unmarshal(raw, &m); err != nil {
		return
	}
	if m.Method != "cmd.automation.run" {
		return
	}
	name, _ := m.Params["task"].(string)
	def, ok := cfg.Tasks[name]
	if !ok {
		log.Printf("[%s] Unbekannter Task %q", serviceID, name)
		return
	}
	log.Printf("[%s] cmd.automation.run -> Task %q", serviceID, name)
	go runner.New(sink{c: c}).Run(ctx, cfg.runnerTask(name, def))
}

func startWatchers(ctx context.Context, c *wsclient.Client, cfg Config, r *runner.Runner) {
	var wg sync.WaitGroup
	for _, wd := range cfg.Watchers {
		wg.Add(1)
		w := wd
		triggers := watcherTrigger(w.Triggers)
		go func() {
			defer wg.Done()
			log.Printf("[%s] Watcher auf %s (Triggers: %v -> %s)", serviceID, w.Path, triggers, w.Then)
			err := watcher.Watch(ctx, w.Path, func(path, change string) {
				if !watcher.Match(w.Path, change, path, triggers) {
					return
				}
				log.Printf("[%s] %s %s -> Task %q", serviceID, change, filepath.Base(path), w.Then)
				_ = c.Notify(ctx, "event.file.changed", map[string]any{"path": path, "change": change})
				go r.Run(ctx, cfg.runnerTask(w.Then, cfg.Tasks[w.Then]))
			})
			if err != nil && ctx.Err() == nil {
				log.Printf("[%s] Watcher %s: %v", serviceID, w.Path, err)
			}
		}()
	}
}
