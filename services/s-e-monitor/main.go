// Command s-e-monitor ist der Coding & Build Monitor (S-E).
package main

// Zuletzt geändert: 2026-08-14

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hasinatori/NEXUS-HUD/services/s-e-monitor/buildlog"
	"github.com/hasinatori/NEXUS-HUD/services/s-e-monitor/gitstatus"
	"github.com/hasinatori/NEXUS-HUD/services/s-e-monitor/metrics"
	"github.com/hasinatori/NEXUS-HUD/shared/version"
	"github.com/hasinatori/NEXUS-HUD/shared/wsclient"
)

const (
	source    = "S-E"
	serviceID = "s-e-monitor"
)

func main() {
	port := flag.Int("port", wsclient.PortFromEnv(), "Port des lokalen WebSocket-Bus")
	showVersion := flag.Bool("version", false, "Version ausgeben und beenden")
	metricsInterval := flag.Duration("metrics-interval", 5*time.Second, "Intervall für Systemmetriken (0 = aus)")
	gitDir := flag.String("git-dir", "", "Zu überwachendes Git-Repo (leer = aus)")
	gitInterval := flag.Duration("git-interval", 15*time.Second, "Intervall für den Git-Status")
	buildLog := flag.String("build-log", "", "Zu überwachendes Build-Log (leer = aus)")
	buildProject := flag.String("build-project", "build", "Projektname für Build-Events")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.SEMonitor)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("%s (%s) %s startet, Bus-Port %d", serviceID, source, version.SEMonitor, *port)

	c := wsclient.New(*port, source, serviceID, version.SEMonitor)
	if err := c.Connect(ctx); err != nil {
		log.Fatalf("%s: %v", serviceID, err)
	}
	go c.RunHelloLoop(ctx)

	if *metricsInterval > 0 {
		go runMetrics(ctx, c, *metricsInterval)
	}
	if *gitDir != "" {
		go runGit(ctx, c, *gitDir, *gitInterval)
	}
	if *buildLog != "" {
		go runBuildLog(ctx, c, *buildLog, *buildProject)
	}

	<-ctx.Done()
	c.Close()
}

func runMetrics(ctx context.Context, c *wsclient.Client, interval time.Duration) {
	r := metrics.NewReader()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
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
