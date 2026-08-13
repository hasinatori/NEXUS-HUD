# scripts/ — Dev-, Test- & Build-Skripte

<!-- Zweck: Skripte für Entwicklung, Tests und Packaging des NEXUS HUD. -->

Vorhanden:
- `hello-check` (Go) — Phase-1-Verifikation: erwartet 5 `event.system.hello` (Bus + S-B bis S-E) in 15 s. Start: `go run ./scripts/hello-check` (Port/Timeout über `-port`/`-timeout`).

Geplant:
- `dev-up.ps1` / `dev-up.sh` — startet UI + alle Services mit dem IPC-Bus.
- `test-all` — Testlauf je Modul (Headless).
- `package.ps1` — Installer-Build (InnoSetup bzw. Tauri-Bundler) für das v1.0-Release.

Konvention:
- Windows-Skripte als `.ps1` (primäre Zielplattform), `.sh` nur für Dev-Helfer auf Crostini.
- Skripte rufen Module nur über den IPC-Bus bzw. deren CLI auf — keine direkten Abhängigkeiten.
