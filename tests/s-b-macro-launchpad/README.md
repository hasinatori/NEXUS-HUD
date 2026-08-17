# tests/s-b-macro-launchpad

<!-- Zweck: Tests für das Macro- & Launchpad-System (S-B). -->

Tests sind als inline `#[cfg(test)]`-Module in den jeweiligen Source-Dateien implementiert:

- `bus.rs`: JSON-RPC-Nachrichten-Builder, Serialisierung, Struktur-Validierung.
- `hotkey.rs`: HotkeyManager (Register/Find/ID-Inkrement), parse_modifiers/parse_vkey (Windows-only).
- `process.rs`: launch_app Stub-Rückgabewerte (non-Windows).
- `window.rs`: Stub-Funktionen (find_window, get_window_rect, set_window_pos, etc.).
- `clipboard.rs`: ClipboardManager (has_changed, set_last_content), ClipboardWatcher.

Ausführung: `cargo test` im Service-Verzeichnis.
