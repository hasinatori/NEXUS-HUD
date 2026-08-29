# CONTRIBUTING.md

<!-- Zweck: Regeln und Ablauf für Beiträge zum NEXUS HUD (Issues, PRs, Entwicklung). -->
<!-- Zuletzt geändert: 2026-08-29 -->

## Projekt im Überblick
NEXUS HUD ist ein Desktop-HUD für den 2. Monitor mit entkoppelter Modul-Architektur
(S-A bis S-E). Grundlagen:
- **README.md** — Projektübersicht, Module, Roadmap.
- **ARCHITECTURE.md** — Technische Spec: IPC, Event-Schema, Security, RAM-Budget.

Aktueller Stand: **Phase 2 abgeschlossen, Phase 3 teilweise umgesetzt** — IPC-Protokoll v1,
Dev-Bus, alle Service-Implementierungen, IF-Bedingungen, Profil-Switching, VoIP-Anrufe
und Tests für S-B/S-C/S-D stehen (39 IPC-Methoden). Versions-/Release-Prozess:
siehe `docs/releasing.md`.

## Entwicklungsumgebung
- Zielplattform: Windows. Entwicklung/Testing auch auf Linux (Crostini) möglich.
- Je Modul wird eine Toolchain benötigt — Details im "Getting Started"-Abschnitt der README.md.
- Tech-Stacks: S-A = WinUI 3 (.NET 8), S-B = Rust, S-C = Go, S-D = Node/TS, S-E = Go.

## Arbeitsweise
1. **Issue anlegen oder verlinken** für jede Änderung (Bug, Feature, Doku).
2. **Branch** benennen: `feature/<kurzbeschreibung>`, `fix/<kurzbeschreibung>` oder `docs/<kurzbeschreibung>`.
3. Commit-Konvention: prägnante deutsche Betreffzeile, klare Beschreibung.
4. **PR gegen `main`** mit Bezug auf das Issue; CI muss grün sein.

## CI-Pflichten
- `schema/events.schema.json` darf nicht ungültig werden (Workflow `check-schema`).
- Markdown-Dateien müssen lint-frei sein (Workflow `markdownlint`).
- Die Ordnerstruktur der Module (S-A bis S-E) bleibt erhalten (Workflow `structure-check`).
- **Rust:** `cargo test` muss grün sein (Workflow `rust`).
- **S-D (Node/TS):** `npm test` muss grün sein; neue Integrationen brauchen Mock-Server-Tests.
- **Go:** `go vet`, `go build`, `go test` müssen grün sein.
- **Schema:** `validate_events.py` muss alle Payloads validieren.
- **Versionen müssen konsistent sein** (Workflow `version-check`): `VERSION.json`,
  `shared/version/version.go`, `Cargo.toml` (S-B) und `package.json` (S-D) immer
  gemeinsam ändern; bei neuen Features die `[Unreleased]`-Sektion der `CHANGELOG.md`
  ergänzen. Den Bump übernimmt `scripts/bump-version.py` (siehe `docs/releasing.md`).

## Konventionen
- Doku und GUI-Texte auf **Deutsch**; Provider-/Produktnamen bleiben original.
- Keine Code-Kommentare außer auf ausdrücklichen Wunsch.
- **Änderungsdatum:** Bei jeder Änderung an einer Datei wird das aktuelle Datum
  als Kommentar im Dateikopf aktualisiert:
  - Markdown: `<!-- Zuletzt geändert: JJJJ-MM-TT -->`
  - Go/Rust/TypeScript/Python/Shell: `// Zuletzt geändert: JJJJ-MM-TT` bzw. `#`
  - **Ausnahmen:** JSON-Dateien (`VERSION.json`, `package.json`, `schema/*.json`,
    Lockfiles — Kommentare dort nicht möglich) und generierte Dateien
    (`shared/version/version.go` — wird vom Bump-Skript überschrieben).
- JSON-Schema in `schema/events.schema.json` ist die single source of truth für IPC-Verträge.
  Bei jeder Schema-Änderung die generierte Bus-Kopie neu erzeugen und committen:
  `python3 scripts/generate-schema.py` (erzeugt `shared/bus/events.schema.gen.go`).
- Sicherheitsrelevantes bitte über SECURITY.md melden, nicht als öffentliches Issue.
