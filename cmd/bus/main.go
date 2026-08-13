// Command bus startet den lokalen WebSocket-Bus für Phase 1.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hasinatori/NEXUS-HUD/shared/bus"
	"github.com/hasinatori/NEXUS-HUD/shared/version"
)

func main() {
	port := flag.Int("port", bus.DefaultPort, "Port des lokalen WebSocket-Bus")
	showVersion := flag.Bool("version", false, "Version ausgeben und beenden")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Bus)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := bus.New()
	server.Port = *port
	if err := server.ListenAndServe(ctx); err != nil {
		log.Fatal(err)
	}
}
