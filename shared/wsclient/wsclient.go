// Package wsclient verbindet einen Go-Service mit dem lokalen Bus und
// sendet in der Phase 1 periodisch event.system.hello.
package wsclient

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/coder/websocket"

	"github.com/hasinatori/NEXUS-HUD/shared/bus"
)

const helloInterval = 5 * time.Second

// PortFromEnv liefert den Bus-Port aus NEXUS_WS_PORT oder den Standard.
func PortFromEnv() int {
	if p := os.Getenv("NEXUS_WS_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 && n <= 65535 {
			return n
		}
	}
	return bus.DefaultPort
}

// RunHelloLoop verbindet zum Bus und sendet periodisch event.system.hello.
// Die Verbindung wird automatisch wiederhergestellt (Retry alle 2 s).
func RunHelloLoop(ctx context.Context, source, serviceID, version string, port int) error {
	url := fmt.Sprintf("ws://127.0.0.1:%d/", port)
	for {
		err := runOnce(ctx, url, source, serviceID, version)
		if ctx.Err() != nil {
			return nil
		}
		fmt.Printf("[%s] Verbindung beendet (%v); neuer Versuch in 2 s\n", serviceID, err)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(2 * time.Second):
		}
	}
}

func runOnce(ctx context.Context, url, source, serviceID, version string) error {
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("Verbindung zu %s fehlgeschlagen: %w", url, err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")
	fmt.Printf("[%s] verbunden mit %s\n", serviceID, url)

	go discardIncoming(ctx, conn)

	ticker := time.NewTicker(helloInterval)
	defer ticker.Stop()

	if err := sendHello(ctx, conn, source, serviceID, version); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := sendHello(ctx, conn, source, serviceID, version); err != nil {
				return err
			}
		}
	}
}

func sendHello(ctx context.Context, conn *websocket.Conn, source, serviceID, version string) error {
	data, err := json.Marshal(bus.Hello(source, serviceID, version))
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, data)
}

func discardIncoming(ctx context.Context, conn *websocket.Conn) {
	for {
		if _, _, err := conn.Read(ctx); err != nil {
			return
		}
	}
}
