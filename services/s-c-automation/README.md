# S-C — Automation Engine

<!-- Zweck: Autonomer Hintergrund-Dienst für File-Watching, Task-Runner und eine
     IF-THIS-THEN-THAT-Regel-Engine. Alle Automations laufen lokal, ohne Cloud-Zwang. -->

**Stack (laut README):** `Go (Golang)` oder `Node.js / Python`

Hauptaufgaben:
- [ ] File-Watcher System (Downloads, Screenshots etc. überwachen).
- [ ] Task-Runner Engine (Dateien konvertieren, Skripte triggern, Aufräum-Pipelines).
- [ ] Node/Workflow Engine Parser (IF-THIS-THEN-THAT-Regeln).

IPC:
- Sendet: `event.automation.started`, `event.automation.finished`, `event.file.changed`.
- Empfängt: `cmd.automation.run`, `event.profile.switched` (aktive Regeln je Profil).

Deliverable: Ein autonomer Hintergrund-Dienst, der Custom-Automationen ausführt.
