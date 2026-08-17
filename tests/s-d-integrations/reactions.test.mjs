// Zuletzt geaendert: 2026-08-17
// Tests fuer Cross-Module Event-Reactions.

import { describe, it, mock } from "node:test";
import assert from "node:assert/strict";

import { handleEventReaction } from "../services/s-d-integrations/reactions.ts";

function makeContext(overrides: Record<string, unknown> = {}) {
  const sent: string[] = [];
  return {
    spotify: null as unknown,
    discord: null as unknown,
    notify: (method: string, extra: Record<string, unknown> = {}) =>
      JSON.stringify({ jsonrpc: "2.0", method, params: { source: "S-D", ...extra } }),
    wsSend: (data: string) => sent.push(data),
    sent,
    ...overrides,
  };
}

describe("handleEventReaction", () => {
  it("ignoriert unbekannte Events", async () => {
    const ctx = makeContext();
    await handleEventReaction(ctx as any, "event.unknown", {});
    assert.equal(ctx.sent.length, 0);
  });

  it("reagiert auf event.build.failed mit Log", async () => {
    const logs: string[] = [];
    const origLog = console.log;
    console.log = (...args: unknown[]) => logs.push(args.join(" "));
    try {
      const ctx = makeContext();
      await handleEventReaction(ctx as any, "event.build.failed", { project: "myapp" });
      assert.ok(logs.some((l) => l.includes("Build fehlgeschlagen: myapp")));
    } finally {
      console.log = origLog;
    }
  });

  it("reagiert auf event.build.succeeded mit Log", async () => {
    const logs: string[] = [];
    const origLog = console.log;
    console.log = (...args: unknown[]) => logs.push(args.join(" "));
    try {
      const ctx = makeContext();
      await handleEventReaction(ctx as any, "event.build.succeeded", { project: "myapp" });
      assert.ok(logs.some((l) => l.includes("Build erfolgreich: myapp")));
    } finally {
      console.log = origLog;
    }
  });

  it("reagiert auf event.profile.switched mit Log", async () => {
    const logs: string[] = [];
    const origLog = console.log;
    console.log = (...args: unknown[]) => logs.push(args.join(" "));
    try {
      const ctx = makeContext();
      await handleEventReaction(ctx as any, "event.profile.switched", { profile: "gaming" });
      assert.ok(logs.some((l) => l.includes("Profil-Switch empfangen: gaming")));
    } finally {
      console.log = origLog;
    }
  });
});
