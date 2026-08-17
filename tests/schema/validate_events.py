#!/usr/bin/env python3
# Zuletzt geändert: 2026-08-14
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
        "method": "event.system.hello",
        "params": {
            "source": "bus",
            "protocol_version": 1,
            "service_id": "bus",
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
            "gpu_temp_c": 61,
            "ts": "2026-08-14T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.build.failed",
        "params": {
            "source": "S-E",
            "protocol_version": 1,
            "project": "NEXUS-HUD",
            "ok": False,
            "task": "test",
            "output": "Test fehlgeschlagen",
            "ts": "2026-08-14T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.build.succeeded",
        "params": {
            "source": "S-E",
            "protocol_version": 1,
            "project": "NEXUS-HUD",
            "ok": True,
            "duration_ms": 1250,
            "ts": "2026-08-14T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.git.status",
        "params": {
            "source": "S-E",
            "protocol_version": 1,
            "repo_path": "/home/sam/NEXUS",
            "branch": "main",
            "staged": 0,
            "uncommitted": 2,
            "ahead": 1,
            "behind": 0,
            "ts": "2026-08-14T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.file.changed",
        "params": {
            "source": "S-C",
            "protocol_version": 1,
            "path": "/tmp/beobachtet/datei.txt",
            "change": "write",
            "ts": "2026-08-14T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.automation.started",
        "params": {
            "source": "S-C",
            "protocol_version": 1,
            "id": "run-1",
            "name": "aufraeumen",
            "ts": "2026-08-14T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.automation.finished",
        "params": {
            "source": "S-C",
            "protocol_version": 1,
            "id": "run-1",
            "name": "aufraeumen",
            "exit_code": 0,
            "ts": "2026-08-14T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.media.state",
        "params": {
            "source": "S-D",
            "protocol_version": 1,
            "playing": True,
            "track": "Neon Sun",
            "artist": "System Overdrive",
            "album": "Night Drive",
            "album_art_url": "https://example.com/cover.jpg",
            "duration_ms": 224000,
            "progress_ms": 42000,
            "ts": "2026-08-16T10:00:00Z",
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
        "id": "cmd-2",
        "method": "cmd.media.volume",
        "params": {"source": "S-A", "protocol_version": 1, "volume": 40},
    },
    {
        "jsonrpc": "2.0",
        "id": "cmd-2",
        "method": "cmd.automation.run",
        "params": {"source": "S-A", "protocol_version": 1, "task": "backup"},
    },
    {
        "jsonrpc": "2.0",
        "method": "error.protocol",
        "params": {
            "source": "S-A",
            "protocol_version": 1,
            "code": -32602,
            "message": "Ungueltiges Event",
            "ts": "2026-08-14T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.presence.changed",
        "params": {
            "source": "S-D",
            "protocol_version": 1,
            "service": "discord",
            "status": "online",
            "activity": "Visual Studio Code",
            "activity_type": "playing",
            "ts": "2026-08-17T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.profile.switched",
        "params": {
            "source": "S-B",
            "protocol_version": 1,
            "profile": "gaming",
            "ts": "2026-08-17T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.call.initiated",
        "params": {
            "source": "S-D",
            "protocol_version": 1,
            "call_id": "CA123",
            "to": "+15559876543",
            "from": "+15551234567",
            "status": "queued",
            "ts": "2026-08-17T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.call.ended",
        "params": {
            "source": "S-D",
            "protocol_version": 1,
            "call_id": "CA123",
            "status": "completed",
            "duration_sec": 120,
            "ts": "2026-08-17T10:02:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.call.status",
        "params": {
            "source": "S-D",
            "protocol_version": 1,
            "call_id": "CA123",
            "to": "+15559876543",
            "from": "+15551234567",
            "status": "in-progress",
            "duration_sec": 30,
            "ts": "2026-08-17T10:00:30Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "id": "cmd-3",
        "method": "cmd.call.make",
        "params": {
            "source": "S-A",
            "protocol_version": 1,
            "to": "+15559876543",
        },
    },
    {
        "jsonrpc": "2.0",
        "id": "cmd-4",
        "method": "cmd.call.hangup",
        "params": {
            "source": "S-A",
            "protocol_version": 1,
            "call_id": "CA123",
        },
    },
    {
        "jsonrpc": "2.0",
        "id": "cmd-5",
        "method": "cmd.call.status",
        "params": {
            "source": "S-A",
            "protocol_version": 1,
            "call_id": "CA123",
        },
    },
    {
        "jsonrpc": "2.0",
        "id": "cmd-6",
        "method": "cmd.profile.switch",
        "params": {
            "source": "S-A",
            "protocol_version": 1,
            "profile": "dev",
        },
    },
    {
        "jsonrpc": "2.0",
        "id": "cmd-7",
        "method": "cmd.window.focus",
        "params": {
            "source": "S-A",
            "protocol_version": 1,
            "window_title": "VS Code",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.automation.rule.triggered",
        "params": {
            "source": "S-C",
            "protocol_version": 1,
            "rule_name": "build-failed-alert",
            "event_method": "event.build.failed",
            "ts": "2026-08-17T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "cmd.media.set_activity",
        "params": {
            "source": "S-C",
            "protocol_version": 1,
            "activity_type": 0,
            "name": "NEXUS HUD — Coding",
            "details": "Profil: dev",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "cmd.media.set_activity",
        "params": {
            "source": "S-C",
            "protocol_version": 1,
            "activity_type": 2,
            "name": "NEXUS HUD — Gaming",
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
    {
        "jsonrpc": "2.0",
        "method": "event.system.hello",
        "params": {"source": "S-A", "protocol_version": 1},  # service_id/version/ts fehlen
    },
    {
        "jsonrpc": "2.0",
        "method": "event.system.metrics",
        "params": {"source": "S-E", "protocol_version": 1, "ts": "2026-08-14T10:00:00Z"},  # cpu/ram fehlen
    },
    {
        "jsonrpc": "2.0",
        "method": "event.system.metrics",
        "params": {
            "source": "S-E",
            "protocol_version": 1,
            "cpu": 10,
            "ram": {"used_mb": 100},  # total_mb fehlt
            "ts": "2026-08-14T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.git.status",
        "params": {"source": "S-E", "protocol_version": 1, "repo_path": "/tmp/repo"},  # branch fehlt
    },
    {
        "jsonrpc": "2.0",
        "method": "event.build.failed",
        "params": {"source": "S-E", "protocol_version": 1, "project": "X", "ts": "2026-08-14T10:00:00Z"},  # ok fehlt
    },
    {
        "jsonrpc": "2.0",
        "method": "event.file.changed",
        "params": {
            "source": "S-C",
            "protocol_version": 1,
            "path": "/tmp/f.txt",
            "change": "unbekannt",  # kein erlaubter change-Wert
            "ts": "2026-08-14T10:00:00Z",
        },
    },
    {
        "jsonrpc": "2.0",
        "method": "event.automation.finished",
        "params": {"source": "S-C", "protocol_version": 1, "id": "r1", "name": "t"},  # exit_code fehlt
    },
    {
        "jsonrpc": "2.0",
        "method": "cmd.automation.run",  # task fehlt
        "params": {"source": "S-A", "protocol_version": 1},
    },
    {
        "jsonrpc": "2.0",
        "method": "error.protocol",
        "params": {"source": "S-A", "protocol_version": 1, "message": "x"},  # code/ts fehlen
    },
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
