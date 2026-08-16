// Zuletzt geändert: 2026-08-16
// Dünner Client für die Spotify Web API (nur was das HUD braucht):
// aktueller Track, Play/Pause, next/previous, Lautstärke.
// fetchImpl ist für Tests injizierbar.

import type { MediaState } from "./media.ts";
import { toMediaState } from "./media.ts";

export type FetchLike = typeof fetch;

const DEFAULT_API_BASE = "https://api.spotify.com/v1";

/** Basis-URL der Web API; per SPOTIFY_API_BASE überschreibbar (Tests/Staging). */
function apiBase(): string {
  return process.env.SPOTIFY_API_BASE ?? DEFAULT_API_BASE;
}

export class SpotifyApiError extends Error {}

export class SpotifyClient {
  private readonly fetchImpl: FetchLike;

  constructor(
    private token: string,
    fetchImpl: FetchLike = fetch,
  ) {
    this.fetchImpl = fetchImpl;
  }

  /** Aktuell gespielter Track bzw. null, wenn nichts abgespielt wird. */
  async getCurrentlyPlaying(): Promise<MediaState | null> {
    const res = await this.request("/me/player/currently-playing", { method: "GET" });
    if (res.status === 204) {
      return null;
    }
    const raw = (await res.json()) as Parameters<typeof toMediaState>[0];
    return toMediaState(raw);
  }

  async setPlaying(playing: boolean): Promise<void> {
    await this.request(`/me/player/${playing ? "play" : "pause"}`, { method: "PUT" });
  }

  async next(): Promise<void> {
    await this.request("/me/player/next", { method: "POST" });
  }

  async previous(): Promise<void> {
    await this.request("/me/player/previous", { method: "POST" });
  }

  async setVolume(percent: number): Promise<void> {
    if (!Number.isInteger(percent) || percent < 0 || percent > 100) {
      throw new SpotifyApiError(`Lautstärke ungültig: ${percent} (0-100 erwartet).`);
    }
    await this.request(`/me/player/volume?volume_percent=${percent}`, { method: "PUT" });
  }

  private async request(path: string, init: RequestInit): Promise<Response> {
    let res: Response;
    try {
      res = await this.fetchImpl(`${apiBase()}${path}`, {
        ...init,
        headers: {
          ...init.headers,
          Authorization: `Bearer ${this.token}`,
        },
      });
    } catch (err) {
      throw new SpotifyApiError(`Spotify-API nicht erreichbar (${path}): ${(err as Error).message}`);
    }
    if (res.status === 401) {
      throw new SpotifyApiError(`Nicht autorisiert (${path}) — Token erneuern.`);
    }
    if (res.status === 404) {
      // Kein aktives Gerät/Playback: als "nichts spielt" behandeln.
      if (path === "/me/player/currently-playing") {
        return res;
      }
      throw new SpotifyApiError(`Spotify-API 404 (${path}) — kein aktives Gerät?`);
    }
    if (!res.ok) {
      throw new SpotifyApiError(`Spotify-API ${res.status} (${path}).`);
    }
    return res;
  }
}
