// Zuletzt geaendert: 2026-08-17
// S-D — Integrated Apps: verbindet mit dem Dev-Bus, sendet Hello und — sofern
// Credentials gesetzt sind — periodisch Events sendet und cmd.*-Kommandos verarbeitet.
// Spotify: SPOTIFY_*-Env-Vars | Discord: DISCORD_*-Env-Vars |
// WhatsApp: WHATSAPP_*-Env-Vars | VoIP: TWILIO_*-Env-Vars

import { createRequire } from "module";
import { dirname, join } from "path";
import { fileURLToPath } from "url";
import WebSocket from "ws";

import { loadSpotifyConfig } from "./spotify/config.ts";
import type { MediaState } from "./spotify/media.ts";
import { SpotifySession } from "./spotify/session.ts";

import { loadDiscordConfig } from "./discord/config.ts";
import type { PresenceState } from "./discord/presence.ts";
import { DiscordSession } from "./discord/session.ts";

import { loadWhatsAppConfig } from "./whatsapp/config.ts";
import type { WhatsAppMessage } from "./whatsapp/message.ts";
import { WhatsAppClient } from "./whatsapp/client.ts";

import { loadVoIPConfig } from "./voip/config.ts";
import type { CallState } from "./voip/call.ts";
import { VoIPSession } from "./voip/session.ts";

import { handleEventReaction, type ReactionContext } from "./reactions.ts";

const require = createRequire(import.meta.url);
const here = dirname(fileURLToPath(import.meta.url));
const VERSION = (require(join(here, "..", "package.json")) as { version: string }).version;

const PORT = Number(process.env.NEXUS_WS_PORT ?? 49152);
const SOURCE = "S-D";
const SERVICE_ID = "s-d-integrations";
const PROTOCOL_VERSION = 1;
const HELLO_INTERVAL_MS = 5000;

const url = `ws://127.0.0.1:${PORT}/`;
const ws = new WebSocket(url);

function baseParams(extra: Record<string, unknown> = {}): Record<string, unknown> {
  return {
    source: SOURCE,
    protocol_version: PROTOCOL_VERSION,
    ts: new Date().toISOString(),
    ...extra,
  };
}

function notify(method: string, extra: Record<string, unknown> = {}): string {
  return JSON.stringify({ jsonrpc: "2.0", method, params: baseParams(extra) });
}

function hello(): string {
  return notify("event.system.hello", {
    service_id: SERVICE_ID,
    version: VERSION,
  });
}

// --- Spotify ---
const spotifyConfig = loadSpotifyConfig();
const spotify = spotifyConfig ? new SpotifySession(spotifyConfig) : null;

// --- Discord ---
const discordConfig = loadDiscordConfig();
const discord = discordConfig ? new DiscordSession(discordConfig) : null;

// --- WhatsApp ---
const whatsappConfig = loadWhatsAppConfig();
const whatsapp = whatsappConfig ? new WhatsAppClient(whatsappConfig.apiUrl, whatsappConfig.apiToken) : null;

// --- VoIP ---
const voipConfig = loadVoIPConfig();
const voip = voipConfig ? new VoIPSession(voipConfig) : null;

// --- Aktive Anrufe ---
const activeCalls = new Map<string, CallState>();

// --- Event-Reactions ---
const reactionCtx: ReactionContext = {
  spotify: spotify as unknown as import("./spotify/session.ts").SpotifySession | null,
  discord: discord as unknown as import("./discord/session.ts").DiscordSession | null,
  notify,
  wsSend: (data: string) => ws.send(data),
};

// --- Spotify Polling ---
let lastMediaJson = "";

async function pollSpotify(): Promise<void> {
  if (!spotify) return;
  try {
    const state: MediaState | null = await spotify.getCurrentlyPlaying();
    const json = JSON.stringify(state ?? null);
    if (json !== lastMediaJson) {
      lastMediaJson = json;
      if (state) {
        ws.send(notify("event.media.state", { ...state }));
      }
    }
  } catch (err) {
    console.error(`[${SERVICE_ID}] Spotify-Poll fehlgeschlagen: ${(err as Error).message}`);
  }
}

// --- Discord Polling ---
let lastPresenceJson = "";

async function pollDiscord(): Promise<void> {
  if (!discord) return;
  try {
    const userId = process.env.DISCORD_USER_ID ?? "";
    const guildId = process.env.DISCORD_GUILD_ID ?? "";
    if (!userId || !guildId) return;
    const state: PresenceState | null = await discord.getPresence(userId, guildId);
    const json = JSON.stringify(state ?? null);
    if (json !== lastPresenceJson) {
      lastPresenceJson = json;
      ws.send(notify("event.presence.changed", { service: "discord", ...state }));
    }
  } catch (err) {
    console.error(`[${SERVICE_ID}] Discord-Poll fehlgeschlagen: ${(err as Error).message}`);
  }
}

// --- Command Handlers ---

async function handleMediaCommand(method: string, params: Record<string, unknown>): Promise<void> {
  if (!spotify) {
    console.warn(`[${SERVICE_ID}] Spotify nicht konfiguriert — ${method} ignoriert.`);
    return;
  }
  try {
    switch (method) {
      case "cmd.media.toggle": {
        const current = await spotify.getCurrentlyPlaying();
        await spotify.setPlaying(!(current?.playing ?? false));
        break;
      }
      case "cmd.media.next":
        await spotify.next();
        break;
      case "cmd.media.volume": {
        const volume = Number(params.volume);
        if (!Number.isInteger(volume)) {
          console.warn(`[${SERVICE_ID}] cmd.media.volume ohne gueltiges volume (bekam ${params.volume}).`);
          return;
        }
        await spotify.setVolume(volume);
        break;
      }
      default:
        console.warn(`[${SERVICE_ID}] unbekanntes Medien-Kommando: ${method}`);
    }
  } catch (err) {
    console.error(`[${SERVICE_ID}] ${method} fehlgeschlagen: ${(err as Error).message}`);
  }
}

