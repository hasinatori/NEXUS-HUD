# shared/ — Gemeinsame Basis

<!-- Zweck: Sprachneutraler, gemeinsam genutzter Code und Konstanten für alle Module (S-A bis S-E). -->

Geplante Inhalte:
- Protocol-Konstanten: Event-Methodennamen, Pipe-Name (`\\.\pipe\nexus-hud`), WS-Endpoint.
- Event-Datentypen / Typ-Definitionen für die IPC-Verträge.
- Validierungshilfen gegen `schema/events.schema.json`.

Wichtig:
- Die Module nutzen verschiedene Sprachen (C#, Rust, Go, Node/TS, Python).
- Daher Typen/Validator je Sprache **aus dem JSON-Schema generieren** bzw. damit abgleichen —
  `schema/events.schema.json` ist die single source of truth für IPC-Verträge.
- Keine Implementierung hier, die nur ein einziges Modul nutzt — das gehört ins jeweilige Modul.
