// Zuletzt geaendert: 2026-08-28
// Tests fuer Cross-Module Event-Reactions.

import { describe, it } from "node:test";
import assert from "node:assert/strict";

import { handleEventReaction } from "../../services/s-d-integrations/reactions.ts";

function makeContext(overrides = {}) {
  const sent = [];
  return {
    spotify: null,
    discord: null,
    notify: (method, extra = {}) =>
      JSON.stringify({ jsonrpc: "2.0", method, params: { source: "S-D", ...extra } }),
    wsSend: (data) => sent.push(data),
    sent,
    ...overrides,
  };
}

describe("handleEventReaction", () => {
  it("ignoriert unbekannte Events", async () => {
    const ctx = makeContext();
    await handleEventReaction(ctx, "event.unknown", {});
    assert.equal(ctx.sent.length, 0);
  });

  it("reagiert auf event.build.failed mit Log", async () => {
    const logs = [];
    const origLog = console.log;
    console.log = (...args) => logs.push(args.join(" "));
    try {
      const ctx = makeContext();
      await handleEventReaction(ctx, "event.build.failed", { project: "myapp" });
      assert.ok(logs.some((l) => l.includes("Build fehlgeschlagen: myapp")));
    } finally {
      console.log = origLog;
    }
  });

  it("reagiert auf event.build.succeeded mit Log", async () => {
    const logs = [];
    const origLog = console.log;
    console.log = (...args) => logs.push(args.join(" "));
    try {
      const ctx = makeContext();
      await handleEventReaction(ctx, "event.build.succeeded", { project: "myapp" });
      assert.ok(logs.some((l) => l.includes("Build erfolgreich: myapp")));
    } finally {
      console.log = origLog;
    }
  });

  it("reagiert auf event.profile.switched mit Log", async () => {
    const logs = [];
    const origLog = console.log;
    console.log = (...args) => logs.push(args.join(" "));
    try {
      const ctx = makeContext();
      await handleEventReaction(ctx, "event.profile.switched", { profile: "gaming" });
      assert.ok(logs.some((l) => l.includes("Profil-Switch empfangen: gaming")));
    } finally {
      console.log = origLog;
    }
  });
});