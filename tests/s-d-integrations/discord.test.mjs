// Zuletzt geaendert: 2026-08-17
// Mock-Server-Tests fuer den Discord-Service (S-D) — ohne echte Credentials.
// Testet Presence-Mapping, Client-API und Session gegen einen lokalen HTTP-Server.

import assert from "node:assert/strict";
import { createServer } from "node:http";
import { test } from "node:test";

process.env.DISCORD_API_BASE = "http://127.0.0.1:1/api";

const { loadDiscordConfig } = await import("../../services/s-d-integrations/dist/discord/config.js");
const { toPresenceState } = await import("../../services/s-d-integrations/dist/discord/presence.js");
const { DiscordClient, DiscordApiError } = await import("../../services/s-d-integrations/dist/discord/client.js");
const { DiscordSession } = await import("../../services/s-d-integrations/dist/discord/session.js");

function startMockServer(handlers) {
  const calls = [];
  const server = createServer((req, res) => {
    let body = "";
    req.on("data", (c) => (body += c));
    req.on("end", () => {
      const url = new URL(req.url, "http://127.0.0.1");
      calls.push({ method: req.method, path: url.pathname, headers: req.headers, body, url });
      const handler = handlers[`${req.method} ${url.pathname}`];
      if (!handler) {
        res.writeHead(404, { "content-type": "application/json" });
        res.end("{}");
        return;
      }
      handler(req, res, url);
    });
  });
  const ready = new Promise((resolve) => server.listen(0, "127.0.0.1", () => resolve(server)));
  return { server, calls, ready: ready.then(() => ({ port: server.address().port, calls })) };
}

test("loadDiscordConfig gibt null bei fehlenden Credentials", () => {
  assert.equal(loadDiscordConfig({}), null);
  assert.equal(loadDiscordConfig({ DISCORD_BOT_TOKEN: "tok" }), null);
  assert.equal(loadDiscordConfig({ DISCORD_APPLICATION_ID: "app" }), null);
});

test("loadDiscordConfig liefert Config bei gueltigen Credentials", () => {
  const cfg = loadDiscordConfig({
    DISCORD_BOT_TOKEN: "tok-123",
    DISCORD_APPLICATION_ID: "app-123",
  });
  assert.notEqual(cfg, null);
  assert.equal(cfg.botToken, "tok-123");
  assert.equal(cfg.applicationId, "app-123");
  assert.ok(cfg.pollIntervalMs > 0);
});

test("toPresenceState mappt Status und Aktivitaet", () => {
  const raw = {
    status: "online",
    activities: [{
      type: 0,
      name: "Visual Studio Code",
      details: "Editing index.ts",
      state: "workspace: nexus-hud",
      assets: { large_image: "vscode", large_text: "VS Code" },
    }],
  };
  const state = toPresenceState(raw);
  assert.notEqual(state, null);
  assert.equal(state.status, "online");
  assert.equal(state.activity, "Visual Studio Code");
  assert.equal(state.activityType, "playing");
  assert.equal(state.details, "Editing index.ts");
  assert.equal(state.state, "workspace: nexus-hud");
  assert.equal(state.largeImageKey, "vscode");
  assert.equal(state.largeImageText, "VS Code");
});

test("toPresenceState gibt null bei leerem Status", () => {
  assert.equal(toPresenceState({}), null);
  assert.equal(toPresenceState({ status: undefined }), null);
});

test("toPresenceState behandelt unbekannte Status-Werte", () => {
  const state = toPresenceState({ status: "bogus" });
  assert.equal(state.status, "offline");
});

test("toPresenceState mappt verschiedene Activity-Typen", () => {
  assert.equal(toPresenceState({ status: "online", activities: [{ type: 2, name: "Spotify" }] }).activityType, "listening");
  assert.equal(toPresenceState({ status: "idle", activities: [{ type: 3, name: "YouTube" }] }).activityType, "watching");
  assert.equal(toPresenceState({ status: "dnd", activities: [{ type: 5, name: "Chess" }] }).activityType, "competing");
});

test("DiscordClient ruft API mit Bot-Token", async () => {
  const { server, calls, ready } = startMockServer({
    "GET /api/guilds/g1/members/u1": (_req, res) => {
      res.writeHead(200, { "content-type": "application/json" });
      res.end(JSON.stringify({
        user: { id: "u1" },
        presence: { status: "online", activities: [{ type: 0, name: "Game" }] },
      }));
    },
  });
  try {
    process.env.DISCORD_API_BASE = `http://127.0.0.1:${(await ready).port}/api`;
    const client = new DiscordClient("bot-token");
    const state = await client.getPresence("u1", "g1");
    assert.equal(state.status, "online");
    assert.equal(state.activity, "Game");
    assert.match(calls[0].headers.authorization, /Bot bot-token/);
  } finally {
    server.close();
  }
});

test("DiscordClient liefert null bei 404", async () => {
  const { server, ready } = startMockServer({});
  try {
    process.env.DISCORD_API_BASE = `http://127.0.0.1:${(await ready).port}/api`;
    const client = new DiscordClient("tok");
    assert.equal(await client.getPresence("u1", "g1"), null);
  } finally {
    server.close();
  }
});

test("DiscordClient wirft bei 401", async () => {
  const { server, ready } = startMockServer({
    "GET /api/guilds/g1/members/u1": (_req, res) => {
      res.writeHead(401, { "content-type": "application/json" });
      res.end("{}");
    },
  });
  try {
    process.env.DISCORD_API_BASE = `http://127.0.0.1:${(await ready).port}/api`;
    const client = new DiscordClient("bad-token");
    await assert.rejects(() => client.getPresence("u1", "g1"), DiscordApiError);
  } finally {
    server.close();
  }
});

test("DiscordSession setzt Activity", async () => {
  const { server, calls, ready } = startMockServer({
    "PATCH /api/users/@me/status": (_req, res) => {
      res.writeHead(200, { "content-type": "application/json" });
      res.end("{}");
    },
  });
  try {
    process.env.DISCORD_API_BASE = `http://127.0.0.1:${(await ready).port}/api`;
    const session = new DiscordSession({ botToken: "tok", applicationId: "app", pollIntervalMs: 5000 });
    await session.setActivity(0, "NEXUS HUD", "Building...");
    assert.equal(calls[0].method, "PATCH");
    const body = JSON.parse(calls[0].body);
    assert.equal(body.activities[0].name, "NEXUS HUD");
  } finally {
    server.close();
  }
});
