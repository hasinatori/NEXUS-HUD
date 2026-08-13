// Command hello-check prüft die Phase-1-Verifikation: Alle 5 Services
// (Bus + S-B bis S-E) müssen innerhalb des Timeouts event.system.hello senden.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/coder/websocket"

	"github.com/hasinatori/NEXUS-HUD/shared/wsclient"
)

var expected = []string{
	busID,
	sbServiceID,
	scServiceID,
	sdServiceID,
	seServiceID,
}

const (
	busID       = "bus"
	sbServiceID = "s-b-macro-launchpad"
	scServiceID = "s-c-automation"
	sdServiceID = "s-d-integrations"
	seServiceID = "s-e-monitor"
)

func main() {
	port := flag.Int("port", wsclient.PortFromEnv(), "Port des lokalen WebSocket-Bus")
	timeout := flag.Duration("timeout", 15*time.Second, "Wartezeit für alle hellos")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	url := fmt.Sprintf("ws://127.0.0.1:%d/", *port)
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		log.Fatalf("Verbindung zum Bus fehlgeschlagen: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")
	fmt.Printf("hello-check: sammle hellos von %s (Timeout %s)\n", url, *timeout)

	seen := map[string]bool{}
	expect := make(map[string]bool, len(expected))
	for _, id := range expected {
		expect[id] = true
	}

	for ctx.Err() == nil {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Fatalf("Lesefehler: %v", err)
		}
		var msg struct {
			Method string `json:"method"`
			Params struct {
				ServiceID string `json:"service_id"`
			} `json:"params"`
		}
		if err := json.Unmarshal(data, &msg); err != nil || msg.Method != "event.system.hello" {
			continue
		}
		if _, wanted := expect[msg.Params.ServiceID]; wanted && !seen[msg.Params.ServiceID] {
			seen[msg.Params.ServiceID] = true
			fmt.Printf("  hello empfangen: %s\n", msg.Params.ServiceID)
		}
		if len(seen) == len(expected) {
			break
		}
	}

	fmt.Println("\nErgebnis:")
	all := true
	ids := append([]string(nil), expected...)
	sort.Strings(ids)
	for _, id := range ids {
		status := "OK"
		if !seen[id] {
			status = "FEHLT"
			all = false
		}
		fmt.Printf("  %-22s %s\n", id, status)
	}
	if all {
		fmt.Println("Alle 5 Services haben event.system.hello gesendet.")
		return
	}
	fmt.Println("Es fehlen hellos — laufen Bus und alle Services?")
	os.Exit(1)
}
