// Zuletzt geaendert: 2026-08-17
// Konfiguration fuer den VoIP-Service (Twilio) aus Umgebungsvariablen.
// Twilio Credentials werden fuer Telefonate benoetigt.

export interface VoIPConfig {
  accountSid: string;
  authToken: string;
  fromNumber: string;
  apiUrl: string;
}

const DEFAULT_API_BASE = "https://api.twilio.com/2010-04-01";

/** Laedt die VoIP-Konfiguration aus der Umgebung. Gibt null zurueck, wenn
 *  keine Credentials gesetzt sind (Service bleibt dann reiner Stub). */
export function loadVoIPConfig(env: Record<string, string | undefined> = process.env): VoIPConfig | null {
  const accountSid = env.TWILIO_ACCOUNT_SID?.trim() ?? "";
  const authToken = env.TWILIO_AUTH_TOKEN?.trim() ?? "";
  const fromNumber = env.TWILIO_FROM_NUMBER?.trim() ?? "";
  if (!accountSid || !authToken || !fromNumber) {
    return null;
  }
  const apiUrl = env.TWILIO_API_URL?.trim() ?? DEFAULT_API_BASE;
  return { accountSid, authToken, fromNumber, apiUrl };
}
