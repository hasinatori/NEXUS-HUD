# tests/ — Test-Scaffolding

<!-- Zweck: Ort für alle Tests des NEXUS HUD, je Modul getrennt. -->
<!-- Zuletzt geändert: 2026-08-14 -->

Aufteilung:
- `s-a-ui-shell/` — Headless-C#-Tests (xunit) für BusClient (Reconnect), Protocol und MainViewModel.
- `s-b-macro-launchpad/` — Tests für Hotkeys, Launcher, Window-/Clipboard-Manager.
- `s-c-automation/` — Tests für File-Watcher, Task-Runner, Regel-Engine.
- `s-d-integrations/` — Tests für Spotify/Discord/WhatsApp-Services (mit Mock-Servern).
- `s-e-monitor/` — Tests für Git-Watcher, Build-Log-Parser, Metriken.

Konventionen:
- Stack-neutral: Der konkrete Test-Framework je Modul wird mit dem Tech-Stack gewählt.
- Headless-Tests sind Pflicht (z. B. Offscreen-Rendering für die UI).
- CI-Workflow ruft die Tests je Modul auf, sobald Module Code haben.
