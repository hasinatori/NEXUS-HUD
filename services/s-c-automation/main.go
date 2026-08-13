// Command s-c-automation ist der Phase-1-Stub der Automation Engine (S-C).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hasinatori/NEXUS-HUD/shared/version"
	"github.com/hasinatori/NEXUS-HUD/shared/wsclient"
)

const (
	source    = "S-C"
	serviceID = "s-c-automation"
)

func main() {
	port := flag.Int("port", wsclient.PortFromEnv(), "Port des lokalen WebSocket-Bus")
	showVersion := flag.Bool("version", false, "Version ausgeben und beenden")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.SCAutomation)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("%s (%s) %s startet, Bus-Port %d", serviceID, source, version.SCAutomation, *port)
	if err := wsclient.RunHelloLoop(ctx, source, serviceID, version.SCAutomation, *port); err != nil {
		log.Fatalf("%s: %v", serviceID, err)
	}
}
