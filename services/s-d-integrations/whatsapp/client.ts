// Zuletzt geaendert: 2026-08-28
// Duenner Client fuer die WhatsApp-Integration (offizielle Webhook-API).
// fetchImpl ist fuer Tests injizierbar.

import type { WhatsAppMessage } from "./message.ts";
import { toWhatsAppMessage } from "./message.ts";

/** Struktur des WhatsApp-Webhook-Values. */
interface WhatsAppWebhookValue {
  messages?: Array<Record<string, unknown>>;
  contacts?: Array<{ profile?: { name?: string }; wa_id?: string }>;
}

export type FetchLike = typeof fetch;

export class WhatsAppApiError extends Error {}

export class WhatsAppClient {
  private readonly fetchImpl: FetchLike;

  constructor(
    private apiUrl: string,
    private apiToken: string,
    fetchImpl: FetchLike = fetch,
  ) {
    this.fetchImpl = fetchImpl;
  }

  /** Sendet eine Textnachricht ueber die WhatsApp-API. */
  async sendMessage(to: string, text: string): Promise<string> {
    const body = {
      messaging_product: "whatsapp",
      to,
      type: "text",
      text: { body: text },
    };
    const res = await this.request("/messages", {
      method: "POST",
      body: JSON.stringify(body),
      headers: { "Content-Type": "application/json" },
    });
    const data = (await res.json()) as { messages?: Array<{ id?: string }> };
    return data.messages?.[0]?.id ?? "";
  }

  /** Verarbeitet einen eingehenden Webhook-Payload. */
  parseWebhook(payload: Record<string, unknown>): WhatsAppMessage | null {
    const entry = (payload.entry as Array<{ changes?: Array<{ value?: WhatsAppWebhookValue }> }>)?.[0];
    const change = entry?.changes?.[0]?.value;
    const rawMsg = change?.messages?.[0];
    if (!rawMsg) {
      return null;
    }
    const contacts = (change?.contacts as Array<{ profile?: { name?: string }; wa_id?: string }>) ?? [];
    const contact = contacts.find((c) => c.wa_id === rawMsg.from);
    return toWhatsAppMessage({
      id: rawMsg.id as string,
      from: rawMsg.from as string,
      from_name: contact?.profile?.name,
      body: (rawMsg.text as { body?: string })?.body ?? "",
      timestamp: rawMsg.timestamp as string,
      type: rawMsg.type as string,
    });
  }

  private async request(path: string, init: RequestInit = {}): Promise<Response> {
    let res: Response;
    try {
      res = await this.fetchImpl(`${this.apiUrl}${path}`, {
        ...init,
        headers: {
          ...init.headers,
          Authorization: `Bearer ${this.apiToken}`,
        },
      });
    } catch (err) {
      throw new WhatsAppApiError(`WhatsApp-API nicht erreichbar (${path}): ${(err as Error).message}`);
    }
    if (res.status === 401) {
      throw new WhatsAppApiError(`Nicht autorisiert (${path}) — Token erneuern.`);
    }
    if (!res.ok) {
      throw new WhatsAppApiError(`WhatsApp-API ${res.status} (${path}).`);
    }
    return res;
  }
}
