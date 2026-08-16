# tests/s-d-integrations

<!-- Zweck: Tests für die API-Integrationen (S-D): Spotify, Discord, WhatsApp. -->
<!-- Zuletzt geändert: 2026-08-16 -->

Abgedeckt (Mock-Server, keine echten API-Credentials im CI):
- Spotify: OAuth2-Refresh-Flow (`refreshAccessToken`), API-Aufrufe mit Bearer-Token
  (`getCurrentlyPlaying`, `setVolume`-Validierung), `toMediaState`-Mapping, `SpotifySession`
  (Token holen bei Bedarf + 401-Retry). Siehe `spotify.test.mjs`.

Ausführen (aus `services/s-d-integrations`, baut vorher `dist/`):

```bash
npm test
```

Geplant:
- Discord: Rich-Presence-Updates und Trigger-Word-Logik (offizielle Bot-API).
- WhatsApp: Webhook-Verarbeitung und Event-Weiterleitung an das HUD.

Konvention: Alle Tests laufen gegen Mock-Server — **keine** echten API-Credentials im CI.
