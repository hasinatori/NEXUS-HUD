// Command bus startet den lokalen WebSocket-Bus für Phase 1.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hasinatori/NEXUS-HUD/shared/bus"
)

func main() {
	port := flag.Int("port", bus.DefaultPort, "Port des lokalen WebSocket-Bus")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := bus.New()
	server.Port = *port
	if err := server.ListenAndServe(ctx); err != nil {
		log.Fatal(err)
	}
}
