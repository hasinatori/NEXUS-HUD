# S-B — Macro- & Launchpad-System

<!-- Zweck: Tiefe OS-Integration — globale Hotkeys, Prozesssteuerung, Fenster-Management
     und Clipboard-Verwaltung. Läuft als OS-Service auch im Vollbild/Spiel. -->

**Stack:** `Rust` mit `windows-sys` für Win32 API

## Modulstruktur

| Datei | Verantwortung |
| :--- | :--- |
| `src/bus.rs` | IPC-Protokoll: Message-Typen, Hello/Heartbeat, Event-Builder |
| `src/hotkey.rs` | Global Hotkey Listener via `RegisterHotKey`/`UnregisterHotKey` |
| `src/process.rs` | Process Launcher via `CreateProcessW`, Fokussierung via `EnumWindows` |
| `src/window.rs` | Window Manager via `SetWindowPos`, `GetWindowRect` |
| `src/clipboard.rs` | Clipboard Manager via `AddClipboardFormatListener`, History |
| `src/main.rs` | Service-Loop: WS-Verbindung, Cmd-Handling, Event-Versand |

## Implementierte Features

- [x] **Global Hotkey Listener** — RegisterHotKey mit Message-Loop, Hotkey-IDs aus `cmd.hotkey.register`
- [x] **Process Launcher** — `CreateProcessW` mit optionaler Fokussierung, `event.process.started` an Bus
- [x] **Window Manager** — `SetWindowPos` für Titel-basierte Fenstersuche, `event.window.moved` an Bus
- [x] **Clipboard Manager** — `AddClipboardFormatListener` für Monitor, `SetClipboardData`/`GetClipboardData`, History (max. 50 Einträge), `event.clipboard.changed` an Bus
- [x] **IPC Integration** — Empfängt `cmd.hotkey.register`, `cmd.app.launch`, `cmd.window.move`, `cmd.clipboard.set`
- [x] **Event Sending** — Sendet `event.hotkey.triggered`, `event.process.started`, `event.window.moved`, `event.clipboard.changed`

## Noch offen

- [ ] Multi-Monitor-Awareness (Monitor-Auswahl für Fenster-Positionierung)

## IPC

- Sendet: `event.hotkey.triggered`, `event.process.started`, `event.window.moved`.
- Empfängt: `cmd.hotkey.register`, `cmd.app.launch`, `cmd.window.move`.

## Build & Run

```sh
# Build (Windows)
cargo build --release

# Starten
cargo run --release
```

Der Service verbindet sich mit dem IPC-Bus auf `ws://127.0.0.1:49152/` (Port konfigurierbar via `NEXUS_WS_PORT`).

## Hotkey-Format

Hotkeys werden als JSON registriert:

```json
{
  "source": "S-A",
  "protocol_version": 1,
  "hotkey_id": "dev-mode-toggle",
  "modifiers": ["CTRL", "SHIFT"],
  "key": "F1"
}
```

Unterstützte Modifier: `ALT`, `CTRL`/`CONTROL`, `SHIFT`, `WIN`/`META`
Unterstützte Keys: `F1`-`F12`, `A`-`Z`, `0`-`9`, `SPACE`, `ENTER`, `TAB`, `ESC`, `DELETE`, `INSERT`, `HOME`, `END`, `PAGEUP`, `PAGEDOWN`, Pfeiltasten

Deliverable: Ein OS-Service, der Prozesse steuern und Shortcuts systemweit abfangen kann.
