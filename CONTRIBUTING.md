# CONTRIBUTING.md

<!-- Zweck: Regeln und Ablauf für Beiträge zum NEXUS HUD (Issues, PRs, Entwicklung). -->

## Projekt im Überblick
NEXUS HUD ist ein Desktop-HUD für den 2. Monitor mit entkoppelter Modul-Architektur
(S-A bis S-E). Grundlagen:
- **README.md** — Projektübersicht, Module, Roadmap.
- **ARCHITECTURE.md** — Technische Spec: IPC, Event-Schema, Security, RAM-Budget.

Aktueller Stand: **Phase 1** — IPC-Protokoll v1, Dev-Bus und Service-Stubs
funktionsfähig; Versionierung und Release-Setup eingerichtet (siehe `docs/releasing.md`).

## Entwicklungsumgebung
- Zielplattform: Windows. Entwicklung/Testing auch auf Linux (Crostini) möglich.
- Je Modul wird eine Toolchain benötigt — Details im "Getting Started"-Abschnitt der README.md.
- Tech-Stacks: S-B = Rust, S-C = Go, S-D = Node/TS, S-E = Go, S-A = offen (Tauri oder WinUI 3).

## Arbeitsweise
1. **Issue anlegen oder verlinken** für jede Änderung (Bug, Feature, Doku).
2. **Branch** benennen: `feature/<kurzbeschreibung>`, `fix/<kurzbeschreibung>` oder `docs/<kurzbeschreibung>`.
3. Commit-Konvention: prägnante deutsche Betreffzeile, klare Beschreibung.
4. **PR gegen `main`** mit Bezug auf das Issue; CI muss grün sein.

## CI-Pflichten
- `schema/events.schema.json` darf nicht ungültig werden (Workflow `check-schema`).
- Markdown-Dateien müssen lint-frei sein (Workflow `markdownlint`).
- Die Ordnerstruktur der Module (S-A bis S-E) bleibt erhalten (Workflow `structure-check`).
- **Versionen müssen konsistent sein** (Workflow `version-check`): `VERSION.json`,
  `shared/version/version.go`, `Cargo.toml` (S-B) und `package.json` (S-D) immer
  gemeinsam ändern; bei neuen Features die `[Unreleased]`-Sektion der `CHANGELOG.md`
  ergänzen. Den Bump übernimmt `scripts/bump-version.py` (siehe `docs/releasing.md`).

## Konventionen
- Doku und GUI-Texte auf **Deutsch**; Provider-/Produktnamen bleiben original.
- Keine Code-Kommentare außer auf ausdrücklichen Wunsch.
- JSON-Schema in `schema/events.schema.json` ist die single source of truth für IPC-Verträge.
- Sicherheitsrelevantes bitte über SECURITY.md melden, nicht als öffentliches Issue.
