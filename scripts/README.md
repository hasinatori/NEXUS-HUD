# scripts/ — Dev-, Test- & Build-Skripte

<!-- Zweck: Skripte für Entwicklung, Tests und Packaging des NEXUS HUD. -->
<!-- Zuletzt geändert: 2026-08-29 -->

Vorhanden:
- `hello-check` (Go) — Phase-1-Verifikation: erwartet 5 `event.system.hello` (Bus + S-B bis S-E) in 15 s. Start: `go run ./scripts/hello-check` (Port/Timeout über `-port`/`-timeout`).
- `measure-ram.sh` — RAM-Budget (Phase 3): baut die Go-Binaries, startet Bus + S-C + S-E + S-D (Node) auf Test-Port und misst den RSS je Modul (Details: `docs/ram-budget.md`). In CI als Job „RAM-Budget" mit <150-MB-Grenze.

Geplant:
- `dev-up.ps1` / `dev-up.sh` — startet UI + alle Services mit dem IPC-Bus.
- `test-all` — Testlauf je Modul (Headless).
- `package.ps1` — Installer-Build (InnoSetup) für das v1.0-Release.

Konvention:
- Windows-Skripte als `.ps1` (primäre Zielplattform), `.sh` nur für Dev-Helfer auf Crostini.
- Skripte rufen Module nur über den IPC-Bus bzw. deren CLI auf — keine direkten Abhängigkeiten.
