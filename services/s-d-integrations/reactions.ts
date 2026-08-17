// Zuletzt geaendert: 2026-08-17
// Cross-Module Event-Reactions: Reagiert auf IPC-Events anderer Services
// und fuehrt Media-/Presence-Aktionen aus (z.B. Build-Fehler -> Discord-Status).

import type { SpotifySession } from "../spotify/session.ts";
import type { DiscordSession } from "../discord/session.ts";
import type { CallState } from "../voip/call.ts";

export interface ReactionContext {
  spotify: SpotifySession | null;
  discord: DiscordSession | null;
  notify: (method: string, extra?: Record<string, unknown>) => string;
  wsSend: (data: string) => void;
}

const ACTIVITY_PLAYING = 0;
const ACTIVITY_LISTENING = 2;

export async function handleEventReaction(
  ctx: ReactionContext,
  method: string,
  params: Record<string, unknown>,
): Promise<void> {
  switch (method) {
    case "event.build.failed":
      await onBuildFailed(ctx, params);
      break;
    case "event.build.succeeded":
      await onBuildSucceeded(ctx, params);
      break;
    case "event.profile.switched":
      await onProfileSwitched(ctx, params);
      break;
  }
}

async function onBuildFailed(
  ctx: ReactionContext,
  params: Record<string, unknown>,
): Promise<void> {
  const project = String(params.project ?? "unknown");
  console.log(`[S-D] Build fehlgeschlagen: ${project}`);

  if (ctx.discord) {
    try {
      const userId = process.env.DISCORD_USER_ID ?? "";
      const guildId = process.env.DISCORD_GUILD_ID ?? "";
      if (userId && guildId) {
        const client = (ctx.discord as { client?: unknown }).client;
        if (client && typeof (client as { setActivity?: unknown }).setActivity === "function") {
          await (client as { setActivity: (type: number, name: string, details?: string) => Promise<void> })
            .setActivity(ACTIVITY_PLAYING, `Build-Fehler: ${project}`, `Letzter Build fehlgeschlagen`);
        }
      }
    } catch (err) {
      console.error(`[S-D] Discord-Status-Update fehlgeschlagen: ${(err as Error).message}`);
    }
  }
}

async function onBuildSucceeded(
  ctx: ReactionContext,
  params: Record<string, unknown>,
): Promise<void> {
  const project = String(params.project ?? "unknown");
  console.log(`[S-D] Build erfolgreich: ${project}`);

  if (ctx.discord) {
    try {
      const userId = process.env.DISCORD_USER_ID ?? "";
      const guildId = process.env.DISCORD_GUILD_ID ?? "";
      if (userId && guildId) {
        const client = (ctx.discord as { client?: unknown }).client;
        if (client && typeof (client as { setActivity?: unknown }).setActivity === "function") {
          await (client as { setActivity: (type: number, name: string, details?: string) => Promise<void> })
            .setActivity(ACTIVITY_PLAYING, `Build OK: ${project}`, `Letzter Build erfolgreich`);
        }
      }
    } catch (err) {
      console.error(`[S-D] Discord-Status-Update fehlgeschlagen: ${(err as Error).message}`);
    }
  }
}

async function onProfileSwitched(
  ctx: ReactionContext,
  params: Record<string, unknown>,
): Promise<void> {
  const profile = String(params.profile ?? "");
  console.log(`[S-D] Profil-Switch empfangen: ${profile}`);

  if (ctx.discord) {
    try {
      const userId = process.env.DISCORD_USER_ID ?? "";
      const guildId = process.env.DISCORD_GUILD_ID ?? "";
      if (userId && guildId) {
        const statusText = profile === "gaming" ? "Gaming"
          : profile === "afk" ? "AFK"
          : "Coding";
        const activityType = profile === "gaming" ? ACTIVITY_PLAYING : ACTIVITY_LISTENING;
        const client = (ctx.discord as { client?: unknown }).client;
        if (client && typeof (client as { setActivity?: unknown }).setActivity === "function") {
          await (client as { setActivity: (type: number, name: string, details?: string) => Promise<void> })
            .setActivity(activityType, `NEXUS HUD — ${statusText}`, `Profil: ${profile}`);
        }
      }
    } catch (err) {
      console.error(`[S-D] Discord-Status-Update fehlgeschlagen: ${(err as Error).message}`);
    }
  }

  if (ctx.spotify) {
    try {
      if (profile === "gaming") {
        await ctx.spotify.setVolume(30);
      } else if (profile === "afk") {
        await ctx.spotify.setVolume(0);
      } else {
        await ctx.spotify.setVolume(70);
      }
    } catch (err) {
      console.error(`[S-D] Spotify-Lautstaeke-Anpassung fehlgeschlagen: ${(err as Error).message}`);
    }
  }
}
