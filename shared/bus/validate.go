package bus

// Zuletzt geändert: 2026-08-29

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"
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
	// methodCatalog wird aus dem eingebetteten Event-Schema abgeleitet
	// (single source of truth, siehe schema/events.schema.json).
	methodCatalog = func() map[string]bool {
		var doc struct {
			Properties struct {
				Method struct {
					Enum []string `json:"enum"`
				} `json:"method"`
			} `json:"properties"`
		}
		if err := json.Unmarshal([]byte(eventsSchemaJSON), &doc); err != nil || len(doc.Properties.Method.Enum) == 0 {
			return map[string]bool{}
		}
		catalog := make(map[string]bool, len(doc.Properties.Method.Enum))
		for _, name := range doc.Properties.Method.Enum {
			catalog[name] = true
		}
		return catalog
	}()
)

var (
	schemaOnce sync.Once
	schema     *jsonschema.Schema
	schemaErr  error
)

// eventSchema lädt das eingebettete Event-Schema (single source of truth,
// siehe schema/events.schema.json) genau einmal und gibt es zurück.
func eventSchema() (*jsonschema.Schema, error) {
	schemaOnce.Do(func() {
		schema, schemaErr = jsonschema.CompileString("nexus://events.schema.json", eventsSchemaJSON)
	})
	return schema, schemaErr
}

// ValidateParams prüft eine eingehende Nachricht zusätzlich gegen das
// Event-Schema (event-spezifische params-Anforderungen, ARCHITECTURE.md 3.3).
func ValidateParams(data []byte) *ProtocolError {
	sch, err := eventSchema()
	if err != nil {
		return &ProtocolError{Code: CodeInvalidParams, Message: "Eingebettetes Event-Schema nicht ladbar"}
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return &ProtocolError{Code: CodeInvalidRequest, Message: "Ungueltige JSON-RPC-2.0-Nachricht"}
	}
	if err := sch.Validate(v); err != nil {
		return &ProtocolError{Code: CodeInvalidParams, Message: fmt.Sprintf("Event entspricht nicht dem Schema: %v", err)}
	}
	return nil
}

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
