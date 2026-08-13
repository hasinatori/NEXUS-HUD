package bus

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// ProtocolError beschreibt eine abgelehnte Nachricht.
type ProtocolError struct {
	Code    int
	Message string
	Close   bool
}

const (
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
)

var (
	sourcePattern = regexp.MustCompile(`^(S-[A-E]|bus)$`)
	methodCatalog = map[string]bool{
		"event.system.hello":        true,
		"event.system.heartbeat":    true,
		"event.system.metrics":      true,
		"event.build.failed":        true,
		"event.build.succeeded":     true,
		"event.git.status":          true,
		"event.media.state":         true,
		"event.presence.changed":    true,
		"event.profile.switched":    true,
		"event.hotkey.triggered":    true,
		"event.process.started":     true,
		"event.window.moved":        true,
		"event.automation.started":  true,
		"event.automation.finished": true,
		"event.file.changed":        true,
		"event.ide.focus":           true,
		"cmd.media.toggle":          true,
		"cmd.media.next":            true,
		"cmd.media.volume":          true,
		"cmd.app.launch":            true,
		"cmd.hotkey.register":       true,
		"cmd.window.move":           true,
		"cmd.automation.run":        true,
		"cmd.metrics.set_interval":  true,
		"cmd.git.watch":             true,
		"error.protocol":            true,
	}
)

// Parse validiert eine eingehende Nachricht gegen die IPC-Regeln.
// Im Fehlerfall wird ein ProtocolError zurückgegeben; Close ist nur bei
// falschem protocol_version gesetzt (ARCHITECTURE.md 3.4).
func Parse(data []byte) (*Message, *ProtocolError) {
	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, &ProtocolError{Code: CodeInvalidRequest, Message: "Ungueltige JSON-RPC-2.0-Nachricht"}
	}
	if m.JSONRPC != "2.0" {
		return nil, &ProtocolError{Code: CodeInvalidRequest, Message: `jsonrpc muss "2.0" sein`}
	}
	if m.Method == "" {
		return nil, &ProtocolError{Code: CodeInvalidRequest, Message: "Feld method fehlt"}
	}
	if !methodCatalog[m.Method] {
		return nil, &ProtocolError{Code: CodeMethodNotFound, Message: fmt.Sprintf("Unbekannte Methode: %s", m.Method)}
	}
	if m.Params == nil {
		return nil, &ProtocolError{Code: CodeInvalidParams, Message: "Feld params fehlt"}
	}
	source, ok := m.Params["source"].(string)
	if !ok || !sourcePattern.MatchString(source) {
		return nil, &ProtocolError{Code: CodeInvalidParams, Message: "source muss S-A bis S-E oder bus sein"}
	}
	pv, ok := m.Params["protocol_version"].(float64)
	if !ok || int(pv) != ProtocolVersion {
		return nil, &ProtocolError{
			Code:    CodeInvalidParams,
			Message: fmt.Sprintf("protocol_version muss %d sein", ProtocolVersion),
			Close:   true,
		}
	}
	return &m, nil
}
