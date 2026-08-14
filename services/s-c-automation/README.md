# S-C — Automation Engine

<!-- Zweck: Datei-Watcher + Task-Runner + einfache IF-THIS-THEN-THAT-Regeln,
     die lokal ohne Cloud ausgeführt werden. -->
<!-- Zuletzt geändert: 2026-08-14 -->

**Stack:** `Go` (fsnotify, `shared/wsclient`)

Die Automation Engine (S-C) führt lokale Automations aus, die über eine
JSON-Konfiguration definiert sind — eine pragmatische v1 der Rule-Engine
(Syntax-Doku: `docs/rule-engine.md`).

Hauptaufgaben:
- [x] **File-Watcher:** überwacht Ordner per fsnotify, meldet `event.file.changed`.
- [x] **Task-Runner:** führt Kommandos mit Zeitlimit aus, meldet
      `event.automation.started` / `event.automation.finished` (mit Exit-Code).
- [x] **Regeln:** `WHEN file.changed (Trigger) THEN Task` aus der Konfiguration.
- [ ] Bedingungen (`IF profile is dev`) und Trigger auf weitere Events — später.

## Konfiguration

Standardpfad: `services/s-c-automation/automations.json` (überschreibbar mit `-config`).

```json
{
  "tasks": {
    "echo-ok": { "command": ["sh", "-c", "echo OK"], "timeout_ms": 10000 }
  },
  "watchers": [
    { "path": "/tmp/nexus-inbox", "triggers": ["created", "changed"], "then": "echo-ok" }
  ]
}
```

- `tasks.<name>.command` — Kommandozeile (argv), `timeout_ms` optional (0 = kein Limit).
- `watchers[].path` — zu überwachender Ordner.
- `watchers[].triggers` — Schema-Enum `create|write|remove|rename|chmod`
  (bequeme Aliase: `created`/`changed` → `write`, `deleted` → `remove`); leer = alle.
- `watchers[].then` — auszuführender Task bei passendem Ereignis.

## Verhalten

- Nach dem Verbinden: `event.system.hello` (alle 5 s).
- `cmd.automation.run` (params: `task` = Taskname) führt den Task asynchron aus;
  unbekannte Tasks werden nur geloggt.
- Datei-Ereignis im überwachten Ordner → `event.file.changed` + Ausführung des `then`-Tasks.
- Jeder Lauf sendet `event.automation.started` (id, name) und danach
  `event.automation.finished` (id, name, exit_code); Timeout-Abbruch → Exit-Code != 0.

## Build & Run

```sh
# Bus starten (separates Terminal), dann:
go run ./services/s-c-automation -config ./services/s-c-automation/automations.json
```

Flags: `-port`, `-version`, `-config`.

## Tests

```sh
go test ./services/s-c-automation/...
```

- `runner`: started/finished-Events, Exit-Codes (Erfolg, Fehler, Timeout).
- `watcher`: echte fsnotify-Ereignisse (create), Klassifizierung, Trigger-Matching.
- `config`: JSON-Parsing, Validierung (fehlender `then`/`path`, unbekannter Task).

Sendet: `event.system.hello`, `event.file.changed`, `event.automation.started`,
`event.automation.finished`.
Empfängt: `cmd.automation.run`.

Deliverable: Ein autonomer Hintergrund-Dienst, der Custom-Automations ausführt.
