// Zuletzt geändert: 2026-08-16
// SpotifySession: hält den Access-Token, erneuert ihn bei Bedarf über den
// Refresh-Token-Flow und reicht die API-Aufrufe an den SpotifyClient durch.
// Bei 401 wird einmalig ein frisches Token geholt und der Aufruf wiederholt.

import type { SpotifyConfig } from "./config.ts";
import type { FetchLike } from "./client.ts";
import { SpotifyApiError, SpotifyClient } from "./client.ts";
import { refreshAccessToken } from "./auth.ts";
import type { MediaState } from "./media.ts";

const REFRESH_MARGIN_MS = 60_000;

export class SpotifySession {
  private client: SpotifyClient;
  private expiresAtMs = 0;

  constructor(
    private readonly config: SpotifyConfig,
    private readonly fetchImpl: FetchLike = fetch,
  ) {
    this.client = new SpotifyClient("", fetchImpl);
  }

  /** Stellt sicher, dass ein gültiges Access-Token vorliegt (holt es bei Bedarf). */
  async ensureToken(): Promise<void> {
    if (Date.now() < this.expiresAtMs) {
      return;
    }
    const token = await refreshAccessToken(
      this.config.clientId,
      this.config.clientSecret,
      this.config.refreshToken,
      this.fetchImpl,
    );
    this.client = new SpotifyClient(token.access_token, this.fetchImpl);
    this.expiresAtMs = Date.now() + token.expires_in * 1000 - REFRESH_MARGIN_MS;
  }

  async getCurrentlyPlaying(): Promise<MediaState | null> {
    return this.withToken((c) => c.getCurrentlyPlaying());
  }

  async setPlaying(playing: boolean): Promise<void> {
    return this.withToken((c) => c.setPlaying(playing));
  }

  async next(): Promise<void> {
    return this.withToken((c) => c.next());
  }

  async previous(): Promise<void> {
    return this.withToken((c) => c.previous());
  }

  async setVolume(percent: number): Promise<void> {
    return this.withToken((c) => c.setVolume(percent));
  }

  /** Führt fn aus; bei 401 wird einmalig das Token erneuert und erneut versucht. */
  private async withToken<T>(fn: (client: SpotifyClient) => Promise<T>): Promise<T> {
    await this.ensureToken();
    try {
      return await fn(this.client);
    } catch (err) {
      if (err instanceof SpotifyApiError && err.message.startsWith("Nicht autorisiert")) {
        this.expiresAtMs = 0;
        await this.ensureToken();
        return fn(this.client);
      }
      throw err;
    }
  }
}
