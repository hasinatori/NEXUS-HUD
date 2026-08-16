# S-D — Integrated Apps

<!-- Zweck: Unified-API-Module, das Events von Drittanbieter-Services (Spotify, Discord,
     WhatsApp) einsammelt und als einheitliche Events in das NEXUS HUD einspeist. -->
<!-- Zuletzt geändert: 2026-08-16 -->

**Stack (laut README):** `Node.js / TypeScript`

Hauptaufgaben:
- [~] Spotify Service: OAuth2-Refresh-Flow, Play/Pause, next/previous, Lautstärke, Track-Name,
  Album-Art & Progress (`event.media.state`). Abhängig von Credentials in der Umgebung —
  ohne diese läuft S-D als reiner Hello-Stub weiter.
- [ ] Discord Rich Presence & Bot Service: Status-Updates und Trigger-Words.
- [ ] WhatsApp Webhook / Web-Automation: Benachrichtigungen an das HUD weiterleiten.

## Spotify-Service (Stand 2026-08-16)

Aktivierung über Umgebungsvariablen (keine Secrets committen, `.env` ist gitignored):

| Variable | Zweck |
| :--- | :--- |
| `SPOTIFY_CLIENT_ID` | OAuth2-Client-ID. |
| `SPOTIFY_CLIENT_SECRET` | OAuth2-Client-Secret. |
| `SPOTIFY_REFRESH_TOKEN` | Refresh-Token aus dem Authorization-Code-Flow (PKCE/Localhost-Redirect). |
| `SPOTIFY_POLL_INTERVAL_MS` | Poll-Intervall für `event.media.state` (Standard `10000`). |

Ohne Credentials bleibt S-D ein Hello-Stub (kein Fehler, nur Hinweis im Log).

Aufbau (`spotify/`):
- `config.ts` — Konfiguration aus der Umgebung (`loadSpotifyConfig`, null ohne Credentials).
- `auth.ts` — `refreshAccessToken`: Refresh-Token → Access-Token (Form-POST + Basic Auth).
- `client.ts` — `SpotifyClient` (injizierbares `fetch`): `getCurrentlyPlaying`, `setPlaying`,
  `next`, `previous`, `setVolume` gegen die Web API. Base-URL per `SPOTIFY_API_BASE`/
  `SPOTIFY_TOKEN_URL` überschreibbar (Tests/Staging).
- `media.ts` — `toMediaState`: Mapping der API-Antwort auf `event.media.state`-Params.
- `session.ts` — `SpotifySession`: hält das Access-Token, erneuert es bei Bedarf (Refresh-Flow)
  und wiederholt API-Aufrufe einmalig bei 401.

`index.ts` sendet bei Änderungen `event.media.state` (playing, track, artist, album,
album_art_url, duration_ms, progress_ms) und verarbeitet `cmd.media.toggle`, `cmd.media.next`
und `cmd.media.volume` (params `volume` 0–100).

Tests: `tests/s-d-integrations/spotify.test.mjs` (node:test gegen lokalen Mock-Server,
keine echten Credentials/kein Netz). Ausführen: `npm test` in diesem Ordner.

IPC:
- Sendet: `event.media.state`, `event.presence.changed`.
- Empfängt: `cmd.media.toggle`, `cmd.media.next`, `cmd.media.volume`.

Risiken beachten (siehe `../../ARCHITECTURE.md` Abschnitt 9):
- Spotify-Lautstärke ist nicht über die Web API steuerbar.
- WhatsApp-Web-Automation und Discord-Features sind ToS-risky → Feature-Flags.

Deliverable: Ein Unified-API-Module, das Events von Drittanbieter-Services einspeist.
