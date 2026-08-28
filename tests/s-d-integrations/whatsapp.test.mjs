// Zuletzt geaendert: 2026-08-17
// Mock-Server-Tests fuer den WhatsApp-Service (S-D) — ohne echte Credentials.
// Testet Nachrichten-Mapping, Client-API und Webhook-Verarbeitung.

import assert from "node:assert/strict";
import { createServer } from "node:http";
import { test } from "node:test";

process.env.WHATSAPP_API_URL = "http://127.0.0.1:1/api";

const { loadWhatsAppConfig } = await import("../../services/s-d-integrations/dist/whatsapp/config.js");
const { toWhatsAppMessage } = await import("../../services/s-d-integrations/dist/whatsapp/message.js");
const { WhatsAppClient, WhatsAppApiError } = await import("../../services/s-d-integrations/dist/whatsapp/client.js");

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

test("loadWhatsAppConfig gibt null bei fehlenden Credentials", () => {
  assert.equal(loadWhatsAppConfig({}), null);
  assert.equal(loadWhatsAppConfig({ WHATSAPP_API_URL: "http://x" }), null);
});

test("loadWhatsAppConfig liefert Config bei gueltigen Credentials", () => {
  const cfg = loadWhatsAppConfig({
    WHATSAPP_API_URL: "http://localhost:3000",
    WHATSAPP_API_TOKEN: "tok-123",
  });
  assert.notEqual(cfg, null);
  assert.equal(cfg.apiUrl, "http://localhost:3000");
  assert.equal(cfg.apiToken, "tok-123");
  assert.equal(cfg.webhookPath, "/webhook/whatsapp");
});

test("toWhatsAppMessage mappt gueltige Nachricht", () => {
  const msg = toWhatsAppMessage({
    id: "msg-1",
    from: "+1234567890",
    from_name: "Test User",
    body: "Hallo!",
    timestamp: "2026-08-17T10:00:00Z",
    type: "text",
  });
  assert.notEqual(msg, null);
  assert.equal(msg.id, "msg-1");
  assert.equal(msg.from, "+1234567890");
  assert.equal(msg.fromName, "Test User");
  assert.equal(msg.body, "Hallo!");
  assert.equal(msg.type, "text");
});

test("toWhatsAppMessage gibt null bei unvollstaendigen Daten", () => {
  assert.equal(toWhatsAppMessage({}), null);
  assert.equal(toWhatsAppMessage({ id: "m1" }), null);
  assert.equal(toWhatsAppMessage({ from: "+123" }), null);
});

test("toWhatsAppMessage behandelt unbekannten Typ", () => {
  const msg = toWhatsAppMessage({ id: "m1", from: "+1", body: "test", type: "sticker" });
  assert.equal(msg.type, "text");
});

test("WhatsAppClient sendet Nachricht mit Bearer-Token", async () => {
  const { server, calls, ready } = startMockServer({
    "POST /api/messages": (_req, res) => {
      res.writeHead(200, { "content-type": "application/json" });
      res.end(JSON.stringify({ messages: [{ id: "wamid-123" }] }));
    },
  });
  try {
    process.env.WHATSAPP_API_URL = `http://127.0.0.1:${(await ready).port}/api`;
    const client = new WhatsAppClient(`http://127.0.0.1:${(await ready).port}/api`, "tok-abc");
    const id = await client.sendMessage("+1234567890", "Test");
    assert.equal(id, "wamid-123");
    assert.match(calls[0].headers.authorization, /Bearer tok-abc/);
    const body = JSON.parse(calls[0].body);
    assert.equal(body.to, "+1234567890");
    assert.equal(body.text.body, "Test");
  } finally {
    server.close();
  }
});

test("WhatsAppClient wirft bei 401", async () => {
  const { server, ready } = startMockServer({
    "POST /api/messages": (_req, res) => {
      res.writeHead(401, { "content-type": "application/json" });
      res.end("{}");
    },
  });
  try {
    process.env.WHATSAPP_API_URL = `http://127.0.0.1:${(await ready).port}/api`;
    const client = new WhatsAppClient(`http://127.0.0.1:${(await ready).port}/api`, "bad");
    await assert.rejects(() => client.sendMessage("+1", "test"), WhatsAppApiError);
  } finally {
    server.close();
  }
});

test("WhatsAppClient verarbeitet Webhook-Payload", () => {
  const client = new WhatsAppClient("http://x", "tok");
  const payload = {
    entry: [{
      changes: [{
        value: {
          messages: [{ id: "wamid-1", from: "+123", text: { body: "Hallo!" }, timestamp: "1234567890", type: "text" }],
          contacts: [{ profile: { name: "Max" }, wa_id: "+123" }],
        },
      }],
    }],
  };
  const msg = client.parseWebhook(payload);
  assert.notEqual(msg, null);
  assert.equal(msg.from, "+123");
  assert.equal(msg.fromName, "Max");
  assert.equal(msg.body, "Hallo!");
});

test("WhatsAppClient parseWebhook gibt null bei leerem Payload", () => {
  const client = new WhatsAppClient("http://x", "tok");
  assert.equal(client.parseWebhook({}), null);
  assert.equal(client.parseWebhook({ entry: [] }), null);
});
