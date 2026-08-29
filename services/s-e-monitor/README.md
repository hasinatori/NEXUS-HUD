# S-E — Coding & Build Monitor

<!-- Zweck: Der Entwickler-Monitor: System-Metriken, Git-Status und Build-Ergebnisse
     werden als Events an den Bus gesendet. Keine UI — reine Datenerzeugung. -->
<!-- Zuletzt geändert: 2026-08-29 -->

**Stack:** `Go` (Build/Test auf Linux/Crostini und Windows; Metrik-Leser aktuell Linux via `/proc`)

Hauptaufgaben:
- [x] System-Metriken (CPU-Auslastung, RAM) als `event.system.metrics`.
- [x] Git-Status (Branch, staged/uncommitted, ahead/behind) als `event.git.status`.
- [x] Build-Log-Parser als `event.build.succeeded` / `event.build.failed`.
- [x] IDE-Status (`event.ide.focus`) via `-ide-focus` + IDE-Bridge (VS Code, siehe `extensions/vscode-nexus`).

## Verhalten

Nach dem Verbinden sendet S-E das `event.system.hello` (alle 5 s) und meldet dann je
nach Konfiguration:

| Event | Quelle | Intervall |
| :--- | :--- | :--- |
| `event.system.metrics` | CPU + RAM via `/proc` | `-metrics-interval` (Standard 5 s) |
| `event.git.status` | `git`-CLI im überwachten Repo | `-git-interval` (Standard 15 s) |
| `event.build.succeeded`/`failed` | Zustandswechsel im Log | Poll 1 s |
| `event.ide.focus` | Focus-Datei des IDE | Poll 1 s |

Die Metriken brauchen zwei Messpunkte (erster Wert `cpu=0` als Basis). Der Git-Status
degradiert bei fehlendem Upstream (ahead/behind = 0) und detached HEAD (Branch = kurzer
Hash). Der Build-Parser meldet nur **Zustandswechsel** (Erfolg↔Fehler), nicht jede Zeile;
Erkennungs-Muster: `build succeeded/ok/passed`, `success`, `erfolgreich` bzw. `error`,
`fail(ed|ure)`, `fehler`. Der IDE-Fokus kommt aus der Focus-Datei der IDE-Bridge
(`~/.nexus/ide-focus.json`, siehe `extensions/vscode-nexus`) und wird bei jedem
Dateiwechsel (Projekt, Datei, Sprache, Pfad) als `event.ide.focus` gemeldet.

Interaktive Commands: `cmd.metrics.set_interval` (params `interval_ms`, 0 = aus)
ändert das Metriken-Intervall zur Laufzeit; `cmd.git.watch` (params `path`) fügt dem
Monitoring weitere Git-Repos hinzu (jede Registrierung startet einen eigenen Poller).
Unbekannte Nachrichten werden ignoriert.

## Build & Run

```sh
# Bus starten (separates Terminal), dann:
go run ./services/s-e-monitor \
  -metrics-interval 5s \
  -git-dir /pfad/zum/repo \
  -git-interval 15s \
  -build-log /pfad/zum/build.log \
  -build-project mein-projekt \
  -ide-focus ~/.nexus/ide-focus.json
```

Flag-Übersicht (alle optional; ohne Flag wird der Bereich nicht gemeldet):

```text
-port                  Bus-Port (Standard 49152, NEXUS_WS_PORT)
-version               Version ausgeben
-metrics-interval      Intervall für Metriken (0 = aus, Standard 5s)
-git-dir               Zu überwachendes Git-Repo (leer = aus)
-git-interval          Intervall für Git-Status (Standard 15s)
-build-log             Zu überwachendes Build-Log (leer = aus)
-build-project         Projektname für Build-Events (Standard "build")
-ide-focus             IDE-Focus-Datei (leer = aus)
```

## Tests

```sh
go test ./services/s-e-monitor/...
```

- `metrics`: Parser für `/proc/stat` und `/proc/meminfo`, Delta-Logik (Basiswert).
- `gitstatus`: Zählt staged/uncommitted/untracked korrekt, detached HEAD (echtes Temp-Repo).
- `buildlog`: Zustandswechsel-Erkennung, fehlendes File, Zeilen-Klassifizierung.
- `idefocus`: Parse der Focus-Datei, Zustandswechsel, fehlendes/leeres/kaputtes File,
  auto-`ts` (kein Timestamp nötig).

Sendet: `event.system.hello`, `event.system.metrics`, `event.git.status`,
`event.build.succeeded`, `event.build.failed`, `event.ide.focus`.
Empfängt: `cmd.metrics.set_interval` (Intervall für Systemmetriken dynamisch, in
Millisekunden, `0` = aus), `cmd.git.watch` (überwacht zusätzlich ein Git-Repo per Pfad).

Deliverable: Ein Entwickler-Service, der den Arbeits- und Systemzustand in Echtzeit meldet.
