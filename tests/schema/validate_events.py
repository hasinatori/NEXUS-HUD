#!/usr/bin/env python3
# Zweck: Validiert Beispiel-Payloads gegen schema/events.schema.json.
# Lauf in CI (Job "Schema-Validierung") und lokal:  python3 tests/schema/validate_events.py
import json
import sys
from pathlib import Path

import jsonschema

SCHEMA = Path(__file__).resolve().parent.parent.parent / "schema" / "events.schema.json"

VALID = [
    {
        "jsonrpc": "2.0",
        "method": "event.system.hello",
        "params": {
            "source": "S-B",
            "protocol_version": 1,
            "service_id": "s-b-macro-launchpad",
            "version": "0.1.0",
            "ts": "2026-08-13T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.system.heartbeat",
        "params": {"source": "S-C", "protocol_version": 1},
    },
    {
        "jsonrpc": "2.0",
        "method": "event.system.metrics",
        "params": {
            "source": "S-E",
            "protocol_version": 1,
            "cpu": 34.2,
            "ram": {"used_mb": 6400, "total_mb": 16384},
        },
    },
    {
        "jsonrpc": "2.0",
        "id": "cmd-1",
        "method": "cmd.media.toggle",
        "params": {"source": "S-A", "protocol_version": 1},
    },
    {
        "jsonrpc": "2.0",
        "method": "error.protocol",
        "params": {
            "source": "S-A",
            "protocol_version": 1,
            "code": -32602,
            "message": "Ungueltiges Event",
            "ts": "2026-08-13T10:00:00Z",
        },
    },
]

INVALID = [
    {"jsonrpc": "2.0", "method": "event.system.hello"},  # params fehlt
    {
        "jsonrpc": "2.0",
        "method": "event.system.hello",
        "params": {"source": "S-B"},  # protocol_version fehlt
    },
    {
        "jsonrpc": "2.0",
        "method": "event.system.hello",
        "params": {"source": "S-F", "protocol_version": 1},  # unbekanntes Modul
    },
    {
        "jsonrpc": "2.0",
        "method": "event.unknown",  # unbekannte Methode
        "params": {"source": "S-A", "protocol_version": 1},
    },
    {
        "jsonrpc": "2.0",
        "method": "event.system.hello",
        "params": {"source": "S-A", "protocol_version": 2},  # falsche Version
    },
    {"jsonrpc": "1.0", "method": "event.system.hello", "params": {"source": "S-A"}},  # falsche jsonrpc-Version
]


def main() -> int:
    with open(SCHEMA, encoding="utf-8") as f:
        schema = json.load(f)

    failures = 0
    for i, payload in enumerate(VALID, 1):
        try:
            jsonschema.validate(payload, schema)
            print(f"VALID  #{i} ok")
        except jsonschema.ValidationError as e:
            failures += 1
            print(f"VALID  #{i} FEHLER: {e.message}")

    for i, payload in enumerate(INVALID, 1):
        try:
            jsonschema.validate(payload, schema)
            failures += 1
            print(f"INVALID #{i} FEHLER: wurde akzeptiert, sollte aber abgelehnt werden")
        except jsonschema.ValidationError:
            print(f"INVALID #{i} ok (abgelehnt)")

    if failures:
        print(f"\n{failures} Testfall/faelle fehlgeschlagen.")
        return 1
    print(f"\nAlle {len(VALID)} valid + {len(INVALID)} invalid Tests bestanden.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
