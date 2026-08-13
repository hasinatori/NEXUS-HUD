package bus

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coder/websocket"
)

func TestParseValidHello(t *testing.T) {
	raw := []byte(`{"jsonrpc":"2.0","method":"event.system.hello","params":{"source":"S-B","protocol_version":1,"service_id":"s-b-macro-launchpad","version":"0.1.0","ts":"2026-08-13T10:00:00Z"}}`)
	msg, verr := Parse(raw)
	if verr != nil {
		t.Fatalf("unexpected protocol error: %+v", verr)
	}
	if msg.Method != "event.system.hello" {
		t.Fatalf("method = %q, want event.system.hello", msg.Method)
	}
}

func TestParseBusHello(t *testing.T) {
	raw := []byte(`{"jsonrpc":"2.0","method":"event.system.hello","params":{"source":"bus","protocol_version":1,"service_id":"bus"}}`)
	if _, verr := Parse(raw); verr != nil {
		t.Fatalf("bus source should be valid: %+v", verr)
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		code int
	}{
		{"bad json", `{kein json`, CodeInvalidRequest},
		{"wrong jsonrpc", `{"jsonrpc":"1.0","method":"event.system.hello","params":{"source":"S-A","protocol_version":1}}`, CodeInvalidRequest},
		{"missing method", `{"jsonrpc":"2.0","params":{"source":"S-A","protocol_version":1}}`, CodeInvalidRequest},
		{"unknown method", `{"jsonrpc":"2.0","method":"event.unknown","params":{"source":"S-A","protocol_version":1}}`, CodeMethodNotFound},
		{"missing params", `{"jsonrpc":"2.0","method":"event.system.hello"}`, CodeInvalidParams},
		{"invalid source", `{"jsonrpc":"2.0","method":"event.system.hello","params":{"source":"S-F","protocol_version":1}}`, CodeInvalidParams},
		{"wrong version", `{"jsonrpc":"2.0","method":"event.system.hello","params":{"source":"S-A","protocol_version":2}}`, CodeInvalidParams},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, verr := Parse([]byte(tc.raw))
			if verr == nil {
				t.Fatal("expected protocol error, got none")
			}
			if verr.Code != tc.code {
				t.Fatalf("code = %d, want %d", verr.Code, tc.code)
			}
			if tc.name == "wrong version" && !verr.Close {
				t.Fatal("wrong version must close the connection")
			}
		})
	}
}

func dial(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", url, err)
	}
	t.Cleanup(func() { _ = conn.Close(websocket.StatusNormalClosure, "") })
	return conn
}

func readMethod(t *testing.T, conn *websocket.Conn) (string, map[string]any) {
	t.Helper()
	ctx := context.Background()
	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return m.Method, m.Params
}

func TestServerBroadcastAndErrors(t *testing.T) {
	s := New()
	ts := httptest.NewServer(s)
	defer ts.Close()
	url := "ws" + strings.TrimPrefix(ts.URL, "http")

	a := dial(t, url)
	b := dial(t, url)

	method, params := readMethod(t, a)
	if method != "event.system.hello" || params["service_id"] != BusServiceID {
		t.Fatalf("client a: erwartet Bus-Hello, bekam method=%s service_id=%v", method, params["service_id"])
	}
	readMethod(t, b)

	if err := a.Write(context.Background(), websocket.MessageText, []byte(`{"jsonrpc":"2.0","method":"event.system.hello","params":{"source":"S-C","protocol_version":1,"service_id":"s-c-automation","version":"0.1.0","ts":"x"}}`)); err != nil {
		t.Fatalf("write a: %v", err)
	}

	method, params = readMethod(t, b)
	if method != "event.system.hello" || params["service_id"] != "s-c-automation" {
		t.Fatalf("client b: erwartet Broadcast des S-C-Hellos, bekam method=%s service_id=%v", method, params["service_id"])
	}

	if err := a.Write(context.Background(), websocket.MessageText, []byte(`not-json`)); err != nil {
		t.Fatalf("write invalid a: %v", err)
	}
	method, params = readMethod(t, a)
	if method != "error.protocol" {
		t.Fatalf("client a: erwartet error.protocol, bekam %s", method)
	}
	if code, ok := params["code"].(float64); !ok || int(code) != CodeInvalidRequest {
		t.Fatalf("error.protocol code = %v, want %d", params["code"], CodeInvalidRequest)
	}
}

func TestServerClosesOnVersionMismatch(t *testing.T) {
	s := New()
	ts := httptest.NewServer(s)
	defer ts.Close()
	url := "ws" + strings.TrimPrefix(ts.URL, "http")

	a := dial(t, url)
	_, _ = readMethod(t, a)

	if err := a.Write(context.Background(), websocket.MessageText, []byte(`{"jsonrpc":"2.0","method":"event.system.hello","params":{"source":"S-A","protocol_version":2}}`)); err != nil {
		t.Fatalf("write a: %v", err)
	}

	method, _ := readMethod(t, a)
	if method != "error.protocol" {
		t.Fatalf("erwartet error.protocol vor Verbindungsende, bekam %s", method)
	}
	if _, _, err := a.Read(context.Background()); err == nil {
		t.Fatal("Verbindung sollte nach protocol_version-Mismatch geschlossen werden")
	}
}
