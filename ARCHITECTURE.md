# NEXUS HUD — Architektur

<!-- Zuletzt geändert: 2026-08-13 -->

> Technische Spec zum Projekt aus `README.md`. Stand: **Phase 1** — Version
> und Versionsschema siehe `VERSION.json` bzw. `docs/releasing.md`.

---

## 1. Überblick & Ziele

NEXUS HUD verbindet **System-Monitoring, Workflow-Automatisierung, Media-Steuerung und Dev-Monitoring** in einem Desktop-HUD auf dem Sekundärmonitor.

**Ziele**
* Leichtgewichtig: flüssige UI bei niedriger CPU-Last, Gesamt-RAM-Budget **< 150 MB**.
* Entkoppelt: Module laufen als eigenständige Services, Fail fast ohne Gesamtausfall.
* Lokal & privat: keine Cloud-Pflicht für die Kernfunktionen.

**Non-Goals (bewusst nicht im Scope)**
* Kein Multi-User-/Multi-Maschinen-Sync, keine zentrale Cloud-Konsole.
* Kein general-purpose Browser oder Editor-Ersatz.
* Keine Plattform-Anbindungen, die ToS oder lokale APIs verletzen (Risiken: Abschnitt 8).

**Zielplattform:** Windows (Overlay, Global Hotkeys, Win32 API). Dev-Umgebung kann ChromeOS-Crostini (Debian) sein — dort nur Code/Test, kein Overlay-Betrieb.

---

## 2. System-Übersicht

```text
                    +-----------------------------------------------------+
                    |                     NEXUS HUD                        |
                    |                                                      |
   +-----------+    |   +----------+     +-----------+    +-------------+  |
   |  S-A      |<------>|          |<--->|  S-C      |    |  S-D        |  |
   | UI Shell  |    |   |          |     | Automation|    | Integrations|  |
   | (WinUI 3) |    |   |  IPC-Bus |<--->| Engine    |    | (Spotify,   |  |
   |  .NET 8)  |    |   |          |     +-----------+    |  Discord,   |  |
   +-----------+    |   |          |<--->+-------------+  |  WhatsApp)  |  |
   +-----------+    |   |          |     |  S-B        |  +-------------+  |
   | Hotkeys / |<------>|          |<--->|  Macro &    |                   |
   | Prozesse  |    |   +----------+     |  Launchpad  |                   |
   | (Userland)|    |                    +-------------+                   |
   +-----------+    |   +-------------+                                    |
                    |   |  S-E        |                                    |
                    |   | Coding/Build|                                    |
                    |   | Monitor     |                                    |
                    |   +-------------+                                    |
                    +-----------------------------------------------------+
        (Alle Pfeile = lokale Kommunikation über den IPC-Bus)
```

**Prinzipien**
* Die **UI (S-A)** ist reine Zustands-Darstellung — keine Geschäftslogik, keine direkten OS-/API-Zugriffe.
* Jeder Service (S-B…S-E) ist unabhängig startbar; Ausfall eines Services degradiert, crasht aber nie die UI.
* Zentraler **IPC-Bus** (Abschnitt 3) entkoppelt Sender und Empfänger; Module kennen sich nicht gegenseitig.

---

## 3. Kommunikation & IPC

### 3.1 Transport

| Transport | Rolle | Begründung |
| :--- | :--- | :--- |
| **Named Pipes** | Ziel (Windows-Release) | Nativer Windows-IPC, schnell, keine Port-Kollisionen, Zugriffskontrolle über DACL. |
| **Local WebSocket** (127.0.0.1) | **Phase 1 / Dev** | Einfach in Go/Rust/Node zu nutzen, JSON-nativ, auf Crostini testbar. |

* Named Pipe-Name: `\\.\pipe\nexus-hud`
* WebSocket-Endpoint: `ws://127.0.0.1:<port>/` (Port konfigurierbar, Standard z.B. 49152–49162-Bereich).

**Phase-1-Entscheidung:** Der Hello-World-Handshake und die Service-Stubs laufen zunächst über **Local WebSocket** — damit ist alles auf Crostini (Linux) entwickel- und testbar. Named Pipes werden mit der S-A/S-B-Windows-Phase (Overlay, Win32) ergänzt; das Protokoll bleibt transportunabhängig identisch.

