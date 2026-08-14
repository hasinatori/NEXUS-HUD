package wsclient

// Zuletzt geändert: 2026-08-14

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/hasinatori/NEXUS-HUD/shared/bus"
)

const reconnectDelay = 2 * time.Second

// Client ist ein wiederverwendbarer Bus-Client mit Auto-Reconnect. Er sendet
// JSON-RPC-2.0-Notifications (ohne id) und hält optional einen Hello-Loop.
type Client struct {
	Port      int
	Source    string
	ServiceID string
	Version   string

	mu     sync.Mutex
	conn   *websocket.Conn
	closed bool
}

// New erstellt einen Client für die angegebene Quelle.
func New(port int, source, serviceID, version string) *Client {
	return &Client{Port: port, Source: source, ServiceID: serviceID, Version: version}
}

// Connect verbindet zum Bus und sendet das Start-Hello. Bei Fehlern wird alle
// reconnectDelay Sekunden erneut versucht, bis ctx beendet wird.
func (c *Client) Connect(ctx context.Context) error {
	for {
		err := c.ensureConnected(ctx)
		if err == nil {
			return c.SendHello(ctx)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		log.Printf("[%s] Verbindung fehlgeschlagen (%v); neuer Versuch", c.ServiceID, err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(reconnectDelay):
		}
	}
}

// SendHello sendet event.system.hello mit service_id und version.
func (c *Client) SendHello(ctx context.Context) error {
	return c.Send(ctx, "event.system.hello", map[string]any{
		"source":           c.Source,
		"protocol_version": bus.ProtocolVersion,
		"service_id":       c.ServiceID,
		"version":          c.Version,
		"ts":               time.Now().UTC().Format(time.RFC3339),
	})
}

// Notify sendet eine Notification mit Basis-params (source, protocol_version,
// ts) plus den zusätzlichen Feldern extra.
func (c *Client) Notify(ctx context.Context, method string, extra map[string]any) error {
	params := map[string]any{
		"source":           c.Source,
		"protocol_version": bus.ProtocolVersion,
		"ts":               time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range extra {
		params[k] = v
	}
	return c.Send(ctx, method, params)
}

// Send marshalt eine JSON-RPC-2.0-Notification und schreibt sie auf den Bus.
func (c *Client) Send(ctx context.Context, method string, params map[string]any) error {
	data, err := json.Marshal(bus.Message{JSONRPC: "2.0", Method: method, Params: params})
	if err != nil {
		return err
	}
	return c.Write(ctx, data)
}

// Write schreibt Rohdaten auf den Bus. Bei Fehlern wird die Verbindung
// invalidiert und einmalig erneut verbunden.
func (c *Client) Write(ctx context.Context, data []byte) error {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if err := c.ensureConnected(ctx); err != nil {
			lastErr = err
			continue
		}
		c.mu.Lock()
		lastErr = c.conn.Write(ctx, websocket.MessageText, data)
		c.mu.Unlock()
		if lastErr == nil {
			return nil
		}
		c.invalidate()
	}
	return lastErr
}

// RunHelloLoop sendet periodisch Hello (alle 5 s), solange ctx aktiv ist.
func (c *Client) RunHelloLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.SendHello(ctx); err != nil && ctx.Err() == nil {
				log.Printf("[%s] Hello fehlgeschlagen: %v", c.ServiceID, err)
			}
		}
	}
}

// Close schließt die Verbindung und verhindert weitere Versuche.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	if c.conn != nil {
		_ = c.conn.Close(websocket.StatusNormalClosure, "bye")
		c.conn = nil
	}
}

func (c *Client) url() string {
	return fmt.Sprintf("ws://127.0.0.1:%d/", c.Port)
}

func (c *Client) ensureConnected(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("Client geschlossen")
	}
	if c.conn != nil {
		return nil
	}
	conn, _, err := websocket.Dial(ctx, c.url(), nil)
	if err != nil {
		return err
	}
	c.conn = conn
	go c.readLoop(conn)
	fmt.Printf("[%s] verbunden mit %s\n", c.ServiceID, c.url())
	return nil
}

func (c *Client) readLoop(conn *websocket.Conn) {
	for {
		if _, _, err := conn.Read(context.Background()); err != nil {
			c.mu.Lock()
			if c.conn == conn {
				c.conn = nil
			}
			c.mu.Unlock()
			return
		}
	}
}

func (c *Client) invalidate() {
	c.mu.Lock()
	if c.conn != nil {
		_ = c.conn.Close(websocket.StatusNormalClosure, "")
		c.conn = nil
	}
	c.mu.Unlock()
}
