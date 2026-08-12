# scripts/ — Dev-, Test- & Build-Skripte

<!-- Zweck: Skripte für Entwicklung, Tests und Packaging des NEXUS HUD. -->

Geplant:
- `dev-up.ps1` / `dev-up.sh` — startet UI + alle Services mit dem IPC-Bus.
- `hello-check` — Phase-1-Verifikation: prüft, dass alle Services `event.system.hello` senden.
- `test-all` — Testlauf je Modul (Headless).
- `package.ps1` — Installer-Build (InnoSetup bzw. Tauri-Bundler) für das v1.0-Release.

Konvention:
- Windows-Skripte als `.ps1` (primäre Zielplattform), `.sh` nur für Dev-Helfer auf Crostini.
- Skripte rufen Module nur über den IPC-Bus bzw. deren CLI auf — keine direkten Abhängigkeiten.
