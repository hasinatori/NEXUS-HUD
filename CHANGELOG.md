# Changelog

<!-- Zweck: Übersicht aller nennenswerten Änderungen, Format nach Keep a Changelog. -->

Alle nennenswerten Änderungen an diesem Projekt werden hier dokumentiert.

Das Format orientiert sich an [Keep a Changelog](https://keepachangelog.com/de/1.1.0/),
die Versionierung an [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- Projektstruktur: Modul-Ordner S-A bis S-E (`services/`), gemeinsame Basis (`shared/`).
- Architektur-Spec (`ARCHITECTURE.md`) inkl. IPC-Protokoll (Named Pipes / WebSocket, JSON-RPC 2.0).
- IPC-Event-Schema (`schema/events.schema.json`).
- GitHub-CI: Markdown-Linting, Schema-Validierung, Struktur-Check.
- Issue-/PR-Templates, Dependabot-Konfiguration, Contribution- und Security-Richtlinie.

### Changed
- README.md übernimmt die Rolle der Hauptdokumentation (README.txt entfernt).

### Planned (Roadmap)
- Phase 1–4 gemäß README.md: IPC-Grundgerüst, Kernfeatures, Integration, Release v1.0.
