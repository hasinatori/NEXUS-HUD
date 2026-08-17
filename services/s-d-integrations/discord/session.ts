// Zuletzt geaendert: 2026-08-17
// DiscordSession: haelt das Bot-Token und reicht API-Aufrufe an den
// DiscordClient durch. Bei 401 wird das Token als ungueltig markiert.

import type { DiscordConfig } from "./config.ts";
import type { FetchLike } from "./client.ts";
import { DiscordApiError, DiscordClient } from "./client.ts";
import type { PresenceState } from "./presence.ts";

export class DiscordSession {
  private client: DiscordClient;
  private valid = false;

  constructor(
    private readonly config: DiscordConfig,
    private readonly fetchImpl: FetchLike = fetch,
  ) {
    this.client = new DiscordClient(config.botToken, fetchImpl);
    this.valid = true;
  }

  async getPresence(userId: string, guildId: string): Promise<PresenceState | null> {
    return this.withAuth((c) => c.getPresence(userId, guildId));
  }

  async setActivity(
    activityType: number,
    name: string,
    details?: string,
    state?: string,
  ): Promise<void> {
    return this.withAuth((c) => c.setActivity(activityType, name, details, state));
  }

  private async withAuth<T>(fn: (client: DiscordClient) => Promise<T>): Promise<T> {
    if (!this.valid) {
      throw new DiscordApiError("Discord-Session inaktiv — Bot-Token ungueltig.");
    }
    try {
      return await fn(this.client);
    } catch (err) {
      if (err instanceof DiscordApiError && err.message.startsWith("Nicht autorisiert")) {
        this.valid = false;
      }
      throw err;
    }
  }
}
