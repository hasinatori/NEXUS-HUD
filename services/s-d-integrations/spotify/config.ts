// Zuletzt geändert: 2026-08-16
// Konfiguration für den Spotify-Service aus Umgebungsvariablen.
// Keine Secrets committen: Credentials nur über die Umgebung (SPOTIFY_*).

export interface SpotifyConfig {
  clientId: string;
  clientSecret: string;
  refreshToken: string;
  pollIntervalMs: number;
}

const DEFAULT_POLL_INTERVAL_MS = 10_000;

/** Liest die Spotify-Konfiguration aus der Umgebung. Gibt null zurück, wenn
 *  keine Credentials gesetzt sind (Service bleibt dann reiner Stub). */
export function loadSpotifyConfig(env: Record<string, string | undefined> = process.env): SpotifyConfig | null {
  const clientId = env.SPOTIFY_CLIENT_ID?.trim() ?? "";
  const clientSecret = env.SPOTIFY_CLIENT_SECRET?.trim() ?? "";
  const refreshToken = env.SPOTIFY_REFRESH_TOKEN?.trim() ?? "";
  if (!clientId || !clientSecret || !refreshToken) {
    return null;
  }
  const poll = Number(env.SPOTIFY_POLL_INTERVAL_MS ?? DEFAULT_POLL_INTERVAL_MS);
  return {
    clientId,
    clientSecret,
    refreshToken,
    pollIntervalMs: Number.isFinite(poll) && poll > 0 ? poll : DEFAULT_POLL_INTERVAL_MS,
  };
}
