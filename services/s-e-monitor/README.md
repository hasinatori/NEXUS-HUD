# S-E — Coding & Build Monitoring

<!-- Zweck: Developer-Service, der Arbeits- und Systemzustand in Echtzeit meldet —
     Git-Status, Build-Ergebnisse, System-Metriken und IDE-Fokus. -->

**Stack (laut README):** `Go` oder `Python`

Hauptaufgaben:
- [ ] Local Git Watcher: Branch, Uncommitted Changes & Push-Status.
- [ ] Build-Log Parser: überwacht Compiler-Ausgaben/Terminals, sendet Build-Fehler.
- [ ] System-Metrics Gatherer: CPU, RAM, GPU-Temp via OS APIs.
- [ ] Einfaches IDE-Plugin (z. B. VS Code) für Focus/Status-Übertragung.

IPC:
- Sendet: `event.system.metrics`, `event.git.status`, `event.build.failed`,
  `event.build.succeeded`, `event.ide.focus`.
- Empfängt: `cmd.metrics.set_interval`, `cmd.git.watch`.

Deliverable: Ein Developer-Service, der den Arbeits- und Systemzustand in Echtzeit meldet.
