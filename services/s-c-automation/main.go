// Command s-c-automation ist die Automation Engine (S-C): File-Watcher,
// Task-Runner und IF-THIS-THEN-THAT-Regeln (JSON-Konfiguration).
package main

// Zuletzt geaendert: 2026-08-29

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
	"time"

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
	profilesPath := flag.String("profiles", "", "Pfad zur Profile-Konfiguration (optional)")
	eventRulesPath := flag.String("event-rules", "", "Pfad zur Event-Regeln-Konfiguration (optional)")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.SCAutomation)
		return
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("%s: %v", serviceID, err)
	}

	// Profile laden (optional)
	var profilesCfg ProfilesConfig
	if *profilesPath != "" {
		profilesCfg, err = loadProfiles(*profilesPath)
		if err != nil {
			log.Fatalf("%s: Profiles: %v", serviceID, err)
		}
	} else {
		// Standardprofil erzeugen
		profilesCfg = ProfilesConfig{
			Profiles: []ProfileDef{
				{Name: "dev"},
				{Name: "gaming"},
				{Name: "afk"},
			},
		}
	}
	pm := NewProfileManager(profilesCfg)

	// Event-Regeln laden (optional)
	eventRules, err := loadEventRules(*eventRulesPath)
	if err != nil {
		log.Fatalf("%s: Event-Rules: %v", serviceID, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialisiere StateStore fuer IF-Bedingungen
	state := &StateStore{
		Profile:    "dev",
		EventState: map[string]string{},
	}
	timeNowUnixMillis = func() int64 { return time.Now().UnixMilli() }

	log.Printf("%s (%s) %s startet, Bus-Port %d, %d Tasks, %d Watcher, %d Profiles, %d Event-Rules",
		serviceID, source, version.SCAutomation, *port,
		len(cfg.Tasks), len(cfg.Watchers), len(pm.Profiles), len(eventRules.Rules))

	c := wsclient.New(*port, source, serviceID, version.SCAutomation)
	c.OnMessage = func(raw json.RawMessage) { handleMessage(ctx, c, cfg, eventRules, raw, state, pm) }
	if err := c.Connect(ctx); err != nil {
		log.Fatalf("%s: %v", serviceID, err)
	}
	go c.RunHelloLoop(ctx)

	r := runner.New(sink{c: c})
	startWatchers(ctx, c, cfg, r, state, pm)

	<-ctx.Done()
	c.Close()
}

func handleMessage(ctx context.Context, c *wsclient.Client, cfg Config, eventRules EventRulesConfig,
	raw json.RawMessage, state *StateStore, pm *ProfileManager) {

	var m bus.Message
	if err := json.Unmarshal(raw, &m); err != nil {
		return
	}

	switch m.Method {
	case "cmd.automation.run":
		name, _ := m.Params["task"].(string)
		def, ok := cfg.Tasks[name]
		if !ok {
			log.Printf("[%s] Unbekannter Task %q", serviceID, name)
			return
		}
		log.Printf("[%s] cmd.automation.run -> Task %q", serviceID, name)
		go runner.New(sink{c: c}).Run(ctx, cfg.runnerTask(name, def))

	case "cmd.profile.switch":
		profile, _ := m.Params["profile"].(string)
		if profile != "" {
			if pm.SwitchTo(profile) {
				state.SetProfile(profile)
				log.Printf("[%s] Profil gewechselt: %s", serviceID, profile)
			} else {
				log.Printf("[%s] Unbekanntes Profil: %s", serviceID, profile)
			}
		}

	case "event.build.failed":
		state.SetEventField("build.ok", "false")
		project, _ := m.Params["project"].(string)
		if project != "" {
			state.SetEventField("build.project", project)
		}

	case "event.build.succeeded":
		state.SetEventField("build.ok", "true")
		project, _ := m.Params["project"].(string)
		if project != "" {
			state.SetEventField("build.project", project)
		}

	case "event.profile.switched":
		profile, _ := m.Params["profile"].(string)
		if profile != "" {
			pm.SwitchTo(profile)
			state.SetProfile(profile)
			log.Printf("[%s] Profil-Switch-Event empfangen: %s", serviceID, profile)
		}
	}

	// Event-Regeln auswerten
	for _, rule := range eventRules.Rules {
		if !rule.Match(m.Method) {
			continue
		}
		if rule.If != nil && !rule.If.Evaluate(state.EvalContext()) {
			log.Printf("[%s] Event-Regel %q: IF-Bedingung nicht erfuellt", serviceID, rule.Name)
			continue
		}
		if rule.Action.Cmd != "" && rule.Action.Target != "" {
			log.Printf("[%s] Event-Regel %q: %s -> %s (%s)", serviceID, rule.Name, m.Method, rule.Action.Target, rule.Action.Cmd)
			_ = c.Notify(ctx, "event.automation.rule.triggered", map[string]any{
				"rule_name":    rule.Name,
				"event_method": m.Method,
			})
			_ = c.Notify(ctx, rule.Action.Cmd, rule.Action.Params)
		}
	}
}

func startWatchers(ctx context.Context, c *wsclient.Client, cfg Config, r *runner.Runner, state *StateStore, pm *ProfileManager) {
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

				// IF-Bedingung auswerten (falls vorhanden)
				if w.If != nil && !w.If.Evaluate(state.EvalContext()) {
					log.Printf("[%s] %s %s -> Task %q (IF-Bedingung nicht erfuellt)", serviceID, change, filepath.Base(path), w.Then)
					return
				}

				log.Printf("[%s] %s %s -> Task %q", serviceID, change, filepath.Base(path), w.Then)
				_ = c.Notify(ctx, "event.file.changed", map[string]any{"path": path, "change": change})
				state.RecordRun(w.Then)
				go r.Run(ctx, cfg.runnerTask(w.Then, cfg.Tasks[w.Then]))
			})
			if err != nil && ctx.Err() == nil {
				log.Printf("[%s] Watcher %s: %v", serviceID, w.Path, err)
			}
		}()
	}
}
