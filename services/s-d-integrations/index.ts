// Zuletzt geändert: 2026-08-16
// S-D — Integrated Apps: verbindet mit dem Dev-Bus, sendet Hello und — sofern
// Spotify-Credentials gesetzt sind (SPOTIFY_CLIENT_ID/SECRET/REFRESH_TOKEN) —
// periodisch event.media.state und verarbeitet cmd.media.*-Kommandos.

import { createRequire } from "module";
import { dirname, join } from "path";
import { fileURLToPath } from "url";
import WebSocket from "ws";

import { loadSpotifyConfig } from "./spotify/config.ts";
import type { MediaState } from "./spotify/media.ts";
import { SpotifySession } from "./spotify/session.ts";
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

// Spotify nur aktivieren, wenn Credentials vorhanden sind.
const spotifyConfig = loadSpotifyConfig();
const spotify = spotifyConfig ? new SpotifySession(spotifyConfig) : null;

let lastMediaJson = "";

async function pollSpotify(): Promise<void> {
  if (!spotify) {
    return;
  }
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
          console.warn(`[${SERVICE_ID}] cmd.media.volume ohne gültiges volume (bekam ${params.volume}).`);
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
});

ws.on("message", (data) => {
  let msg: { method?: unknown; params?: unknown } | null = null;
  try {
    msg = JSON.parse(String(data)) as { method?: unknown; params?: unknown };
  } catch {
    return; // kein JSON-RPC -> ignorieren
  }
  if (typeof msg?.method === "string" && msg.method.startsWith("cmd.media.")) {
    const params = (msg.params ?? {}) as Record<string, unknown>;
    void handleMediaCommand(msg.method, params);
  }
});

ws.on("error", (err) => {
  console.error(`[${SERVICE_ID}] Fehler: ${err.message}`);
});
