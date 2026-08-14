package main

// Zuletzt geändert: 2026-08-14

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hasinatori/NEXUS-HUD/shared/bus"
	"github.com/hasinatori/NEXUS-HUD/shared/wsclient"
)

const (
	source     = "S-A" // Prototyp-Stand-in für den S-A-HUD-Client
	serviceID  = "prototype-hud"
	hudVersion = "0.2.0"
)

type expect struct {
	method string
	source string
}

func parseExpects(list string) []expect {
	var out []expect
	for _, item := range strings.Split(list, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, ":", 2)
		e := expect{method: parts[0]}
		if len(parts) == 2 {
			e.source = parts[1]
		}
		out = append(out, e)
	}
	return out
}

// seen meldet, welche (method, source)-Paare während des Testfensters ankamen.
func checkExpects(seen map[string]map[string]bool, expects []expect) (missing []string) {
	for _, e := range expects {
		sources := seen[e.method]
		if len(sources) == 0 {
			missing = append(missing, e.method)
			continue
		}
		if e.source != "" && !sources[e.source] {
			missing = append(missing, e.method+":"+e.source)
		}
	}
	return missing
}

func main() {
	port := flag.Int("port", wsclient.PortFromEnv(), "Port des lokalen WebSocket-Bus")
	testMode := flag.Bool("test", false, "Testmodus: Events sammeln und Erwartungen prüfen")
	window := flag.Duration("window", 10*time.Second, "Testfenster (nur mit -test)")
	expects := flag.String("expect", "", "Erwartete Events (nur mit -test), Komma-getrennt: method[:source]")
	cmdMethod := flag.String("cmd", "", "Methode, die nach dem Verbinden gesendet wird (z. B. cmd.automation.run)")
	cmdParams := flag.String("cmd-params", "{}", "JSON-params für -cmd")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	c := wsclient.New(*port, source, serviceID, hudVersion)
	events := make(chan bus.Message, 64)
	c.OnMessage = func(raw json.RawMessage) {
		var m bus.Message
		if err := jsonUnmarshal(raw, &m); err != nil {
			return
		}
		events <- m
	}
	if err := c.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", serviceID, err)
		os.Exit(1)
	}

	if *cmdMethod != "" {
		var params map[string]any
		if err := json.Unmarshal([]byte(*cmdParams), &params); err != nil {
			fmt.Fprintf(os.Stderr, "%s: ungültige -cmd-params: %v\n", serviceID, err)
			os.Exit(1)
		}
		if params == nil {
			params = map[string]any{}
		}
		if err := c.Notify(ctx, *cmdMethod, params); err != nil {
			fmt.Fprintf(os.Stderr, "%s: -cmd senden: %v\n", serviceID, err)
			os.Exit(1)
		}
		fmt.Printf("[%s] gesendet: %s\n", serviceID, *cmdMethod)
	}

	if *testMode {
		os.Exit(runTest(ctx, c, events, *window, parseExpects(*expects)))
	}

	runHUD(ctx, c, events)
}
