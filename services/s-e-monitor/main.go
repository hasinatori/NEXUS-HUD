// Command s-e-monitor ist der Phase-1-Stub des Coding & Build Monitor (S-E).
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
	source    = "S-E"
	serviceID = "s-e-monitor"
)

func main() {
	port := flag.Int("port", wsclient.PortFromEnv(), "Port des lokalen WebSocket-Bus")
	showVersion := flag.Bool("version", false, "Version ausgeben und beenden")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.SEMonitor)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("%s (%s) %s startet, Bus-Port %d", serviceID, source, version.SEMonitor, *port)
	if err := wsclient.RunHelloLoop(ctx, source, serviceID, version.SEMonitor, *port); err != nil {
		log.Fatalf("%s: %v", serviceID, err)
	}
}
