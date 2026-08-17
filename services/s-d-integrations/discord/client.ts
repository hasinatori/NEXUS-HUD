// Zuletzt geaendert: 2026-08-17
// Duenner Client fuer die Discord Bot API (nur was das HUD braucht):
// User-Status abfragen, Rich Presence setzen.
// fetchImpl ist fuer Tests injizierbar.

import type { PresenceState } from "./presence.ts";
import { toPresenceState } from "./presence.ts";

export type FetchLike = typeof fetch;

const DEFAULT_API_BASE = "https://discord.com/api/v10";

function apiBase(): string {
  return process.env.DISCORD_API_BASE ?? DEFAULT_API_BASE;
}

export class DiscordApiError extends Error {}

export class DiscordClient {
  private readonly fetchImpl: FetchLike;

  constructor(
    private token: string,
    fetchImpl: FetchLike = fetch,
  ) {
    this.fetchImpl = fetchImpl;
  }

  /** Holt den aktuellen User-Status eines Users per Guild-Member-Endpoint. */
  async getPresence(userId: string, guildId: string): Promise<PresenceState | null> {
    const res = await this.request(`/guilds/${guildId}/members/${userId}`);
    if (res.status === 404) {
      return null;
    }
    const raw = (await res.json()) as { user?: { id?: string }; presence?: Record<string, unknown> };
    if (!raw.presence) {
      return null;
    }
    return toPresenceState(raw.presence as Parameters<typeof toPresenceState>[0]);
  }

  /** Setzt den Bot-eigenen Status (Activity). */
  async setActivity(
    activityType: number,
    name: string,
    details?: string,
    state?: string,
  ): Promise<void> {
    const body: Record<string, unknown> = {
      activities: [{
        type: activityType,
        name,
        ...(details ? { details } : {}),
        ...(state ? { state } : {}),
      }],
      status: "online",
      since: null,
      afk: false,
    };
    await this.request("/users/@me/status", {
      method: "PATCH",
      body: JSON.stringify(body),
      headers: { "Content-Type": "application/json" },
    });
  }

  private async request(path: string, init: RequestInit = {}): Promise<Response> {
    let res: Response;
    try {
      res = await this.fetchImpl(`${apiBase()}${path}`, {
        ...init,
        headers: {
          ...init.headers,
          Authorization: `Bot ${this.token}`,
        },
      });
    } catch (err) {
      throw new DiscordApiError(`Discord-API nicht erreichbar (${path}): ${(err as Error).message}`);
    }
    if (res.status === 401) {
      throw new DiscordApiError(`Nicht autorisiert (${path}) — Token erneuern.`);
    }
    if (!res.ok && res.status !== 404) {
      throw new DiscordApiError(`Discord-API ${res.status} (${path}).`);
    }
    return res;
  }
}
