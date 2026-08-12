# Dev-Umgebung — Setup & Workflows

<!-- Zweck: Konkrete Einrichtung der Entwicklungsumgebung für das NEXUS HUD. Planungsstand. -->

## Zielplattform
- Primär **Windows**: Overlay, Global Hotkeys, Win32 API (S-A, S-B).
- Entwicklung/Unit-Tests auch auf **ChromeOS Crostini (Debian)** möglich.

## Toolchains (je Modul, noch zu finalisieren)

| Modul | Toolchain |
| :--- | :--- |
| S-A / S-B | .NET 8 SDK (WinUI 3/WPF) oder Rust + Node.js 20+ (Tauri) |
| S-C / S-E | Go 1.22+ oder Python 3.11+ |
| S-D | Node.js 20+ |

## Git-Workflow
- Branch: `feature/*`, `fix/*`, `docs/*`.
- PR gegen `main`, CI muss grün sein (siehe CONTRIBUTING.md).

## Lokale Checks (vor Push)
1. `schema/events.schema.json` valid (JSON-Schema).
2. Markdown-Lint für README/Docs.
3. Struktur-Check (Modul-Ordner S-A bis S-E vorhanden).

## GitHub-Befehle (gh-CLI)
- `gh repo view` — Repo-Info.
- `gh issue list --milestone "Phase 1"` — offene Punkte je Phase.
- `gh workflow run ci.yml` — CI manuell anstoßen.
