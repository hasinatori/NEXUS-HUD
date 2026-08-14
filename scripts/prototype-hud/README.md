# Prototype-HUD (S-A-Stand-in)

<!-- Ein Terminal-HUD, das die Bus-Events live anzeigt (Vorschau für die spätere
     S-A-UI) und im Testmodus die End-to-End-Kette des Busses verifiziert. -->
<!-- Zuletzt geändert: 2026-08-14 -->

**Stack:** `Go` (konsumiert den Bus über `shared/wsclient`)

Der Prototype-HUD ist zweierlei:

1. **Live-Ansicht** (`-test` weggelassen): rendert Hellos, System-Metriken,
   Git-Status und Build-Ergebnis als Terminal-HUD — die früheste sichtbare
   Ausgabe des Systems, bevor S-A (WinUI 3) die UI übernimmt.
2. **End-to-End-Test** (`-test`): verbindet sich für ein Zeitfenster, sammelt
   alle eingehenden Events und prüft, dass die erwarteten `method[:source]`
   angekommen sind. Exit-Code 0 nur bei vollständiger Erfüllung. Diese Prüfung
   läuft im CI (`E2E`-Job) gegen Bus + S-E.

`source` ist `S-A` (Stand-in), `service_id` `prototype-hud` — das Hello ist
damit schema-konform und wird vom Bus akzeptiert.

## Build & Run

```sh
# Live-HUD (Bus muss laufen):
go run ./scripts/prototype-hud

# Testmodus mit Erwartungen:
go run ./scripts/prototype-hud -test -window 10s \
  -expect "event.system.hello:bus,event.system.metrics:S-E,event.git.status:S-E,event.build.failed:S-E"
```

Flags: `-port`, `-test`, `-window` (Testfenster, Standard 10 s), `-expect`
(Erwartete Events, Komma-getrennt, Format `method[:source]`).

## Tests

```sh
go test ./scripts/prototype-hud/...
```

- `parseExpects`/`checkExpects`: Parsen und Abgleich der Erwartungen.
- `hudState.apply`: Event-Parsing in die Anzeige-Struktur.

Sendet: `event.system.hello`.
Empfängt: alle Bus-Broadcasts (rendert relevante).
