# Rule-Engine — Automations-Syntax (S-C)

<!-- Zweck: Spec der IF-THIS-THEN-THAT-Regeln der Automation Engine (S-C). Planungsstand. -->

## Grundform
```
WHEN <Trigger>
IF <Bedingung>
THEN <Aktion>
```

## Trigger (Vorschlag)
- `file.created` / `file.changed` / `file.deleted` in überwachten Ordnern.
- Events des IPC-Bus, z. B. `event.build.failed`, `event.media.state`.
- Zeitpläne (cron-artig).

## Bedingungen
- Vergleichsoperatoren: `=`, `!=`, `>`, `<`, `contains`.
- Profil-Kontext: `profile is dev` o. ä. — Regeln reagieren auf `event.profile.switched`.

## Aktionen
- Skript ausführen, Datei konvertieren, Aufräum-Pipeline starten.
- Event an den Bus senden (z. B. Sound über S-D, UI-Flash über S-A).

## Beispiel
```
WHEN event.build.failed
IF profile is dev
THEN cmd.media.play_sound("error.wav") + cmd.ui.flash("Build")
```

## Hinweise
- Regeln leben in S-C, sie sind ausführbar ohne Cloud.
- Konkrete Syntax wird beim ersten Code der Rule-Engine finalisiert und hier dokumentiert.
