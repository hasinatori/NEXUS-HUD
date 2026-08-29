# S-C — Automation Engine

<!-- Zweck: Datei-Watcher + Task-Runner + einfache IF-THIS-THEN-THAT-Regeln,
     die lokal ohne Cloud ausgeführt werden. -->
<!-- Zuletzt geändert: 2026-08-29 -->

**Stack:** `Go` (fsnotify, `shared/wsclient`)

Die Automation Engine (S-C) führt lokale Automations aus, die über eine
JSON-Konfiguration definiert sind — mit Profil-System und Event-Regeln
(Syntax-Doku: `docs/rule-engine.md`).

Hauptaufgaben:
- [x] **File-Watcher:** überwacht Ordner per fsnotify, meldet `event.file.changed`.
- [x] **Task-Runner:** führt Kommandos mit Zeitlimit aus, meldet
      `event.automation.started` / `event.automation.finished` (mit Exit-Code).
- [x] **Regeln:** `WHEN Trigger IF Bedingung THEN Task` aus der Konfiguration.
- [x] **Profil-System:** Profile `dev`/`gaming`/`afk` mit Hotkey-/Watcher-/Media-
      Overrides; Switch per `cmd.profile.switch` oder `event.profile.switched`.
- [x] **Event-Regeln:** Cross-Module-Automatisierung auf beliebige Bus-Events
      (Wildcard z. B. `event.build.*`), Bedingungen (Profil, Event-Feld,
      Rate-Limiting), Aktion = `cmd` an Ziel-Service.

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
- `watchers[].if` — optionale IF-Bedingung (siehe `docs/rule-engine.md`), z. B.
  `{ "event_field": "build.ok", "event_value": "false" }`.

### Profile (`-profiles`, optional)

```json
{
  "profiles": [
    { "name": "dev" },
    {
      "name": "gaming",
      "hotkeys": [{ "modifiers": ["alt"], "key": "a", "action": "cmd.media.toggle" }],
      "media": { "volume": 60, "activity_type": 2, "status_text": "Gaming" },
      "watchers": [{ "path": "/tmp/nexus-inbox", "enabled": false }]
    }
  ]
}
```

Aktives Profil wird per `cmd.profile.switch` / `event.profile.switched` gesetzt
und fließt in die IF-Auswertung ein (`profile: "dev"`, `"*"` = immer).
`watchers[].enabled` kann Watcher pro Profil deaktivieren.
Ohne `-profiles` gelten die Standardprofile `dev`, `gaming`, `afk`.

### Event-Regeln (`-event-rules`, optional)

```json
{
  "rules": [
    {
      "name": "build-failed-sound",
      "on_event": "event.build.failed",
      "if": { "profile": "dev" },
      "action": { "cmd": "cmd.media.toggle", "target": "S-D", "params": {} }
    }
  ]
}
```

`on_event` matcht exakt oder per Wildcard (`event.build.*`); `action` sendet ein
Cmd an den Ziel-Service. Jede ausgelöste Regel erzeugt `event.automation.rule.triggered`.

## Verhalten

- Nach dem Verbinden: `event.system.hello` (alle 5 s).
- `cmd.automation.run` (params: `task` = Taskname) führt den Task asynchron aus;
  unbekannte Tasks werden nur geloggt.
- Datei-Ereignis im überwachten Ordner → `event.file.changed` + Ausführung des `then`-Tasks
  (IF-Bedingung wird vorher ausgewertet).
- Jeder Lauf sendet `event.automation.started` (id, name) und danach
  `event.automation.finished` (id, name, exit_code); Timeout-Abbruch → Exit-Code != 0.
- Bus-Event passt auf eine Event-Regel (exakt oder Wildcard) und die IF-Bedingung ist
  erfüllt → `event.automation.rule.triggered` + Aktion (`cmd` an Ziel-Service).

## Build & Run

```sh
# Bus starten (separates Terminal), dann:
go run ./services/s-c-automation -config ./services/s-c-automation/automations.json
```

Flags: `-port`, `-version`, `-config`, `-profiles`, `-event-rules`.

## Tests

```sh
go test ./services/s-c-automation/...
```

- `runner`: started/finished-Events, Exit-Codes (Erfolg, Fehler, Timeout).
- `watcher`: echte fsnotify-Ereignisse (create), Klassifizierung, Trigger-Matching.
- `config`: JSON-Parsing, Validierung (fehlender `then`/`path`, unbekannter Task).
- `profiles`: Profil-Parsing, Switch, Watcher-/Media-/Hotkey-Zugriff.
- `eventrules`/`condition`: Wildcard-Matching (`event.build.*`), IF-Auswertung
  (Profil, Event-Feld, Rate-Limiting), Validierung.

Sendet: `event.system.hello`, `event.file.changed`, `event.automation.started`,
`event.automation.finished`, `event.automation.rule.triggered`.
Empfängt: `cmd.automation.run`, `cmd.profile.switch`, `event.profile.switched`,
`event.build.failed`, `event.build.succeeded`.

Deliverable: Ein autonomer Hintergrund-Dienst, der Custom-Automations ausführt.
