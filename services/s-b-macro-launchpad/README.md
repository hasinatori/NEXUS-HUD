# S-B — Macro- & Launchpad-System

<!-- Zweck: Tiefe OS-Integration — globale Hotkeys, Prozesssteuerung, Fenster-Management
     und Clipboard-Verwaltung. Läuft als OS-Service auch im Vollbild/Spiel. -->

**Stack (laut README):** `C# / C++ / Rust`

Hauptaufgaben:
- [ ] Global Hotkey Listener (reagiert auch im Spiel/IDE auf Tastenkombinationen).
- [ ] Process Launcher (Programme/Spiele fokussieren, starten oder beenden).
- [ ] Window Manager (Fenster per Script auf bestimmte Monitore & Positionen schieben).
- [ ] Clipboard-Manager (Historie verwalten, Code-Snippets abgreifen).

IPC:
- Sendet: `event.hotkey.triggered`, `event.process.started`, `event.window.moved`.
- Empfängt: `cmd.hotkey.register`, `cmd.app.launch`, `cmd.window.move`.

Deliverable: Ein OS-Service, der Prozesse steuern und Shortcuts systemweit abfangen kann.
