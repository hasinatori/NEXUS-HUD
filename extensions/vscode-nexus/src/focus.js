// Zuletzt geaendert: 2026-08-28
// Reine, testbare Fokus-Logik der NEXUS IDE-Bridge (kein vscode-Import).

"use strict";

const os = require("node:os");
const path = require("node:path");
const fs = require("node:fs");

/**
 * Standardpfad der Focus-Datei (~/.nexus/ide-focus.json).
 */
function defaultFocusPath(homeDir = os.homedir()) {
  return path.join(homeDir, ".nexus", "ide-focus.json");
}

/**
 * Loest eine konfigurierte Focus-Datei auf (leer = Standard, "~" erweitern).
 */
function resolveFocusPath(input, homeDir = os.homedir()) {
  if (!input) {
    return defaultFocusPath(homeDir);
  }
  if (input === "~") {
    return homeDir;
  }
  if (input.startsWith("~/")) {
    return path.join(homeDir, input.slice(2));
  }
  return input;
}

/**
 * Baut den Fokus-Payload aus Editor-/Workspace-Infos.
 * Liefert null, wenn keine Datei aktiv ist.
 *
 * editor: { fileName, languageId, workspaceFolders: [{ name, uri: { fsPath } }] }
 */
function buildFocusInput(editor) {
  if (!editor || !editor.fileName) {
    return null;
  }
  const folder = editor.workspaceFolders && editor.workspaceFolders[0];
  return {
    project: (folder && folder.name) || "untitled",
    filename: path.basename(editor.fileName),
    language: editor.languageId || "plaintext",
    path: editor.fileName,
    ts: new Date().toISOString(),
  };
}

/**
 * Schreibt den Fokus-Payload in die Focus-Datei (legt das Verzeichnis an).
 * Gibt den Pfad zurueck oder null bei leerem Payload. Datei-Funktionen sind
 * fuer Tests injizierbar.
 */
function writeFocusFile(focusPath, payload, extra = {}) {
  if (!payload) {
    return null;
  }
  const mkdirSync = extra.mkdirSync || fs.mkdirSync;
  const writeFileSync = extra.writeFileSync || fs.writeFileSync;
  mkdirSync(path.dirname(focusPath), { recursive: true });
  writeFileSync(focusPath, JSON.stringify(payload, null, 2) + "\n", "utf8");
  return focusPath;
}

module.exports = { defaultFocusPath, resolveFocusPath, buildFocusInput, writeFocusFile };