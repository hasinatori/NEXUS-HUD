# Changelog

<!-- Zweck: Übersicht aller nennenswerten Änderungen, Format nach Keep a Changelog. -->
<!-- Zuletzt geändert: 2026-08-14 -->

Alle nennenswerten Änderungen an diesem Projekt werden hier dokumentiert.

Das Format orientiert sich an [Keep a Changelog](https://keepachangelog.com/de/1.1.0/),
die Versionierung an [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- Schema: event-spezifische `params`-Pflichtfelder per `if/then` (u.a. `event.system.metrics` mit `cpu`/`ram`, `event.git.status`, `event.build.*` mit `project`/`ok`, `event.file.changed`, `event.automation.*`, `error.protocol`).
- Bus-Härtung: Der Bus validiert eingehende Nachrichten gegen das eingebettete Event-Schema (`shared/bus/events.schema.gen.go`) und beantwortet Verstöße mit `error.protocol` (`-32602`).
- Generator `scripts/generate-schema.py` erzeugt `shared/bus/events.schema.gen.go` aus dem Schema; CI prüft die Konsistenz der generierten Datei.
- `shared/wsclient`: wiederverwendbarer `Client` mit Auto-Reconnect und Notification-Send (Hello inkl. `source`/`protocol_version`/`ts`).
- S-E Monitor: System-Metriken (`event.system.metrics`, CPU/RAM via `/proc`), Git-Status (`event.git.status`, Branch/staged/uncommitted/ahead-behind) und Build-Log-Parser (`event.build.succeeded`/`event.build.failed`) mit Flags `-metrics-interval`, `-git-dir`, `-git-interval`, `-build-log`, `-build-project`.
- `scripts/prototype-hud`: Terminal-HUD als S-A-Stand-in (Live-Ansicht + `-test`-Modus mit `-expect`-Prüfung); läuft als E2E-Job in der CI gegen Bus + S-E.
- `shared/wsclient`: optionaler `OnMessage`-Callback für eingehende Bus-Events.
- S-C Automation Engine: File-Watcher (fsnotify, `event.file.changed`), Task-Runner (Kommando + Timeout, `event.automation.started`/`finished` mit Exit-Code), JSON-Konfiguration (`-config`, Trigger → Task) und `cmd.automation.run`-Handler.
- Schema: `cmd.automation.run` verlangt `task`; Prototype-HUD kann per `-cmd`/`-cmd-params` Kommandos senden; CI-E2E prüft nun auch S-C (Automation + File-Watch).
- Projektstruktur: Modul-Ordner S-A bis S-E (`services/`), gemeinsame Basis (`shared/`).
- Architektur-Spec (`ARCHITECTURE.md`) inkl. IPC-Protokoll (Named Pipes / WebSocket, JSON-RPC 2.0).
- IPC-Event-Schema (`schema/events.schema.json`).
- GitHub-CI: Markdown-Linting, Schema-Validierung, Struktur-Check.
- Issue-/PR-Templates, Dependabot-Konfiguration, Contribution- und Security-Richtlinie.
- IPC-Protokoll v1: Local-WebSocket als Phase-1-Transport, `protocol_version`, Handshake und Fehlerformat (`error.protocol`) spezifiziert.
- Schema-Test `tests/schema/validate_events.py` (Beispiel-Payloads gegen das Event-Schema, in CI integriert).
- Dev-Bus `shared/bus/` (Go): Local-WebSocket-Server auf `127.0.0.1:49152`, Eingangsvalidierung gegen die Schema-Regeln, Broadcast an alle Clients, `error.protocol` bei ungültigen Nachrichten, eigenes Start-Hello.
- Go-Client `shared/wsclient` für Services (verbindet mit Retry, sendet periodisch `event.system.hello`).
- Service-Stubs S-B (Rust), S-C/S-E (Go), S-D (Node/TS) — senden nach Connect `event.system.hello` (Wiederholung alle 5 s).
- Verifikation `scripts/hello-check` (Go): erwartet 5 `event.system.hello` (Bus + S-B bis S-E) in 15 s.
- CI-Jobs für Go (`vet`/`build`/`test`), Node/TS (`typecheck`) und Rust (`cargo build --locked`).
- Schema: `source` erlaubt zusätzlich `bus` (IPC-Bus als Absender).
- Versionierung: zentrale `VERSION.json` (Projektversion + Modulzähler), abgeleitete Modulversionen in `shared/version/version.go` (Go), `Cargo.toml` (S-B) und `package.json` (S-D); Versionen werden aus `CARGO_PKG_VERSION`/`version` gespeist statt hartkodiert.
- Release-Skripte `scripts/check-version.py` (Konsistenzprüfung), `scripts/bump-version.py` (Bump + CHANGELOG-Schnitt) und `scripts/release-notes.py` (Notizen extrahieren).
- CI-Job „Version-Konsistenz" (prüft VERSION.json gegen alle abgeleiteten Dateien).
- S-A `BusClient` mit Reconnect-Loop (`Connected`/`Disconnected`-Events, erneutes Hello je Verbindung, `StopAsync`); `MainWindow` auf die neue API umgestellt.
- Headless-C#-Tests `tests/s-a-ui-shell` (xunit, `net8.0`): testen `BusClient` (Connect/Hello, Auto-Reconnect nach Bus-Ausfall, Disconnected-Event, Send ohne Verbindung wirft), `Protocol` und `MainViewModel` — laufen auf Linux/Crostini und in CI (neuer Job „C# (S-A Bus + ViewModels headless)").
- `.github/workflows/release.yml`: bereitet per „Run workflow" einen Release-PR vor (Branch `release/vX.Y.Z`, Versions-Bump, CHANGELOG-Schnitt).
- `.github/workflows/tag.yml`: setzt nach Merge eines Release-PRs automatisch Tag `vX.Y.Z` und erstellt das GitHub Release.
- `.github/workflows/publish.yml`: veröffentlicht GitHub Releases für manuell gepushte Tags `vX.Y.Z` (Pre-Release bis v1.0.0).
- `docs/releasing.md`: Versionsschema, SemVer-Politik, Ablauf und Guardrails.

### Changed
- README.md übernimmt die Rolle der Hauptdokumentation (README.txt entfernt).
- README.md: Status auf Phase 1 aktualisiert, Abschnitt „Getting Started" mit konkreten Startbefehlen für den Dev-Bus und alle Services.
- ARCHITECTURE.md 3.3: Bus-Hello und periodisches Hello der Stubs dokumentiert.
- ARCHITECTURE.md/CONTRIBUTING.md: Status und CI-Pflichten an Phase 1 und das Release-Setup angepasst.
- README.md/ARCHITECTURE.md: Zeitraumangaben (Wochen) aus der Roadmap entfernt; Änderungsdatum wird künftig als Kommentar im Dateikopf geführt (Konvention, siehe CONTRIBUTING.md).
- S-A: Stack-Entscheidung fixiert — **WinUI 3 (.NET 8)** (Build/Test nur auf Windows, Crostini nur Vorbereitung); Tauri-Option entfernt; RAM-Ziel (<150 MB) als weicher Richtwert umformuliert (README.md, ARCHITECTURE.md, docs/dev-env.md).

### Planned (Roadmap)
- Phase 1–4 gemäß README.md: IPC-Grundgerüst, Kernfeatures, Integration, Release v1.0.
