// Zuletzt geaendert: 2026-08-17
// Konfiguration fuer den WhatsApp-Service aus Umgebungsvariablen.
// WhatsApp-Web-Automation unterliegt ToS-Risiken — nur Push-Webhooks
// ueber offizielle API (z.B. Twilio WhatsApp) sind sicher.

export interface WhatsAppConfig {
  apiUrl: string;
  apiToken: string;
  webhookPath: string;
}

/** Laedt die WhatsApp-Konfiguration aus der Umgebung. Gibt null zurueck, wenn
 *  keine Credentials gesetzt sind (Service bleibt dann reiner Stub). */
export function loadWhatsAppConfig(env: Record<string, string | undefined> = process.env): WhatsAppConfig | null {
  const apiUrl = env.WHATSAPP_API_URL?.trim() ?? "";
  const apiToken = env.WHATSAPP_API_TOKEN?.trim() ?? "";
  const webhookPath = env.WHATSAPP_WEBHOOK_PATH?.trim() ?? "/webhook/whatsapp";
  if (!apiUrl || !apiToken) {
    return null;
  }
  return { apiUrl, apiToken, webhookPath };
}
