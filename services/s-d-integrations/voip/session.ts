// Zuletzt geaendert: 2026-08-17
// VoIPSession: haelt Twilio-Credentials und reicht API-Aufrufe an den
// VoIPClient durch. Bei 401 wird die Session als ungueltig markiert.

import type { VoIPConfig } from "./config.ts";
import type { FetchLike } from "./client.ts";
import { VoIPApiError, VoIPClient } from "./client.ts";
import type { CallState } from "./call.ts";

export class VoIPSession {
  private client: VoIPClient;
  private valid = false;

  constructor(
    private readonly config: VoIPConfig,
    private readonly fetchImpl: FetchLike = fetch,
  ) {
    this.client = new VoIPClient(
      config.accountSid,
      config.authToken,
      config.fromNumber,
      config.apiUrl,
      fetchImpl,
    );
    this.valid = true;
  }

  async makeCall(to: string, url?: string): Promise<CallState> {
    return this.withAuth((c) => c.makeCall(to, url));
  }

  async getCallStatus(callId: string): Promise<CallState | null> {
    return this.withAuth((c) => c.getCallStatus(callId));
  }

  async hangupCall(callId: string): Promise<void> {
    return this.withAuth((c) => c.hangupCall(callId));
  }

  private async withAuth<T>(fn: (client: VoIPClient) => Promise<T>): Promise<T> {
    if (!this.valid) {
      throw new VoIPApiError("VoIP-Session inaktiv — Credentials ungueltig.");
    }
    try {
      return await fn(this.client);
    } catch (err) {
      if (err instanceof VoIPApiError && err.message.startsWith("Nicht autorisiert")) {
        this.valid = false;
      }
      throw err;
    }
  }
}
