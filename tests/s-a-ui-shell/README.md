# tests/s-a-ui-shell

<!-- Zweck: Tests für die UI-Shell (S-A): Widget-Rendering, Grid-Layout, Overlay-Verhalten. -->

Geplant:
- Frameless-Overlay und Snapping (2. Monitor) — Headless-Render-Tests.
- Grid-Konfiguration: Layout wird korrekt serialisiert/restauriert.
- Event-Anbindung: `event.*`-Notifications aktualisieren Widgets korrekt.

Konvention: Rendering-Tests laufen headless (offscreen), keine echte Desktop-Anzeige nötig.
