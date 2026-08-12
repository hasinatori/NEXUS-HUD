# S-A — Frontend / UI Shell

<!-- Zweck: Das Hauptfenster (frameless Overlay auf dem 2. Monitor) und die komplette
     Widget-Oberfläche. Reine Darstellung — keine Geschäftslogik, keine OS-/API-Zugriffe. -->

**Stack (laut README):** `C# (.NET 8 / WinUI 3)` oder `TypeScript + React (Tauri)`

Hauptaufgaben:
- [ ] Hauptfenster (Frameless Overlay, Snapping für 2. Monitor).
- [ ] Dark-Mode / Cyberpunk UI Design & Grid Layout.
- [ ] Widget-Bibliothek (Gauges, Knöpfe, Status-Badges).
- [ ] Anbindung der lokalen WebSocket/Named-Pipe-Schnittstelle (Empfang von Events).

IPC:
- Sendet: `cmd.*` (Hotkeys, Launch, Automation, Media).
- Empfängt: alle `event.*`-Notifications.

Deliverable: Eine voll bedienbare UI, die Daten via JSON entgegennimmt und darstellt.