async function handleCallCommand(method: string, params: Record<string, unknown>): Promise<void> {
  if (!voip) {
    console.warn(`[${SERVICE_ID}] VoIP nicht konfiguriert — ${method} ignoriert.`);
    return;
  }
  try {
    switch (method) {
      case "cmd.call.make": {
        const to = String(params.to ?? "");
        if (!to) {
          console.warn(`[${SERVICE_ID}] cmd.call.make ohne Zielnummer.`);
          return;
        }
        const state = await voip.makeCall(to);
        activeCalls.set(state.callId, state);
        ws.send(notify("event.call.initiated", {
          call_id: state.callId,
          to: state.to,
          from: state.from,
          status: state.status,
        }));
        break;
      }
      case "cmd.call.hangup": {
        const callId = String(params.call_id ?? "");
        if (!callId) {
          console.warn(`[${SERVICE_ID}] cmd.call.hangup ohne call_id.`);
          return;
        }
        await voip.hangupCall(callId);
        activeCalls.delete(callId);
        ws.send(notify("event.call.ended", {
          call_id: callId,
          status: "completed",
        }));
        break;
      }
      case "cmd.call.status": {
        const callId = String(params.call_id ?? "");
        if (!callId) return;
        const state = await voip.getCallStatus(callId);
        if (state) {
          activeCalls.set(callId, state);
          ws.send(notify("event.call.status", {
            call_id: state.callId,
            to: state.to,
            from: state.from,
            status: state.status,
            duration_sec: state.durationSec ?? 0,
          }));
        }
        break;
      }
      default:
        console.warn(`[${SERVICE_ID}] unbekanntes Call-Kommando: ${method}`);
    }
  } catch (err) {
    console.error(`[${SERVICE_ID}] ${method} fehlgeschlagen: ${(err as Error).message}`);
  }
}

async function handleProfileCommand(method: string, params: Record<string, unknown>): Promise<void> {
  if (method === "cmd.profile.switch") {
    const profile = String(params.profile ?? "");
    if (!profile) {
      console.warn(`[${SERVICE_ID}] cmd.profile.switch ohne profile.`);
      return;
    }
    ws.send(notify("event.profile.switched", { profile }));
    console.log(`[${SERVICE_ID}] Profil gewechselt: ${profile}`);
  }
}

async function handleMessage(method: string, params: Record<string, unknown>): Promise<void> {
  if (method.startsWith("cmd.media.")) {
    await handleMediaCommand(method, params);
  } else if (method.startsWith("cmd.call.")) {
    await handleCallCommand(method, params);
  } else if (method.startsWith("cmd.profile.")) {
    await handleProfileCommand(method, params);
  } else if (method.startsWith("event.")) {
    await handleEventReaction(reactionCtx, method, params);
  }
}

// --- WebSocket Events ---

ws.on("open", () => {
  console.log(`[${SERVICE_ID}] verbunden mit ${url}`);
  ws.send(hello());
  setInterval(() => {
    if (ws.readyState === WebSocket.OPEN) {
      ws.send(hello());
    }
  }, HELLO_INTERVAL_MS);

  if (spotify) {
    console.log(`[${SERVICE_ID}] Spotify-Service aktiv (Poll alle ${spotifyConfig!.pollIntervalMs} ms).`);
    void pollSpotify();
    setInterval(() => void pollSpotify(), spotifyConfig!.pollIntervalMs);
  } else {
    console.log(`[${SERVICE_ID}] Spotify nicht konfiguriert (SPOTIFY_CLIENT_ID/SECRET/REFRESH_TOKEN) — reiner Stub.`);
  }

  if (discord) {
    console.log(`[${SERVICE_ID}] Discord-Service aktiv (Poll alle ${discordConfig!.pollIntervalMs} ms).`);
    void pollDiscord();
    setInterval(() => void pollDiscord(), discordConfig!.pollIntervalMs);
  } else {
    console.log(`[${SERVICE_ID}] Discord nicht konfiguriert (DISCORD_BOT_TOKEN/APPLICATION_ID) — reiner Stub.`);
  }

  if (whatsapp) {
    console.log(`[${SERVICE_ID}] WhatsApp-Service aktiv.`);
  } else {
    console.log(`[${SERVICE_ID}] WhatsApp nicht konfiguriert (WHATSAPP_API_URL/API_TOKEN) — reiner Stub.`);
  }

  if (voip) {
    console.log(`[${SERVICE_ID}] VoIP-Service aktiv (Twilio).`);
  } else {
    console.log(`[${SERVICE_ID}] VoIP nicht konfiguriert (TWILIO_ACCOUNT_SID/AUTH_TOKEN/FROM_NUMBER) — reiner Stub.`);
  }
});

ws.on("message", (data) => {
  let msg: { method?: unknown; params?: unknown } | null = null;
  try {
    msg = JSON.parse(String(data)) as { method?: unknown; params?: unknown };
  } catch {
    return;
  }
  if (typeof msg?.method === "string") {
    const params = (msg.params ?? {}) as Record<string, unknown>;
    void handleMessage(msg.method, params);
  }
});

ws.on("error", (err) => {
  console.error(`[${SERVICE_ID}] Fehler: ${err.message}`);
});
