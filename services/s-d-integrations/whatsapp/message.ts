// Zuletzt geaendert: 2026-08-17
// WhatsApp-Nachrichten-Typ und Mapping. Reine Funktion -> ohne HTTP testbar.

export interface WhatsAppMessage {
  id: string;
  from: string;
  fromName?: string;
  body: string;
  timestamp: string;
  type: "text" | "image" | "audio" | "video" | "document";
}

interface RawMessage {
  id?: string;
  from?: string;
  from_name?: string;
  body?: string;
  timestamp?: string;
  type?: string;
}

const VALID_TYPES = ["text", "image", "audio", "video", "document"];

/** Mappt eine rohe Webhook-Nachricht auf WhatsAppMessage. */
export function toWhatsAppMessage(raw: RawMessage): WhatsAppMessage | null {
  if (!raw.id || !raw.from || !raw.body) {
    return null;
  }
  const msgType = VALID_TYPES.includes(raw.type ?? "") ? (raw.type as WhatsAppMessage["type"]) : "text";
  return {
    id: raw.id,
    from: raw.from,
    fromName: raw.from_name,
    body: raw.body,
    timestamp: raw.timestamp ?? new Date().toISOString(),
    type: msgType,
  };
}
