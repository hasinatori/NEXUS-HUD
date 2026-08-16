// Zuletzt geändert: 2026-08-16
// Mock-Server-Tests für den Spotify-Service (S-D) — ohne echte Credentials.
// Testet Token-Refresh, API-Aufrufe, Media-Mapping und die Session (Auto-Refresh,
// 401-Retry) gegen einen lokalen HTTP-Server. Ausführen: npm test (in
// services/s-d-integrations). Läuft nach npm run build gegen dist/.

import assert from "node:assert/strict";
import { createServer } from "node:http";
import { test } from "node:test";

process.env.SPOTIFY_TOKEN_URL = "http://127.0.0.1:1/token"; // wird je Test überschrieben
process.env.SPOTIFY_API_BASE = "http://127.0.0.1:1/v1";

const { refreshAccessToken, SpotifyAuthError } = await import("../../services/s-d-integrations/dist/spotify/auth.js");
const { SpotifyClient, SpotifyApiError } = await import("../../services/s-d-integrations/dist/spotify/client.js");
const { toMediaState } = await import("../../services/s-d-integrations/dist/spotify/media.js");
const { SpotifySession } = await import("../../services/s-d-integrations/dist/spotify/session.js");

const FIXTURE = {
  is_playing: true,
  progress_ms: 42000,
  item: {
    name: "Neon Sun",
    artists: [{ name: "System Overdrive" }, { name: "VAX" }],
    album: { name: "Night Drive", images: [{ url: "https://example.com/cover.jpg" }] },
    duration_ms: 224000,
  },
};

/** Startet einen Mock-Server; handler ist { "METHOD /path": (req, res, url) => void }. */
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
        res.end(JSON.stringify({ error: "nicht gefunden" }));
        return;
      }
      handler(req, res, url);
    });
  });
  const ready = new Promise((resolve) => server.listen(0, "127.0.0.1", () => resolve(server)));
  return { server, calls, ready: ready.then(() => ({ port: server.address().port, calls })) };
}

test("refreshAccessToken sendet Basic Auth und liefert Token", async () => {
  const { server, calls, ready } = startMockServer({
    "POST /token": (_req, res) => {
      res.writeHead(200, { "content-type": "application/json" });
      res.end(JSON.stringify({ access_token: "tok-123", token_type: "Bearer", scope: "user-modify", expires_in: 3600 }));
    },
  });
  try {
    process.env.SPOTIFY_TOKEN_URL = `http://127.0.0.1:${(await ready).port}/token`;
    const token = await refreshAccessToken("id-1", "secret-1", "refresh-1");
    assert.equal(token.access_token, "tok-123");
    assert.equal(token.expires_in, 3600);

    const call = calls[0];
    assert.equal(call.path, "/token");
    assert.equal(call.headers.authorization, `Basic ${Buffer.from("id-1:secret-1").toString("base64")}`);
    assert.match(call.body, /grant_type=refresh_token/);
    assert.match(call.body, /refresh_token=refresh-1/);
  } finally {
    server.close();
  }
});

test("refreshAccessToken wirft bei nicht-200", async () => {
  const { server, ready } = startMockServer({
    "POST /token": (_req, res) => {
      res.writeHead(400, { "content-type": "application/json" });
      res.end("{}");
    },
  });
  try {
    process.env.SPOTIFY_TOKEN_URL = `http://127.0.0.1:${(await ready).port}/token`;
    await assert.rejects(() => refreshAccessToken("id", "sec", "ref"), SpotifyAuthError);
  } finally {
    server.close();
  }
});

test("toMediaState mappt Track/Artist/Album und gibt null bei leerem Playback", () => {
  const state = toMediaState(FIXTURE);
  assert.equal(state.playing, true);
  assert.equal(state.track, "Neon Sun");
  assert.equal(state.artist, "System Overdrive, VAX");
  assert.equal(state.album, "Night Drive");
  assert.equal(state.album_art_url, "https://example.com/cover.jpg");
  assert.equal(state.duration_ms, 224000);
  assert.equal(state.progress_ms, 42000);

  assert.equal(toMediaState({ is_playing: false }), null);
  assert.equal(toMediaState({ is_playing: false, item: null }), null);
});

