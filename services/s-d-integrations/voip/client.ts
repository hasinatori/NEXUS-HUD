// Zuletzt geaendert: 2026-08-17
// Twilio VoIP-Client: Anrufe initiieren, Status abfragen, beenden.
// fetchImpl ist fuer Tests injizierbar.

import type { CallState } from "./call.ts";
import { toCallState } from "./call.ts";

export type FetchLike = typeof fetch;

export class VoIPApiError extends Error {}

export class VoIPClient {
  private readonly fetchImpl: FetchLike;

  constructor(
    private accountSid: string,
    private authToken: string,
    private fromNumber: string,
    private apiUrl: string,
    fetchImpl: FetchLike = fetch,
  ) {
    this.fetchImpl = fetchImpl;
  }

  /** Initiert einen Anruf an die angegebene Telefonnummer. */
  async makeCall(to: string, url?: string): Promise<CallState> {
    const body = new URLSearchParams({
      To: to,
      From: this.fromNumber,
      ...(url ? { Url: url } : {}),
    });
    const res = await this.request("/Accounts.json/Calls.json", {
      method: "POST",
      body,
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
    });
    const data = (await res.json()) as Record<string, unknown>;
    const state = toCallState(data as Parameters<typeof toCallState>[0]);
    if (!state) {
      throw new VoIPApiError("Ungueltige Twilio-Antwort beim Anrufstart.");
    }
    return state;
  }

  /** Fragt den Status eines laufenden Anrufs ab. */
  async getCallStatus(callId: string): Promise<CallState | null> {
    const res = await this.request(`/Accounts.json/Calls/${callId}.json`);
    if (res.status === 404) {
      return null;
    }
    const data = (await res.json()) as Record<string, unknown>;
    return toCallState(data as Parameters<typeof toCallState>[0]);
  }

  /** Beendet einen laufenden Anruf. */
  async hangupCall(callId: string): Promise<void> {
    const body = new URLSearchParams({ Status: "completed" });
    await this.request(`/Accounts.json/Calls/${callId}.json`, {
      method: "POST",
      body,
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
    });
  }

  private async request(path: string, init: RequestInit = {}): Promise<Response> {
    let res: Response;
    try {
      res = await this.fetchImpl(`${this.apiUrl}${path}`, {
        ...init,
        headers: {
          ...init.headers,
          Authorization: `Basic ${Buffer.from(`${this.accountSid}:${this.authToken}`).toString("base64")}`,
        },
      });
    } catch (err) {
      throw new VoIPApiError(`Twilio-API nicht erreichbar (${path}): ${(err as Error).message}`);
    }
    if (res.status === 401) {
      throw new VoIPApiError(`Nicht autorisiert (${path}) — Credentials pruefen.`);
    }
    if (!res.ok) {
      const body = await res.text();
      throw new VoIPApiError(`Twilio-API ${res.status} (${path}): ${body}`);
    }
    return res;
  }
}
