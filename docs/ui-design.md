# UI-Design — NEXUS HUD (S-A)

<!-- Zweck: Design-Spec für die UI-Shell (S-A): Look & Feel, Layout, Widgets. Planungsstand. -->

## Leitmotiv
Cyberpunk-Dark-Mode: dunkler Hintergrund, neon-akzentuierte Akzente, dezente Scanline-/Grid-Effekte.
Ziel: hohe Ablesbarkeit bei geringer CPU-Last, keine Ablenkung im Vollbild-Spiel.

## Farbpalette (Vorschlag)

| Rolle | Farbe | Hex |
| :--- | :--- | :--- |
| Hintergrund | Anthrazit | `#0d1117` |
| Flächen | Dunkelgrau | `#161b22` |
| Primär-Akzent | Cyan | `#00d4ff` |
| Erfolg | Grün | `#3fb950` |
| Fehler | Rot | `#f85149` |
| Warnung | Gelb | `#d29922` |

## Layout
- Frameless Overlay, am 2. Monitor andockbar (Snapping), konfigurierbares Grid (S-A).
- Widgets als modulare Komponenten: Gauges, Buttons, Status-Badges, Mini-Grafen.
- Kontext-Profile (Dev/Gaming/AFK) ändern Sichtbarkeit und Layout-Set.

## Komponenten
- **Gauge** — CPU/RAM/GPU in Echtzeit (Events `event.system.metrics`).
- **Build-Badge** — Status aus `event.build.failed`/`event.build.succeeded`.
- **Media-Card** — Track/Album-Art/Play-Pause aus `event.media.state`.
- **Git-Widget** — Branch + Uncommitted aus `event.git.status`.

## Barrierefreiheit (später)
- Kontrastverhältnisse prüfen, Fokus-Tastaturbedienung, reduzierte Animationen optional.
