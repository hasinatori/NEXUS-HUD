// Zuletzt geaendert: 2026-08-17
// Konfiguration fuer den Discord-Service aus Umgebungsvariablen.
// Discord Bot Token wird fuer Rich Presence und Bot-Funktionen benoetigt.

export interface DiscordConfig {
  botToken: string;
  applicationId: string;
  pollIntervalMs: number;
}

const DEFAULT_POLL_INTERVAL_MS = 30_000;

/** Laedt die Discord-Konfiguration aus der Umgebung. Gibt null zurueck, wenn
 *  keine Credentials gesetzt sind (Service bleibt dann reiner Stub). */
export function loadDiscordConfig(env: Record<string, string | undefined> = process.env): DiscordConfig | null {
  const botToken = env.DISCORD_BOT_TOKEN?.trim() ?? "";
  const applicationId = env.DISCORD_APPLICATION_ID?.trim() ?? "";
  if (!botToken || !applicationId) {
    return null;
  }
  const poll = Number(env.DISCORD_POLL_INTERVAL_MS ?? DEFAULT_POLL_INTERVAL_MS);
  return {
    botToken,
    applicationId,
    pollIntervalMs: Number.isFinite(poll) && poll > 0 ? poll : DEFAULT_POLL_INTERVAL_MS,
  };
}
