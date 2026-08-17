# tests/ — Test-Scaffolding

<!-- Zweck: Ort fuer alle Tests des NEXUS HUD, je Modul getrennt. -->
<!-- Zuletzt geaendert: 2026-08-17 -->

Aufteilung:
- `s-a-ui-shell/` — Headless-C#-Tests (xunit) fuer BusClient (Reconnect, Heartbeat, Watchdog),
  Protocol und ViewModels (inkl. Widget-Layout).
- `s-b-macro-launchpad/` — Inline-Rust-Tests in den Source-Dateien (bus.rs, hotkey.rs,
  process.rs, window.rs, clipboard.rs). Ausfuehrung via `cargo test`.
- `s-c-automation/` — Go-Tests fuer File-Watcher, Task-Runner, Regel-Engine und
  IF-Bedingungen (condition.go). Inline `_test.go` neben den Quelldateien.
- `s-d-integrations/` — Mock-Server-Tests fuer Spotify, Discord, WhatsApp und VoIP:
  OAuth2-Refresh-Flow, API-Calls, Media-Mapping, Presence-Mapping, Call-Mapping,
  401-Retry. Ausfuehrung via `npm test` (Node built-in test runner).
- `s-e-monitor/` — Go-Tests fuer Git-Watcher, Build-Log-Parser, Metriken.
- `schema/` — Python-Validierungsskript fuer IPC-Event-Payloads gegen JSON-Schema.

Konventionen:
- Stack-neutral: Der konkrete Test-Framework je Modul wird mit dem Tech-Stack gewaehlt.
- Headless-Tests sind Pflicht (z. B. Offscreen-Rendering fuer die UI).
- CI-Workflow ruft die Tests je Modul auf, sobald Module Code haben.
