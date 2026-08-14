package wsclient

// Zuletzt geändert: 2026-08-14

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func testPort(t *testing.T, ts *httptest.Server) int {
	t.Helper()
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("url parse: %v", err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatalf("port: %v", err)
	}
	return port
}

func TestClientConnectAndNotify(t *testing.T) {
	received := make(chan map[string]any, 8)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		go func() {
			for {
				_, data, err := conn.Read(context.Background())
				if err != nil {
					return
				}
				var msg map[string]any
				if json.Unmarshal(data, &msg) == nil {
					received <- msg
				}
			}
		}()
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	c := New(testPort(t, ts), "S-E", "s-e-monitor", "0.2.0-s-e.1")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	hello := <-received
	if hello["method"] != "event.system.hello" {
		t.Fatalf("erwartet hello, bekam %v", hello["method"])
	}
	params := hello["params"].(map[string]any)
	if params["service_id"] != "s-e-monitor" || params["source"] != "S-E" {
		t.Fatalf("hello params unerwartet: %v", params)
	}

	if err := c.Notify(ctx, "event.system.metrics", map[string]any{
		"cpu": 12.5,
		"ram": map[string]any{"used_mb": 100, "total_mb": 200},
	}); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	msg := <-received
	if msg["method"] != "event.system.metrics" {
		t.Fatalf("erwartet metrics, bekam %v", msg["method"])
	}
	p := msg["params"].(map[string]any)
	if _, ok := p["ts"]; !ok {
		t.Fatal("metrics params ohne ts")
	}
}

func TestClientReconnectsAfterDrop(t *testing.T) {
	var mu sync.Mutex
	connCount := 0
	received := make(chan string, 8)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		mu.Lock()
		n := connCount
		connCount++
		mu.Unlock()
		if n == 0 {
			go func() {
				if _, _, err := conn.Read(context.Background()); err == nil {
					_ = conn.Close(websocket.StatusPolicyViolation, "drop")
				}
			}()
			return
		}
		go func() {
			for {
				_, data, err := conn.Read(context.Background())
				if err != nil {
					return
				}
				received <- string(data)
			}
		}()
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	c := New(testPort(t, ts), "S-E", "s-e-monitor", "0.2.0-s-e.1")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	time.Sleep(200 * time.Millisecond)

	for i := 0; i < 2; i++ {
		if err := c.Notify(ctx, "event.system.metrics", map[string]any{
			"cpu": float64(i + 1),
			"ram": map[string]any{"used_mb": 100, "total_mb": 200},
		}); err != nil {
			t.Fatalf("Notify nach Drop: %v", err)
		}
	}

	select {
	case data := <-received:
		if data == "" {
			t.Fatal("leere Nachricht empfangen")
		}
	case <-ctx.Done():
		t.Fatal("Timeout: keine Nachricht nach Reconnect empfangen")
	}
}
