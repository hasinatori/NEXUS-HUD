# SECURITY.md

<!-- Zweck: Sicherheitsrichtlinie für das NEXUS HUD und Meldeprozess für Schwachstellen. -->

## Unterstützte Versionen
Das Projekt befindet sich in der Planungsphase — es gibt noch kein stabiles Release.

| Version | Unterstützt |
| :--- | :--- |
| main (Entwicklung) | ja, mit Vorbehalt |
| Vor v1.0 | keine Sicherheitsgarantien |

## Sicherheitsprinzipien
- IPC ausschließlich lokal: Named Pipes bzw. WebSocket nur auf `127.0.0.1`.
- Origin-Check und Sitzungs-Token gegen Fremd-Clients (Details: ARCHITECTURE.md, Abschnitt 4).
- Keine Secrets im Klartext: OAuth-Tokens verschlüsselt ablegen (Windows DPAPI bzw. Keytar).
- CI validiert, dass keine `.env`-Dateien oder Credentials eingecheckt werden (siehe `.gitignore`).

## Meldung einer Schwachstelle
Melde Sicherheitsprobleme **nicht** als öffentliches Issue. Schreib eine E-Mail an den Maintainer
(Repository-Owner `hasinatori`) oder öffne einen privaten Security-Advisory über den
GitHub-Button "Report a vulnerability" im Repo.

Erwartete Antwortzeit: sobald wie möglich, in der Regel innerhalb von 14 Tagen.
