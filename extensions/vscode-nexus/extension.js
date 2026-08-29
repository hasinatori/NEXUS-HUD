// Zuletzt geaendert: 2026-08-28
// NEXUS HUD IDE-Bridge (VS Code): schreibt die aktive Datei als JSON in die
// Focus-Datei, die der S-E-Monitor (services/s-e-monitor, Flag -ide-focus)
// auswertet und als event.ide.focus an den Bus sendet.

"use strict";

const vscode = require("vscode");
const {
  resolveFocusPath,
  buildFocusInput,
  writeFocusFile,
} = require("./src/focus");

let statusBar;

function activate(context) {
  const focusPath = resolveFocusPath(
    vscode.workspace.getConfiguration("nexus").get("focusFile", ""),
  );
  console.log(`[nexus] Focus-Datei: ${focusPath}`);

  statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
  statusBar.text = "$(book) NEXUS";
  statusBar.command = "nexus.ide.ping";
  statusBar.show();
  context.subscriptions.push(statusBar);

  context.subscriptions.push(
    vscode.commands.registerCommand("nexus.ide.ping", () => {
      const payload = buildFocusInput(activeEditorInfo());
      void vscode.window.showInformationMessage(
        payload ? `NEXUS: ${payload.project} — ${payload.filename}` : "NEXUS: keine aktive Datei",
      );
    }),
  );

  const pushFocus = () => {
    try {
      const payload = buildFocusInput(activeEditorInfo());
      writeFocusFile(focusPath, payload);
      statusBar.text = payload ? `$(book) ${payload.filename}` : "$(book) NEXUS";
    } catch (err) {
      console.error(`[nexus] Focus-Datei schreiben fehlgeschlagen: ${err.message}`);
    }
  };

  pushFocus();

  context.subscriptions.push(
    vscode.window.onDidChangeActiveTextEditor(pushFocus),
    vscode.window.onDidChangeWindowState(pushFocus),
  );
}

function activeEditorInfo() {
  const ed = vscode.window.activeTextEditor;
  if (!ed || !ed.document) {
    return null;
  }
  return {
    fileName: ed.document.fileName,
    languageId: ed.document.languageId,
    workspaceFolders: (vscode.workspace.workspaceFolders || []).map((f) => ({
      name: f.name,
      uri: { fsPath: f.uri.fsPath },
    })),
  };
}

function deactivate() {}

module.exports = { activate, deactivate };