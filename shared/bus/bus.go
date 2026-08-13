// Package bus implementiert den lokalen WebSocket-Bus (Phase-1-Transport).
// Er bindet nur auf 127.0.0.1, validiert eingehende Nachrichten gegen die
// IPC-Regeln aus schema/events.schema.json und verteilt gültige Nachrichten
// an alle verbundenen Clients.
package bus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	DefaultPort      = 49152
	ProtocolVersion  = 1
	BusServiceID     = "bus"
	BusVersion       = "0.1.0"
	readWriteTimeout = 30 * time.Second
)

// Message ist eine JSON-RPC-2.0-Nachricht auf dem Bus.
type Message struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id,omitempty"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
}

// Hello erzeugt eine event.system.hello-Notification für die angegebene Quelle.
func Hello(source, serviceID, version string) Message {
	return Message{
		JSONRPC: "2.0",
		Method:  "event.system.hello",
		Params: map[string]any{
			"source":           source,
			"protocol_version": ProtocolVersion,
			"service_id":       serviceID,
			"version":          version,
			"ts":               time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// ErrorProtocol erzeugt eine error.protocol-Notification (ARCHITECTURE.md 3.5).
func ErrorProtocol(code int, message string) Message {
	return Message{
		JSONRPC: "2.0",
		Method:  "error.protocol",
		Params: map[string]any{
			"source":           BusServiceID,
			"protocol_version": ProtocolVersion,
			"code":             code,
			"message":          message,
			"ts":               time.Now().UTC().Format(time.RFC3339),
		},
	}
}

type client struct {
	conn *websocket.Conn
	addr string
	mu   sync.Mutex
}

// Server ist der lokale IPC-Bus.
type Server struct {
	Port int
	Log  *log.Logger

	mu      sync.Mutex
	clients map[*client]struct{}
}

// New erstellt einen Server mit Standardwerten.
func New() *Server {
	return &Server{
		Port:    DefaultPort,
		Log:     log.New(os.Stdout, "[bus] ", log.LstdFlags),
		clients: make(map[*client]struct{}),
	}
}

// ListenAndServe startet den Bus auf 127.0.0.1:<Port> und blockiert bis zum
// Beenden über ctx (Signal) oder einem Fehler.
func (s *Server) ListenAndServe(ctx context.Context) error {
	addr := fmt.Sprintf("127.0.0.1:%d", s.Port)
	srv := &http.Server{Addr: addr, Handler: s}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	s.Log.Printf("Bus lauscht auf ws://%s/ (protocol_version=%d)", addr, ProtocolVersion)
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// ServeHTTP akzeptiert eine WebSocket-Verbindung und betreut sie.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*", "127.0.0.1:*"},
	})
	if err != nil {
		s.Log.Printf("Verbindung abgelehnt: %v", err)
		return
	}

	c := &client{conn: conn, addr: r.RemoteAddr}
	s.add(c)
	s.Log.Printf("Client verbunden (%s)", c.addr)
	defer func() {
		s.remove(c)
		_ = conn.Close(websocket.StatusNormalClosure, "")
		s.Log.Printf("Client getrennt (%s)", c.addr)
	}()

	s.send(c, Hello(BusServiceID, BusServiceID, BusVersion))
	s.readLoop(c)
}

func (s *Server) readLoop(c *client) {
	for {
		_, data, err := c.conn.Read(context.Background())
		if err != nil {
			if !errors.Is(err, io.EOF) && websocket.CloseStatus(err) == -1 {
				s.Log.Printf("Lesefehler (%s): %v", c.addr, err)
			}
			return
		}

		msg, verr := Parse(data)
		if verr != nil {
			s.send(c, ErrorProtocol(verr.Code, verr.Message))
			s.Log.Printf("Ungueltige Nachricht abgelehnt (%s): code=%d %q", c.addr, verr.Code, verr.Message)
			if verr.Close {
				return
			}
			continue
		}

		if msg.Method == "event.system.hello" {
			s.Log.Printf("Hello empfangen: %s (Quelle %v)", msg.Params["service_id"], msg.Params["source"])
		}
		s.broadcast(c, data)
	}
}

func (s *Server) add(c *client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[c] = struct{}{}
}

func (s *Server) remove(c *client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, c)
}

func (s *Server) send(c *client, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), readWriteTimeout)
	defer cancel()
	if err := c.conn.Write(ctx, websocket.MessageText, data); err != nil {
		s.Log.Printf("Schreiben fehlgeschlagen (%s): %v", c.addr, err)
	}
}

func (s *Server) broadcast(from *client, raw []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for c := range s.clients {
		if c == from {
			continue
		}
		c.mu.Lock()
		ctx, cancel := context.WithTimeout(context.Background(), readWriteTimeout)
		err := c.conn.Write(ctx, websocket.MessageText, raw)
		cancel()
		c.mu.Unlock()
		if err != nil {
			s.Log.Printf("Broadcast fehlgeschlagen (%s): %v", c.addr, err)
		}
	}
}
