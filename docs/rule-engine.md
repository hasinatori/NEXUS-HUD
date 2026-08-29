# Rule-Engine — Automations-Syntax (S-C)

<!-- Zweck: Spec der IF-THIS-THEN-THAT-Regeln der Automation Engine (S-C).
     v1 ist als JSON-Konfiguration umgesetzt (services/s-c-automation). -->
<!-- Zuletzt geändert: 2026-08-29 -->

## Grundform

```text
WHEN <Trigger>
IF <Bedingung>
THEN <Aktion>
```

## Stand der Umsetzung (v1, JSON)

Die v1 lebt in der Datei `services/s-c-automation/automations.json`:

```json
{
  "tasks": {
    "backup": { "command": ["sh", "-c", "cp -r src bak"], "timeout_ms": 30000 }
  },
  "watchers": [
    { "path": "/home/sam/Downloads", "triggers": ["created"], "then": "backup" }
  ]
}
```

- `WHEN file.created/changed/deleted in <path>` → `watchers[]` mit `path` +
  `triggers` (`create`, `write`, `remove`, `rename`, `chmod`).
- `THEN <Aktion>` → `then` verweist auf einen `tasks`-Eintrag (Kommando + optionales
  Zeitlimit). Ausführung und Events: siehe `services/s-c-automation/README.md`.
- `IF <Bedingung>` → optionale `if`-Bedingung von `watchers[]` sowie der Event-Regeln
  (siehe unten).

## IF-Bedingungen (umgesetzt)

```jsonc
// Bedingungs-Objekte sind optional; alle Felder sind optional (leer = kein Filter).
{
  "profile": "dev",            // "dev" | "gaming" | "afk" | "*" (immer, nicht gesetzt)
  "event_field": "build.ok",   // Zustandswert eines vergangenen Events
  "event_value": "false",      // exakter Gleichheitsvergleich
  "max_runs": 2,               // Rate-Limiting: höchstens N Ausführungen ...
  "run_window_ms": 60000       //   ... innerhalb dieses Zeitfensters (0 = unbegrenzt)
}
```

- **Profil:** `profile` filtert nach dem aktuellen Kontext-Profil (gesetzt per
  `cmd.profile.switch` / `event.profile.switched`).
- **Event-Feld:** `event_field`/`event_value` prüfen den letzten bekannten Zustand —
  S-C merkt sich aus `event.build.failed`/`event.build.succeeded`
  `build.ok` (`"true"`/`"false"`) und `build.project`. Beispiel: nur bei Build-Fehler
  reagieren → `{ "event_field": "build.ok", "event_value": "false" }`.
- **Rate-Limiting:** `max_runs` + `run_window_ms` begrenzen Wiederholungen
  (0 = unbegrenzt).

## Event-Regeln (Cross-Module, umgesetzt)

Reagieren auf beliebige Bus-Events und stoßen ein `cmd` an einen Ziel-Service:

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

- `on_event` matcht exakt oder per Wildcard-Suffix, z. B. `event.build.*`.
- `if` = IF-Bedingung wie oben (optional).
- `action.cmd`/`target` senden einen Befehl an den Ziel-Service (`params` optional).
- Jede ausgelöste Regel sendet `event.automation.rule.triggered` (rule_name, event_method).

## Geplant (später)

### Trigger
- Zeitpläne (cron-artig).

### Bedingungen
- Weitere Vergleichsoperatoren: `!=`, `>`, `<`, `contains` (aktuell: exakter Gleichheitsvergleich).

## Beispiel (Langform)

```text
WHEN event.build.failed
IF profile is dev
THEN cmd.media.play_sound("error.wav") + cmd.ui.flash("Build")
```

## Hinweise
- Regeln leben in S-C, sie sind ausführbar ohne Cloud.
- Aktionen an andere Module laufen über `cmd`-Events an den Bus (z. B. Sound über
  S-D, UI-Flash über S-A) — Module bleiben entkoppelt.
