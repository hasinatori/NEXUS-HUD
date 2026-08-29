// Zuletzt geaendert: 2026-08-28
// Headless-Tests fuer die Fokus-Logik der NEXUS IDE-Bridge (ohne VSCode-Runtime).

"use strict";

const { test } = require("node:test");
const assert = require("node:assert/strict");
const path = require("node:path");
const fs = require("node:fs");
const os = require("node:os");

const {
  defaultFocusPath,
  resolveFocusPath,
  buildFocusInput,
  writeFocusFile,
} = require("../src/focus");

test("defaultFocusPath nutzt ~/.nexus/ide-focus.json", () => {
  assert.equal(defaultFocusPath("/home/tester"), path.join("/home/tester", ".nexus", "ide-focus.json"));
});

test("resolveFocusPath: leer -> Standard, ~ -> Home, ~/ -> Home, sonst Pfad", () => {
  const home = "/home/tester";
  assert.equal(resolveFocusPath("", home), path.join(home, ".nexus", "ide-focus.json"));
  assert.equal(resolveFocusPath("~", home), home);
  assert.equal(resolveFocusPath("~/custom/focus.json", home), path.join(home, "custom", "focus.json"));
  assert.equal(resolveFocusPath("/abs/pfad/focus.json", home), "/abs/pfad/focus.json");
});

test("buildFocusInput baut Payload aus Editor-Infos", () => {
  const editor = {
    fileName: "/work/NEXUS/services/s-e-monitor/main.go",
    languageId: "go",
    workspaceFolders: [{ name: "NEXUS-HUD", uri: { fsPath: "/work/NEXUS" } }],
  };
  const payload = buildFocusInput(editor);
  assert.equal(payload.project, "NEXUS-HUD");
  assert.equal(payload.filename, "main.go");
  assert.equal(payload.language, "go");
  assert.equal(payload.path, "/work/NEXUS/services/s-e-monitor/main.go");
  assert.ok(!Number.isNaN(Date.parse(payload.ts)), "ts muss ISO-8601 sein");
});

test("buildFocusInput liefert null ohne Datei", () => {
  assert.equal(buildFocusInput(null), null);
  assert.equal(buildFocusInput({ languageId: "go" }), null);
});

test("buildFocusInput defaults fuer untitled/plaintext ohne Workspace", () => {
  const payload = buildFocusInput({ fileName: "/tmp/x.txt", languageId: "" });
  assert.equal(payload.project, "untitled");
  assert.equal(payload.language, "plaintext");
  assert.equal(payload.filename, "x.txt");
});

test("writeFocusFile legt Verzeichnis an und schreibt JSON", () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "nexus-focus-"));
  const file = path.join(dir, "nested", "ide-focus.json");
  const payload = { project: "P", filename: "f.go", language: "go", path: "/tmp/f.go", ts: "2026-08-28T10:00:00Z" };
  const written = writeFocusFile(file, payload);
  assert.equal(written, file);
  assert.equal(JSON.parse(fs.readFileSync(file, "utf8")).filename, "f.go");
});

test("writeFocusFile: null-Payload schreibt nichts", () => {
  const calls = [];
  const written = writeFocusFile("/x/focus.json", null, {
    mkdirSync: () => calls.push("mkdir"),
    writeFileSync: () => calls.push("write"),
  });
  assert.equal(written, null);
  assert.equal(calls.length, 0);
});