### 3.2 Protokoll: JSON-RPC 2.0

Alle Nachrichten folgen [JSON-RPC 2.0](https://www.jsonrpc.org/specification):

```json
{
  "jsonrpc": "2.0",
  "id": "msg-<uuid>",
  "method": "event.system.metrics",
  "params": {
    "source": "S-E",
    "cpu": 34.2,
    "ram": { "used_mb": 6400, "total_mb": 16384 },
    "gpu_temp_c": 61
  }
}
```

* **Requests** (`id` gesetzt) — Aufruf mit Antwort.
* **Notifications** (`id` fehlt) — Feuer-und-vergessen, Standard für Events.
* **Konvention:** Methodennamen als `event.<domain>.<action>` bzw. `cmd.<domain>.<action>`.

### 3.3 Event-Katalog (Standard-Schema)

| Event | Richtung | Bedeutung |
| :--- | :--- | :--- |
| `event.system.hello` | Service → UI | Service ist gestartet (Handshake nach Connect). |
| `event.system.heartbeat` | Service → UI | Lebendig-Meldung (Interval, z.B. 15 s). |
| `event.system.metrics` | S-E → UI | CPU/RAM/GPU-Metriken. |
| `event.build.failed` | S-E → UI | Build-Fehler erkannt. |
| `event.build.succeeded` | S-E → UI | Build erfolgreich. |
| `event.git.status` | S-E → UI | Branch/Uncommitted/Push-Status. |
| `event.media.state` | S-D → UI | Track, Album-Art, Play/Pause, Lautstärke. |
| `event.presence.changed` | S-D → UI | Discord/WhatsApp-Status-Trigger. |
| `event.profile.switched` | IPC-Bus → alle | Context-Profile gewechselt (Dev/Gaming/AFK). |
| `cmd.media.toggle` | UI → S-D | Play/Pause-Kommando. |
| `cmd.app.launch` | UI → S-B | Programm/Spiel starten oder fokussieren. |
| `cmd.hotkey.register` | UI → S-B | Globalen Hotkey registrieren. |
| `cmd.automation.run` | UI → S-C | Automation/Task starten. |

**Hello-World-Definition (Phase 1):** Jeder Service sendet nach Connect
`event.system.hello` mit `{ "source": "S-B", "service_id": "s-b-macro-launchpad", "version": "0.1.0", "ts": "<ISO8601>" }` — die UI zeigt dies als Status-Badge an.
Der Bus sendet nach jedem Connect selbst ein Hello mit `source: "bus"` und `service_id: "bus"`.
In der Phase 1 wiederholen die Service-Stubs das Hello alle 5 s, damit `scripts/hello-check` alle Teilnehmer unabhängig vom Verbindungszeitpunkt erkennt.

### 3.4 Protokoll-Versionierung

* Jede Nachricht trägt `protocol_version` im `params`-Objekt (aktuell `1`).
* Mismatch (`protocol_version` != erwartet) wird vom Empfänger mit dem Fehlerformat (Abschnitt 3.5) beantwortet und die Verbindung geschlossen.
* Das JSON-Schema in `schema/events.schema.json` ist die single source of truth; Änderungen am Event-Katalog erhöhen die `protocol_version`.

### 3.5 Fehlerformat

Antwort/Notification bei Fehlern, JSON-RPC-2.0-konform:

```json
{
  "jsonrpc": "2.0",
  "method": "error.protocol",
  "params": {
    "protocol_version": 1,
    "source": "S-A",
    "code": -32602,
    "message": "Ungültiges Event: unbekannte Methode",
    "ts": "2026-08-13T10:00:00Z"
  }
}
```

* Fehlercodes: JSON-RPC-Standardcodes (`-32600` Invalid Request, `-32601` Method not found, `-32602` Invalid params).
* `error.protocol` wird als Notification gesendet (keine `id`).

---

## 4. Security

* **Bind nur auf Loopback:** Named Pipes lokal bzw. WebSocket nur auf `127.0.0.1` — niemals auf `0.0.0.0`/`::`.
* **Origin-Check (WebSocket):** Server validiert `Origin`-Header; nur lokale UI-Origin wird akzeptiert (Schutz vor Browser-CSRF/XSS).
* **Auth-Token:** Pro Sitzung generiertes Token (in Pipe-/WS-Connect-Handshake ausgetauscht) verhindert Fremd-Clients.
* **Pipe-DACL:** Named Pipe wird mit restriktiver Zugriffssteuerung erstellt (nur aktueller User).
* **Eingangsvalidierung:** Alle Events werden gegen das JSON-Schema validiert, bevor sie weitergegeben werden; unbekannte Methoden werden verworfen.
* **Keine Secrets im Klartext:** OAuth2-Tokens (S-D) verschlüsselt im User-Profil speichern (Windows DPAPI bzw. `keytar`).

---

## 5. Module (S-A bis S-E)

Technologie je Modul wie in `README.md` spezifiziert (inkl. Alternativen).

### S-A — Frontend / UI Shell

* **Stack:** `C# (.NET 8 / WinUI 3)` — Build/Test nur auf Windows (Dev-Env: docs/dev-env.md)
* **Verantwortung:** Frameless Overlay, Grid-Layout, Widget-Bibliothek, Darstellung aller Events.
* **Sendet:** `cmd.*` (Hotkeys, Launch, Automation, Media).
* **Empfängt:** alle `event.*`-Notifications.

### S-B — Macro- & Launchpad-System

* **Stack:** `C# / C++ / Rust`
* **Verantwortung:** Global Hotkeys, Process Launcher, Window Manager, Clipboard-Manager.
* **Sendet:** `event.hotkey.triggered`, `event.process.started`, `event.window.moved`.
* **Empfängt:** `cmd.hotkey.register`, `cmd.app.launch`, `cmd.window.move`.

### S-C — Automation Engine

* **Stack:** `Go` oder `Node.js / Python`
* **Verantwortung:** File-Watcher, Task-Runner, IF-THIS-THEN-THAT-Regel-Engine.
* **Sendet:** `event.automation.started`, `event.automation.finished`, `event.file.changed`.
* **Empfängt:** `cmd.automation.run`, `event.profile.switched` (aktive Regeln je Profil).

### S-D — Integrated Apps

* **Stack:** `Node.js / TypeScript`
* **Verantwortung:** Spotify, Discord, WhatsApp → Unified-Events.
* **Sendet:** `event.media.state`, `event.presence.changed`.
* **Empfängt:** `cmd.media.toggle`, `cmd.media.next`, `cmd.media.volume`.

### S-E — Coding & Build Monitoring

* **Stack:** `Go` oder `Python`
* **Verantwortung:** Git-Watcher, Build-Log-Parser, System-Metriken, IDE-Status.
* **Sendet:** `event.system.metrics`, `event.git.status`, `event.build.failed`, `event.build.succeeded`, `event.ide.focus`.
* **Empfängt:** `cmd.metrics.set_interval`, `cmd.git.watch` (Pfad registrieren).

---

## 6. Cross-Module-Beispiel

**Szenario "Build Failed → Sound → UI-Flash" (Phase 3):**

```text
S-E  Build-Log-Parser erkennt "Build Failed"
  └── event.build.failed  → IPC-Bus
        ├── S-A  → UI flasht rot (Widget "Build")
        └── S-D  → spielt kurzen Warnton ab (lokaler Sound, kein Spotify)
```

**Regelwerk (S-C):** Der eigentliche Trigger-Logik liegt als Automation-Regel vor:
`IF event.build.failed AND profile=Dev THEN (flash UI) + (play sound)`.
S-C wertet Events aus und stößt Aktionen an — Module bleiben entkoppelt.

---

## 7. Context Profiles

| Profil | UI | Hotkeys (S-B) | Hintergrund-Tasks (S-C) | Beispiele |
| :--- | :--- | :--- | :--- | :--- |
| **Dev Mode** | Git/Build-Widgets im Fokus | Editor-Shortcuts | Build-Pipelines, Watcher aktiv | VS Code, Terminal |
| **Gaming Mode** | HUD reduziert, minimale Overlays | Game-Specific Keys | Performance-Logging | Steam, Spiele |
| **AFK/Focus Mode** | Ausgeblendet, Nur-Störungsmelder | Pause-Stummschaltung | Media-Auto-Pause | Nichtstun/Fokus |

* **Auslöser:** Globaler Hotkey (S-B) → Event `event.profile.switched` → jeder Service reagiert eigenständig.

---

## 8. RAM-Budget (Richtwert)

Kein hartes Ziel: ~150 MB Gesamtverbrauch als Richtwert, erst in Phase 3 mit
echten Messungen verifiziert. Schätzung je Runtime:

| Komponente | Ungefährer Speicher | Hinweis |
| :--- | :--- | :--- |
| UI S-A (WinUI 3) | 40–80 MB | Frameless Overlay + Widgets |
| S-B (Rust, Service) | 10–30 MB | Rust sparsamer als C# |
| S-C (Go) | 10–15 MB | Go sehr leicht |
| S-D (Node.js) | 40–80 MB | Node-Basis ist der größte Block |
| S-E (Go/Python) | 10–30 MB | Python ~30 MB, Go ~10 MB |
| **Summe** | **110–235 MB** | **Richtwert, nicht bindend** |

**Einordnung:** Der Verbrauch hängt am stärksten von S-D (Node.js) ab; ein
hartes <150-MB-Ziel ist aufgegeben. Messungen erfolgen in Phase 3, ggf. mit
Anpassungen (z. B. Metriken nur auf Anforderung statt Dauer-Polling).

---

## 9. Risiken & Offene Punkte

| Risiko | Details | Mitigation |
| :--- | :--- | :--- |
| **Spotify-Lautstärke** | Web API unterstützt keine Volume-Steuerung. | Lokale Spotify-Instanz-API / Spotify-Connect; Volume-Widget nur anzeigen. |
| **WhatsApp-Web-Automation** | ToS-Risiko, fragile DOM-Selektoren. | Nur Push-Benachrichtigungen via offiziellen Webhooks; Automations-Teil optional (Feature-Flag). |
| **Discord-Bot** | Self-Bot verstößt gegen Discord-ToS. | Nur offizielle Bot-API (User-Account-Features vermeiden). |
| **Toolchain-Pflege (5 Stacks)** | 5 Sprachen = 5 Build-Systeme + Dependencies. | Module so klein halten, dass Abhängigkeiten minimal bleiben; Fix auf ein Tool-Doku (dieses File). |
| **Overlay-Test auf Crostini** | Linux/Wayland ≠ Windows-Overlay. | Overlay-Tests nur auf Windows-Zielgerät; Crostini nur für Service-Unit-Tests. |
| **JSON-Schema-Drift** | Events divergieren zwischen Modulen. | Zentrales Schema-Verzeichnis im Repo + Validierung beim Start (Phase 1). |

---

## 10. Roadmap-Detail

Siehe `README.md` für den Überblick. Hier die technischen Deliverables:

**Phase 1: Grundgerüst & IPC**
- Repo-Struktur, CI (Build je Modul), JSON-Schema-Verzeichnis.
- S-A: Fenster auf Bildschirm (Skeleton, Dark-Theme-Basis).
- S-B–S-E: Je ein Service mit `event.system.hello`-Handshake.

**Phase 2: Kernfeatures isoliert**
- S-A: konfigurierbares Dashboard-Grid.
- S-B: Launcher + erste Global Hotkeys.
- S-C: ein stabiler File-Watcher mit Regel.
- S-D: Spotify Play/Pause + Track-Info.
- S-E: CPU/RAM-Metriken + Git-Status-Anzeige.

**Phase 3: Integration & Polish**
- `event.profile.switched` über den Bus wirksam.
- Cross-Module-Automation (Abschnitt 6).
- RAM-Messung und Optimierung gegen das <150-MB-Budget.

**Phase 4: Testing & Dogfooding**
- Bugfixes, UX-Feintuning.
- Installer/Packaging (InnoSetup).
- v1.0 Release.

---

*Ergänzungen/Änderungen an dieser Spec gehen über die Module-Sektionen — das JSON-Schema in Abschnitt 3 ist die single source of truth für die IPC-Verträge.*