test("SpotifyClient ruft API mit Bearer-Token und mappt die Antwort", async () => {
  const { server, calls, ready } = startMockServer({
    "GET /v1/me/player/currently-playing": (_req, res) => {
      res.writeHead(200, { "content-type": "application/json" });
      res.end(JSON.stringify(FIXTURE));
    },
  });
  try {
    process.env.SPOTIFY_API_BASE = `http://127.0.0.1:${(await ready).port}/v1`;
    const client = new SpotifyClient("tok-abc");
    const state = await client.getCurrentlyPlaying();
    assert.equal(state.track, "Neon Sun");
    assert.equal(calls[0].headers.authorization, "Bearer tok-abc");
  } finally {
    server.close();
  }
});

test("SpotifyClient liefert null bei 204 (nichts spielt)", async () => {
  const { server, ready } = startMockServer({
    "GET /v1/me/player/currently-playing": (_req, res) => {
      res.writeHead(204);
      res.end();
    },
  });
  try {
    process.env.SPOTIFY_API_BASE = `http://127.0.0.1:${(await ready).port}/v1`;
    const client = new SpotifyClient("tok-abc");
    assert.equal(await client.getCurrentlyPlaying(), null);
  } finally {
    server.close();
  }
});

test("setVolume lehnt Werte außerhalb 0-100 ab", () => {
  const client = new SpotifyClient("tok");
  assert.rejects(() => client.setVolume(101), SpotifyApiError);
  assert.rejects(() => client.setVolume(-1), SpotifyApiError);
});

test("SpotifySession holt Token beim ersten Aufruf (Refresh-Flow)", async () => {
  const { server, calls, ready } = startMockServer({
    "POST /token": (_req, res) => {
      res.writeHead(200, { "content-type": "application/json" });
      res.end(JSON.stringify({ access_token: "tok-fresh", expires_in: 3600, token_type: "Bearer", scope: "" }));
    },
    "GET /v1/me/player/currently-playing": (_req, res) => {
      res.writeHead(200, { "content-type": "application/json" });
      res.end(JSON.stringify(FIXTURE));
    },
  });
  try {
    const { port } = await ready;
    process.env.SPOTIFY_TOKEN_URL = `http://127.0.0.1:${port}/token`;
    process.env.SPOTIFY_API_BASE = `http://127.0.0.1:${port}/v1`;
    const session = new SpotifySession({
      clientId: "id",
      clientSecret: "sec",
      refreshToken: "ref",
      pollIntervalMs: 1000,
    });
    const state = await session.getCurrentlyPlaying();
    assert.equal(state.track, "Neon Sun");
    assert.equal(calls.filter((c) => c.path === "/token").length, 1);
    assert.equal(calls.find((c) => c.path.endsWith("/currently-playing")).headers.authorization, "Bearer tok-fresh");
  } finally {
    server.close();
  }
});

test("SpotifySession wiederholt bei 401 mit frischem Token", async () => {
  let apiCalls = 0;
  const { server, calls, ready } = startMockServer({
    "POST /token": (_req, res) => {
      res.writeHead(200, { "content-type": "application/json" });
      res.end(JSON.stringify({ access_token: `tok-${apiCalls}`, expires_in: 3600, token_type: "Bearer", scope: "" }));
    },
    "GET /v1/me/player/currently-playing": (_req, res) => {
      apiCalls += 1;
      if (apiCalls === 1) {
        res.writeHead(401, { "content-type": "application/json" });
        res.end("{}");
        return;
      }
      res.writeHead(200, { "content-type": "application/json" });
      res.end(JSON.stringify(FIXTURE));
    },
  });
  try {
    const { port } = await ready;
    process.env.SPOTIFY_TOKEN_URL = `http://127.0.0.1:${port}/token`;
    process.env.SPOTIFY_API_BASE = `http://127.0.0.1:${port}/v1`;
    const session = new SpotifySession({
      clientId: "id",
      clientSecret: "sec",
      refreshToken: "ref",
      pollIntervalMs: 1000,
    });
    const state = await session.getCurrentlyPlaying();
    assert.equal(state.track, "Neon Sun");
    assert.equal(apiCalls, 2);
    // Zweiter Token-Refresh nach dem 401.
    assert.equal(calls.filter((c) => c.path === "/token").length, 2);
  } finally {
    server.close();
  }
});
