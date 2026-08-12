# tests/s-d-integrations

<!-- Zweck: Tests für die API-Integrationen (S-D): Spotify, Discord, WhatsApp. -->

Geplant:
- Spotify: OAuth2-Flow und Play/Pause/Track-Abruf gegen Mock-API.
- Discord: Rich-Presence-Updates und Trigger-Word-Logik (offizielle Bot-API).
- WhatsApp: Webhook-Verarbeitung und Event-Weiterleitung an das HUD.

Konvention: Alle Tests laufen gegen Mock-Server — **keine** echten API-Credentials im CI.
