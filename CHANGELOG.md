# Changelog

<!-- Zweck: Übersicht aller nennenswerten Änderungen, Format nach Keep a Changelog. -->
<!-- Zuletzt geändert: 2026-08-17 -->

Alle nennenswerten Änderungen an diesem Projekt werden hier dokumentiert.

Das Format orientiert sich an [Keep a Changelog](https://keepachangelog.com/de/1.1.0/),
die Versionierung an [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- S-C Profile-System: `profiles.json` mit Context-Profilen (dev/gaming/afk), Hotkey-Bindings, Watcher-Overrides, Media-Einstellungen pro Profil; `ProfileManager` fuer Profil-Switching mit Validierung.
- S-C Event-Regeln: `event_rules.json` fuer Cross-Module-Automatisierung; `EventRule` mit Wildcard-Matching (`event.build.*`), IF-Bedingungen und cmd-Aktionen (z.B. `event.build.failed` -> `cmd.media.toggle` -> S-D).
- S-D Event-Reactions: `reactions.ts` reagiert auf `event.build.failed`, `event.build.succeeded`, `event.profile.switched` und passt Discord-Activity und Spotify-Lautstaerke an.
- S-B Profil-Switching: `cmd.profile.switch` zurueckt alle Hotkeys via `clear_all()` fuer Neuregistrierung.
- IPC-Events: `event.automation.rule.triggered` (Regel-Trigger), `cmd.media.set_activity` (Discord-Activity setzen).
- Schema: 38 Methoden (2 neue), `events.schema.gen.go` regeneriert; Schema-Validierung: 28 valid + 15 invalid Payloads.
- S-C Tests: +17 neue Tests (profiles_test.go: 8, eventrules_test.go: 6, reactions.test.mjs: 4).
- S-B Tests: +1 Test (clear_all).
- S-D Tests: reactions.test.mjs mit Mock-Context fuer Event-Reactions.
- S-B Macro Launchpad: Win32-API-Integration mit `windows-sys` — Global Hotkey Listener (`RegisterHotKey`/`UnregisterHotKey` mit Message-Loop), Process Launcher (`CreateProcessW` mit Fokussierung via `EnumWindows`), Window Manager (`SetWindowPos` für Titel-basierte Fenstersuche), Clipboard Manager (`AddClipboardFormatListener` für Monitor, `SetClipboardData`/`GetClipboardData`, History mit max. 50 Einträgen); empfängt `cmd.hotkey.register`/`cmd.app.launch`/`cmd.window.move`/`cmd.clipboard.set`, sendet `event.hotkey.triggered`/`event.process.started`/`event.window.moved`/`event.clipboard.changed` an den Bus.
- S-D Spotify-Service: OAuth2-Refresh-Flow (`spotify/auth.ts`), API-Client (`spotify/client.ts`, injizierbares `fetch`), `toMediaState`-Mapping und `SpotifySession` mit Auto-Refresh + 401-Retry; `index.ts` sendet `event.media.state` (Poll, nur bei Änderung) und verarbeitet `cmd.media.toggle`/`next`/`volume`; aktivierbar über `SPOTIFY_CLIENT_ID`/`SECRET`/`REFRESH_TOKEN`, sonst bleibt S-D Hello-Stub. Base-URLs per `SPOTIFY_API_BASE`/`SPOTIFY_TOKEN_URL` überschreibbar.
- Schema: `event.media.state` (playing/track/artist/album/album_art_url/duration_ms/progress_ms) und `cmd.media.volume` (`volume` 0–100) mit `params`-Pflichtfeldern; `events.schema.gen.go` regeneriert; Schema-Tests erweitert.
- S-D-Tests `tests/s-d-integrations/spotify.test.mjs` (node:test gegen lokalen Mock-Server, keine Credentials) — in CI-Job „Node/TS (S-D)" integriert (`npm test`).
- S-A Widget-Layout: konfigurierbares Dashboard-Grid (`Widgets/WidgetLayout.cs` mit JSON-Parsing/Validierung, Standard-Layout `layout.json`, `DashboardViewModel` mit `ObservableCollection<WidgetViewModel>` und `SetStatus`) — headless per xunit getestet.
- S-A Keepalive: `BusClient` sendet periodisch `event.system.heartbeat` (Schema-konform, Standard 5 s) und erkennt harte Netzabriffe ohne Close-Frame über einen Watchdog (Abort nach ausbleibendem Inbound-Frame, Standard 15 s); beides per Konstruktor `keepAliveInterval`/`keepAliveTimeout` konfigurierbar, in `MainWindow` aktiviert.
- Headless-C#-Tests: Heartbeat-Senden (`event.system.heartbeat` mit `source`/`service_id`/`ts`) und Watchdog-Reconnect gegen einen stillen Testserver (kein Close-Frame).
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
- README.md: Roadmap-Status auf Phase 2/3 aktualisiert, S-C Abschnitt mit IF-Bedingungen, S-D Abschnitt mit Discord/WhatsApp/VoIP.
- CONTRIBUTING.md: Aktueller Stand auf Phase 2/3 aktualisiert, CI-Pflichten mit Rust/S-D/Go/Schema-Test-Erfordernissen erweitert.
- ARCHITECTURE.md 3.3: Event-Katalog um 7 neue Events erweitert (29→36 Methoden), S-C/S-D Abschnitte aktualisiert.
- ARCHITECTURE.md/CONTRIBUTING.md: Zeitraumangaben (Wochen) aus der Roadmap entfernt; Änderungsdatum wird künftig als Kommentar im Dateikopf geführt (Konvention, siehe CONTRIBUTING.md).
- S-A: Stack-Entscheidung fixiert — **WinUI 3 (.NET 8)** (Build/Test nur auf Windows, Crostini nur Vorbereitung); Tauri-Option entfernt; RAM-Ziel (<150 MB) als weicher Richtwert umformuliert (README.md, ARCHITECTURE.md, docs/dev-env.md).

### Planned (Roadmap)
- Phase 1–4 gemäß README.md: IPC-Grundgerüst, Kernfeatures, Integration, Release v1.0.
