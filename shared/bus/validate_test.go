package bus

// Zuletzt geändert: 2026-08-29

import (
	"encoding/json"
	"testing"
)

// TestMethodCatalogMatchesSchema stellt sicher, dass jede im eingebetteten
// Event-Schema deklarierte Methode vom Bus akzeptiert wird. Verhindert, dass
// die Allowlist (Katalog) und das Schema auseinanderlaufen.
func TestMethodCatalogMatchesSchema(t *testing.T) {
	var doc struct {
		Properties struct {
			Method struct {
				Enum []string `json:"enum"`
			} `json:"method"`
		} `json:"properties"`
	}
	if err := json.Unmarshal([]byte(eventsSchemaJSON), &doc); err != nil {
		t.Fatalf("Schema parsen: %v", err)
	}
	if len(doc.Properties.Method.Enum) == 0 {
		t.Fatal("Schema ohne Methoden-Enum")
	}
	for _, name := range doc.Properties.Method.Enum {
		raw := `{"jsonrpc":"2.0","method":` + mustJSONString(name) +
			`,"params":{"source":"S-A","protocol_version":1}}`
		if _, verr := Parse([]byte(raw)); verr != nil {
			t.Errorf("Methode %q wird abgelehnt: %+v", name, verr)
		}
	}
}

func mustJSONString(s string) string {
	data, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(data)
}
