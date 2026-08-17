// Zuletzt geaendert: 2026-08-17
// Mock-Server-Tests fuer den VoIP-Service (S-D) — Twilio-basiert.
// Testet Call-Mapping, Client-API und Session gegen einen lokalen HTTP-Server.

import assert from "node:assert/strict";
import { createServer } from "node:http";
import { test } from "node:test";

process.env.TWILIO_API_URL = "http://127.0.0.1:1";

const { loadVoIPConfig } = await import("../../services/s-d-integrations/dist/voip/config.js");
const { toCallState } = await import("../../services/s-d-integrations/dist/voip/call.js");
const { VoIPClient, VoIPApiError } = await import("../../services/s-d-integrations/dist/voip/client.js");
const { VoIPSession } = await import("../../services/s-d-integrations/dist/voip/session.js");

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

test("loadVoIPConfig gibt null bei fehlenden Credentials", () => {
  assert.equal(loadVoIPConfig({}), null);
  assert.equal(loadVoIPConfig({ TWILIO_ACCOUNT_SID: "AC123" }), null);
});

test("loadVoIPConfig liefert Config bei gueltigen Credentials", () => {
  const cfg = loadVoIPConfig({
    TWILIO_ACCOUNT_SID: "AC123",
    TWILIO_AUTH_TOKEN: "secret",
    TWILIO_FROM_NUMBER: "+15551234567",
  });
  assert.notEqual(cfg, null);
  assert.equal(cfg.accountSid, "AC123");
  assert.equal(cfg.authToken, "secret");
  assert.equal(cfg.fromNumber, "+15551234567");
  assert.match(cfg.apiUrl, /twilio\.com/);
});

test("toCallState mappt gueltigen Anruf", () => {
  const state = toCallState({
    sid: "CA123",
    to: "+15559876543",
    from: "+15551234567",
    status: "completed",
    duration: "120",
    start_time: "2026-08-17T10:00:00Z",
    end_time: "2026-08-17T10:02:00Z",
  });
  assert.notEqual(state, null);
  assert.equal(state.callId, "CA123");
  assert.equal(state.to, "+15559876543");
  assert.equal(state.status, "completed");
  assert.equal(state.durationSec, 120);
});

test("toCallState gibt null bei unvollstaendigen Daten", () => {
  assert.equal(toCallState({}), null);
  assert.equal(toCallState({ sid: "CA123" }), null);
});

test("toCallState behandelt Fehler-Status", () => {
  const state = toCallState({
    sid: "CA456",
    to: "+15559876543",
    from: "+15551234567",
    status: "failed",
    error_code: 21215,
    error_message: "Invalid phone number",
  });
  assert.equal(state.status, "failed");
  assert.equal(state.errorCode, 21215);
  assert.equal(state.errorMessage, "Invalid phone number");
});

test("VoIPClient initiert Anruf mit Basic Auth", async () => {
  const { server, calls, ready } = startMockServer({
    "POST /api/Accounts.json/Calls.json": (_req, res) => {
      res.writeHead(201, { "content-type": "application/json" });
      res.end(JSON.stringify({
        sid: "CA789",
        to: "+15559876543",
        from: "+15551234567",
        status: "queued",
        duration: "0",
      }));
    },
  });
  try {
    process.env.TWILIO_API_URL = `http://127.0.0.1:${(await ready).port}/api`;
    const client = new VoIPClient("AC123", "secret", "+15551234567", `http://127.0.0.1:${(await ready).port}/api`);
    const state = await client.makeCall("+15559876543");
    assert.equal(state.callId, "CA789");
    assert.equal(state.status, "queued");
    assert.match(calls[0].headers.authorization, /Basic /);
    assert.match(calls[0].body, /To=%2B15559876543/);
  } finally {
    server.close();
  }
});

test("VoIPClient fragt Anruf-Status ab", async () => {
  const { server, ready } = startMockServer({
    "GET /api/Accounts.json/Calls/CA789.json": (_req, res) => {
      res.writeHead(200, { "content-type": "application/json" });
      res.end(JSON.stringify({
        sid: "CA789",
        to: "+15559876543",
        from: "+15551234567",
        status: "in-progress",
        duration: "30",
      }));
    },
  });
  try {
    process.env.TWILIO_API_URL = `http://127.0.0.1:${(await ready).port}/api`;
    const client = new VoIPClient("AC123", "secret", "+15551234567", `http://127.0.0.1:${(await ready).port}/api`);
    const state = await client.getCallStatus("CA789");
    assert.equal(state.status, "in-progress");
    assert.equal(state.durationSec, 30);
  } finally {
    server.close();
  }
});

test("VoIPClient beendet Anruf", async () => {
  const { server, calls, ready } = startMockServer({
    "POST /api/Accounts.json/Calls/CA789.json": (_req, res) => {
      res.writeHead(200, { "content-type": "application/json" });
      res.end(JSON.stringify({ sid: "CA789", status: "completed" }));
    },
  });
  try {
    process.env.TWILIO_API_URL = `http://127.0.0.1:${(await ready).port}/api`;
    const client = new VoIPClient("AC123", "secret", "+15551234567", `http://127.0.0.1:${(await ready).port}/api`);
    await client.hangupCall("CA789");
    assert.equal(calls[0].method, "POST");
    assert.match(calls[0].body, /Status=completed/);
  } finally {
    server.close();
  }
});

test("VoIPClient wirft bei 401", async () => {
  const { server, ready } = startMockServer({
    "POST /api/Accounts.json/Calls.json": (_req, res) => {
      res.writeHead(401, { "content-type": "application/json" });
      res.end("{}");
    },
  });
  try {
    process.env.TWILIO_API_URL = `http://127.0.0.1:${(await ready).port}/api`;
    const client = new VoIPClient("AC123", "bad", "+15551234567", `http://127.0.0.1:${(await ready).port}/api`);
    await assert.rejects(() => client.makeCall("+1"), VoIPApiError);
  } finally {
    server.close();
  }
});

test("VoIPSession haelt Session-Status", async () => {
  const { server, ready } = startMockServer({
    "POST /api/Accounts.json/Calls.json": (_req, res) => {
      res.writeHead(401, { "content-type": "application/json" });
      res.end("{}");
    },
  });
  try {
    process.env.TWILIO_API_URL = `http://127.0.0.1:${(await ready).port}/api`;
    const session = new VoIPSession({
      accountSid: "AC123",
      authToken: "bad",
      fromNumber: "+15551234567",
      apiUrl: `http://127.0.0.1:${(await ready).port}/api`,
    });
    await assert.rejects(() => session.makeCall("+1"), VoIPApiError);
    await assert.rejects(() => session.makeCall("+1"), VoIPApiError);
  } finally {
    server.close();
  }
});
