<!-- Zuletzt geändert: 2026-08-28 -->

# NEXUS HUD — IDE-Bridge (VS Code)

Schreibt die aktuell aktive Datei als JSON in eine Focus-Datei. Der
S-E-Monitor (`services/s-e-monitor`, Flag `-ide-focus`) liest diese Datei und
sendet bei einer Änderung `event.ide.focus` an den Bus (Schema 0.3):

```json
{
  "project": "NEXUS-HUD",
  "filename": "main.go",
  "language": "go",
  "path": "/pfad/zu/main.go",
  "ts": "2026-08-28T10:00:00Z"
}
```

## Einrichtung

- Erweiterung installieren (F5 im `extensions/vscode-nexus`-Ordner bei
  aktiver Entwicklungsinstanz, oder via VSIX).
- Standard-Focus-Datei: `~/.nexus/ide-focus.json` — änderbar über die
  Einstellung `nexus.focusFile` (leer = Standard).

## S-E-Monitor starten (mit IDE-Status)

```sh
go run ./cmd/bus &
go run ./services/s-e-monitor -ide-focus ~/.nexus/ide-focus.json
```

Jede Dateiwechsel im Editor erzeugt ein `event.ide.focus`. Ist keine Datei
aktiv (Editor leer/fokussiert), wird die leere Datei nicht überschrieben.

## Befehle

- `NEXUS HUD: Aktiven Fokus anzeigen` — zeigt Projekt + Datei des aktuellen
  Fokus an.

## Entwicklung & Tests

Keine Laufzeit-Abhängigkeiten; Kernlogik liegt getrennt in `src/focus.js`
(ohne `vscode`-Import) und ist headless testbar:

```sh
npm test
```
