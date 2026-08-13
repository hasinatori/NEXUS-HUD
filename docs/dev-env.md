# Dev-Umgebung — Setup & Workflows

<!-- Zweck: Konkrete Einrichtung der Entwicklungsumgebung für das NEXUS HUD. Planungsstand. -->
<!-- Zuletzt geändert: 2026-08-13 -->

## Zielplattform
- Primär **Windows**: Overlay, Global Hotkeys, Win32 API (S-A, S-B).
- Entwicklung/Unit-Tests auch auf **ChromeOS Crostini (Debian)** möglich — S-A (WinUI 3) wird auf Crostini vorbereitet, aber erst auf der Windows-Maschine gebaut und getestet.

## Toolchains (je Modul, noch zu finalisieren)

| Modul | Toolchain |
| :--- | :--- |
| S-A | .NET 8 SDK + Windows App SDK (WinUI 3) — Build nur auf Windows |
| S-B | Rust |
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
