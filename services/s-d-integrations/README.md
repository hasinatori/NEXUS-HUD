# S-D — Integrated Apps

<!-- Zweck: Unified-API-Module, das Events von Drittanbieter-Services (Spotify, Discord,
     WhatsApp) einsammelt und als einheitliche Events in das NEXUS HUD einspeist. -->

**Stack (laut README):** `Node.js / TypeScript`

Hauptaufgaben:
- [ ] Spotify Service: OAuth2, Play/Pause, Track-Name, Album-Art & Lautstärke.
- [ ] Discord Rich Presence & Bot Service: Status-Updates und Trigger-Words.
- [ ] WhatsApp Webhook / Web-Automation: Benachrichtigungen an das HUD weiterleiten.

IPC:
- Sendet: `event.media.state`, `event.presence.changed`.
- Empfängt: `cmd.media.toggle`, `cmd.media.next`, `cmd.media.volume`.

Risiken beachten (siehe `../../ARCHITECTURE.md` Abschnitt 9):
- Spotify-Lautstärke ist nicht über die Web API steuerbar.
- WhatsApp-Web-Automation und Discord-Features sind ToS-risky → Feature-Flags.

Deliverable: Ein Unified-API-Module, das Events von Drittanbieter-Services einspeist.
