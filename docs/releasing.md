# docs/ — Releasing

<!-- Zweck: Wie NEXUS HUD versioniert und veröffentlicht wird. -->

## Versionsschema

**Quell der Wahrheit** ist `VERSION.json` im Repo-Root. Sie enthält die
Projektversion (SemVer `MAJOR.MINOR.PATCH`) und einen Zähler je Modul.

```json
{
  "version": "0.2.0",
  "modules": { "bus": 1, "s-b": 1, "s-c": 1, "s-d": 1, "s-e": 1 }
}
```

Jedes Modul trägt eine abgeleitete Version `<projekt>-<modul>.<zähler>`:

| Modul          | Version             | Wo hinterlegt                                                        |
| -------------- | ------------------- | -------------------------------------------------------------------- |
| Bus (Go)       | `0.2.0-bus.1`       | `shared/version/version.go` (Konstante `Bus`)                        |
| S-B (Rust)     | `0.2.0-s-b.1`       | `services/s-b-macro-launchpad/Cargo.toml`                            |
| S-C (Go)       | `0.2.0-s-c.1`       | `shared/version/version.go` (Konstante `SCAutomation`)               |
| S-D (Node/TS)  | `0.2.0-s-d.1`       | `services/s-d-integrations/package.json`                             |
| S-E (Go)       | `0.2.0-s-e.1`       | `shared/version/version.go` (Konstante `SEMonitor`)                  |

## Skripte

| Skript                            | Zweck                                                                  |
| --------------------------------- | ---------------------------------------------------------------------- |
| `scripts/check-version.py`        | Prüft die Konsistenz aller Versionen (CI-Job „Version-Konsistenz").    |
| `scripts/bump-version.py`         | Erhöht Projekt-/Modulversionen und schneidet `[Unreleased]` aus.       |
| `scripts/release-notes.py`        | Extrahiert die Release-Notizen einer Version aus der `CHANGELOG.md`.   |

## SemVer-Politik

- `MAJOR.MINOR.PATCH` gemäß [Semantic Versioning](https://semver.org/).
- Bis v1.0.0 (`MAJOR = 0`): Releases werden als **Pre-Release** markiert.
- **Protokoll-Regel:** Jede Änderung am IPC-Protokoll
  (`schema/events.schema.json` inkompatibel) erzwingt eine Erhöhung von
  `protocol_version` **und** einen Major-Bump. Bis v1.0.0 gilt die
  Empfehlung, Protokolländerungen zu bündeln (siehe `ARCHITECTURE.md`).
- v1.0.0 erscheint nach Abschluss von Phase 4 (siehe `README.md`).

## Ablauf (automatisch, empfohlen)

1. **Vorbereiten:** Im Repo auf GitHub unter
   **Actions → „release" → „Run workflow"** den Bump-Typ wählen
   (`patch`/`minor`/`major`).
2. Der Workflow:
   - ermittelt die nächste Projektversion (`bump-version.py --next`),
   - erkennt geänderte Module seit dem letzten Tag (Pfadzuordnung unten),
   - erstellt Branch `release/vX.Y.Z`, bumpt Versionen und CHANGELOG,
   - pusht den Branch und erstellt einen **Release-PR** gegen `main`.
3. **Review & Merge:** Der PR läuft durch die CI
   (inkl. „Version-Konsistenz"). Nach dem Merge setzt der Workflow
   **„tag"** den Tag `vX.Y.Z` (aus `VERSION.json`) und erstellt das
   **GitHub Release** direkt (GITHUB_TOKEN-Tags triggern „publish" nicht,
   siehe `.github/workflows/tag.yml`).

## Pfadzuordnung (geänderte Module)

| Modul  | Pfade                                                                    |
| ------ | ------------------------------------------------------------------------ |
| `bus`  | alles außerhalb `services/` (Bus, `shared/`, `cmd/`, Docs, CI, Schema …) |
| `s-b`  | `services/s-b-macro-launchpad/`                                          |
| `s-c`  | `services/s-c-automation/`                                               |
| `s-d`  | `services/s-d-integrations/`                                             |
| `s-e`  | `services/s-e-monitor/`                                                  |

## Manueller Fallback

Falls der Workflow nicht verfügbar ist:

```bash
git checkout -b release/vX.Y.Z
python3 scripts/bump-version.py minor bus s-b   # Typ + Module
git add -A && git commit -m "chore: Release vX.Y.Z"
git push -u origin release/vX.Y.Z               # PR gegen main erstellen
```

Danach wie oben: Review, Merge, Tag:

```bash
git checkout main && git pull
git tag vX.Y.Z && git push origin vX.Y.Z         # triggert "publish"
```

## Guardrails

- `bump-version.py` verweigert das Bumpen auf `main`.
- Leere `[Unreleased]`-Sektion blockiert den Bump.
- `check-version.py` läuft in der CI und blockiert inkonsistente Versionen.
- `CHANGELOG.md` hält sich an Keep a Changelog und SemVer.
