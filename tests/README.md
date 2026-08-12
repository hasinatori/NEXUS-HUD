# tests/ — Test-Scaffolding

<!-- Zweck: Ort für alle Tests des NEXUS HUD, je Modul getrennt. Aktuell Planungsstand. -->

Aufteilung:
- `s-a-ui-shell/` — Tests für die UI (Widget-Rendering, Layout, Event-Anbindung).
- `s-b-macro-launchpad/` — Tests für Hotkeys, Launcher, Window-/Clipboard-Manager.
- `s-c-automation/` — Tests für File-Watcher, Task-Runner, Regel-Engine.
- `s-d-integrations/` — Tests für Spotify/Discord/WhatsApp-Services (mit Mock-Servern).
- `s-e-monitor/` — Tests für Git-Watcher, Build-Log-Parser, Metriken.

Konventionen:
- Stack-neutral: Der konkrete Test-Framework je Modul wird mit dem Tech-Stack gewählt.
- Headless-Tests sind Pflicht (z. B. Offscreen-Rendering für die UI).
- CI-Workflow ruft die Tests je Modul auf, sobald Module Code haben.
