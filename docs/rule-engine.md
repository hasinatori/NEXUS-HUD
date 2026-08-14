# Rule-Engine — Automations-Syntax (S-C)

<!-- Zweck: Spec der IF-THIS-THEN-THAT-Regeln der Automation Engine (S-C).
     v1 ist als JSON-Konfiguration umgesetzt (services/s-c-automation). -->
<!-- Zuletzt geändert: 2026-08-14 -->

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
- `IF`-Bedingungen (Profil, Vergleichsoperatoren) sind noch nicht umgesetzt.

## Geplant (später)

### Trigger
- `file.created` / `file.changed` / `file.deleted` in überwachten Ordnern.
- Events des IPC-Bus, z. B. `event.build.failed`, `event.media.state`.
- Zeitpläne (cron-artig).

### Bedingungen
- Vergleichsoperatoren: `=`, `!=`, `>`, `<`, `contains`.
- Profil-Kontext: `profile is dev` o. ä. — Regeln reagieren auf `event.profile.switched`.

### Aktionen
- Skript ausführen, Datei konvertieren, Aufräum-Pipeline starten.
- Event an den Bus senden (z. B. Sound über S-D, UI-Flash über S-A).

## Beispiel (Langform)

```text
WHEN event.build.failed
IF profile is dev
THEN cmd.media.play_sound("error.wav") + cmd.ui.flash("Build")
```

## Hinweise
- Regeln leben in S-C, sie sind ausführbar ohne Cloud.
- Die konkrete Syntax wird mit der IF-Umsetzung hier finalisiert und dokumentiert.
