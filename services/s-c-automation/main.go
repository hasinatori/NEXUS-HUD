// Command s-c-automation ist der Phase-1-Stub der Automation Engine (S-C).
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hasinatori/NEXUS-HUD/shared/wsclient"
)

const (
	source    = "S-C"
	serviceID = "s-c-automation"
	version   = "0.1.0"
)

func main() {
	port := flag.Int("port", wsclient.PortFromEnv(), "Port des lokalen WebSocket-Bus")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("%s (%s) startet, Bus-Port %d", serviceID, source, *port)
	if err := wsclient.RunHelloLoop(ctx, source, serviceID, version, *port); err != nil {
		log.Fatalf("%s: %v", serviceID, err)
	}
}
