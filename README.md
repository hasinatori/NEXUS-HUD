# NEXUS HUD — Desktop Command Center

<!-- Zuletzt geändert: 2026-08-14 -->

> **Das ultimative Overlay & Dashboard auf dem 2. Monitor für Devs, Gamer & Power-User.**
> *Kein trockenes Office-Tool, sondern eine Performance-Engine für Automation, Shortcuts & System-Monitoring.*

---

## Projekt-Übersicht

NEXUS HUD ist ein leichtgewichtiges, performantes Desktop-HUD (Heads-Up-Display), das permanent im Hintergrund oder auf dem Sekundärmonitor läuft. Es verbindet **System-Monitoring, Workflow-Automatisierung, Media-Steuerung und Dev-Monitoring** in einer einzigen, nahtlosen Oberfläche.

> **Status:** Phase 1 — IPC-Protokoll v1 und Dev-Bus stehen. S-B bis S-E senden `event.system.hello` (Verifikation über `scripts/hello-check`). S-E meldet zudem System-Metriken, Git-Status und Build-Ergebnisse als Events. Der Prototype-HUD zeigt das Ganze live im Terminal und verifiziert es in der CI (E2E-Job). S-A (UI) folgt.

### Kern-Features
* **Context Profiles:** Switch per Hotkey zwischen *Dev Mode*, *Gaming Mode* und *AFK/Focus Mode*.
* **Visual Automation Engine:** Hintergrund-Tasks, File-Watcher und Skripte ohne Cloud-Zwang ausführen.
* **Unified Control Hub:** Spotify, Discord, WhatsApp & System-Stats an einem Ort.
* **Dev & Build Monitor:** Git-Status, CI/CD-Pipelines und IDE-Builds auf einen Blick.

### Zielplattform
* Windows (Overlay, Global Hotkeys, Win32 API).
* Entwicklung auf ChromeOS-Crostini (Debian) möglich — dort nur reine Coding-/Test-Umgebung, kein Overlay-Test.

---

## Architektur & Tech-Stack Spec

Um maximale Performance und eine flüssige UI bei geringer CPU-Last zu garantieren, nutzen wir eine **entkoppelte Micro-Services / Modular-Architektur**.

Die Kommunikation zwischen UI und den Hintergrund-Diensten erfolgt lokal über **Named Pipes** oder **Local WebSockets (JSON-RPC)**. Details siehe [ARCHITECTURE.md](./ARCHITECTURE.md).

| Modul | Bereich | Empfohlener Tech-Stack | Grund für die Wahl |
| :--- | :--- | :--- | :--- |
| **S-A** | **Frontend / UI Shell** | **C# (.NET 8 + WinUI 3)** | Native Windows-UI, GPU-Beschleunigung, frameless Window-Support. |
| **S-B** | **Macro- & Launchpad-System** | **C# / Rust** | Tiefe OS-Integration (Global Hotkeys, Process Manager, Win32 API). |
| **S-C** | **Automation Engine** | **Go (Golang) / Python** | Extrem schnelle, nebenläufige Task-Ausführung, geringer RAM-Verbrauch. |
| **S-D** | **API Integrations** | **TypeScript / Node.js** | Perfekt für Async-Netzwerk-Requests, OAuth2 & WebSocket-Clients. |
| **S-E** | **Coding- & Build-Monitor** | **Go / Python** | Schnelles Log-Parsing, File-System Watching & Git-Cli Interfacing. |

---

### S-A: Frontend & UI-Shell

* **Lead:** Developer A
* **Tech-Stack:** `C# (.NET 8 / WinUI 3)` (Build nur auf Windows, siehe docs/dev-env.md)
* **Hauptaufgaben:**
  * [ ] Erstellung des Hauptfensters (Frameless Overlay, Snapping für 2. Monitor).
  * [ ] Dark-Mode / Cyberpunk UI Design & Grid Layout.
  * [ ] Komponenten-Bibliothek für Widgets (Gauges, Knöpfe, Status-Badges).
  * [ ] Anbindung der lokalen WebSocket/Named-Pipe-Schnittstelle zum Empfang von Events.
* **Deliverable:** Eine voll bedienbare UI, die Daten via JSON entgegennimmt und darstellt.

---

### S-B: Macro- & Launchpad-System

* **Lead:** Developer B
* **Tech-Stack:** `C# / C++ / Rust`
* **Hauptaufgaben:**
  * [ ] Global Hotkey Listener (Reagiert auch im Spiel/IDE auf Tastenkombinationen).
  * [ ] Process Launcher (Programme/Spiele fokussieren, starten oder beenden).
  * [ ] Window Manager (Fenster per Script auf bestimmte Monitore & Positionen schieben).
  * [ ] Clipboard-Manager (Historie verwalten, Code-Snippets abgreifen).
* **Deliverable:** Ein OS-Service, der Prozesse steuern und Shortcuts systemweit abfangen kann.

---

### S-C: Automation Engine

* **Lead:** Developer C
* **Tech-Stack:** `Go (Golang)` oder `Node.js / Python`
* **Hauptaufgaben:**
  * [ ] **File-Watcher System:** Überwachung von Ordnern (z.B. Downloads, Screenshots) auf Datei-Änderungen.
  * [ ] **Task-Runner Engine:** Ausführen lokaler Aktionen (z.B. Dateien konvertieren, Skripte triggern, Aufräum-Pipelines).
  * [ ] **Node/Workflow Engine Parser:** Logik zur Verarbeitung einfacher "IF THIS THEN THAT"-Regeln.
* **Deliverable:** Ein autonomer Hintergrund-Dienst, der Custom-Automationen ausführt.

---

### S-D: Integrated Apps (Spotify, Discord, WhatsApp)

* **Lead:** Developer D
* **Tech-Stack:** `Node.js / TypeScript`
* **Hauptaufgaben:**
  * [ ] **Spotify Service:** OAuth2 Auth, Play/Pause, Track-Name, Album-Art & Lautstärke via Spotify Web API.
  * [ ] **Discord Rich Presence & Bot Service:** Status-Updates und Abfangen bestimmter Trigger-Words.
  * [ ] **WhatsApp Webhook / Web-Automation:** Push-Benachrichtigungen bei bestimmten Kontakten/Wörtern auf das HUD weiterleiten.
* **Deliverable:** Ein Unified-API-Module, das Events von Drittanbieter-Services in das NEXUS HUD einspeist.

> **Hinweis:** Spotify-Lautstärke ist über die Web API nicht direkt steuerbar (nur über lokale Spotify-Instanz/Desktop-Client). WhatsApp-Web-Automation und Discord-Bots unterliegen ToS-Risiken. Siehe Risiken in [ARCHITECTURE.md](./ARCHITECTURE.md).

---

### S-E: Coding & Build Monitoring

* **Lead:** Developer E
* **Tech-Stack:** `Go` oder `Python`
* **Hauptaufgaben:**
  * [ ] **Local Git Watcher:** Auslesen von aktuellem Branch, Uncommitted Changes & Push-Status im Projektordner.
  * [ ] **Build-Log Parser:** Überwacht Compiler-Ausgaben / Terminals und schickt bei Errors ("Build Failed") ein Event an die UI.
  * [ ] **System-Metrics Gatherer:** Auslesen von CPU-Auslastung, RAM-Verbrauch, GPU-Temp via OS APIs.
  * [ ] Simple IDE-Plugin (z.B. für VS Code) zur Übertragung des aktuellen Focus/Status.
* **Deliverable:** Ein Developer-Service, der den Arbeits- und Systemzustand in Echtzeit meldet.

---

## Roadmap & Meilensteine

> Stand: **Phase 1 läuft** — IPC-Protokoll v1 und Dev-Bus funktionieren, die
> Service-Stubs senden Hello-Pings an den Bus. Versions-/Release-Prozess:
> siehe `docs/releasing.md`.

### Phase 1: Core Setup & Inter-Process Communication
* **Ziel:** Alle Architekturgrundlagen stehen, Module können miteinander sprechen.
* **Waypoints:**
  * [x] Repo-Setup & Definition des IPC-Protokolls (JSON Schema für Events).
  * [ ] **S-A:** Skeleton-UI steht auf dem Screen.
  * [x] **S-B bis S-E:** Grundlegende Services senden "Hello World"-Ping an den Bus.

### Phase 2: Core Functionality
* **Ziel:** Die wichtigsten Einzel-Features laufen isoliert.
* **Waypoints:**
  * [ ] **S-A:** Dashboard-Grid lässt sich konfigurieren.
  * [ ] **S-B:** App-Launcher & Global Hotkeys funktionieren.
  * [ ] **S-C:** Erste File-Watcher-Automatisierung läuft stabil.
  * [ ] **S-D:** Spotify-Steuerung ist voll funktionsfähig.
  * [ ] **S-E:** System-Stats (CPU/RAM) und Git-Status werden live angezeigt.

### Phase 3: Integration & Polish
* **Ziel:** Die Module greifen ineinander (Context-Switching & Triggerevents).
* **Waypoints:**
  * [ ] Profil-Switching: Hotkey schaltet UI + Apps + Hintergrund-Tasks gleichzeitig um.
  * [ ] Cross-Module Automation (z. B. *Build Failed in S-E* -> *Spiele Sound über S-D* -> *Flashe UI in S-A*).
  * [ ] Performance-Optimierung (RAM-Budget der Gesamtanwendung < 150 MB).

### Phase 4: Testing & Dogfooding
* **Ziel:** Das Team nutzt das Tool täglich selbst zum Weiter-Coden.
* **Waypoints:**
  * [ ] Bugfixing & UI/UX Fine-tuning.
  * [ ] Installer / Binary-Packaging (z. B. via InnoSetup).
  * [ ] v1.0 Release.

---

## Getting Started

### Voraussetzungen (Windows-Dev)
Je nach umgesetztem Modul:

| Modul | Benötigte Toolchain |
| :--- | :--- |
| **S-A** | .NET 8 SDK + Windows App SDK (WinUI 3) — Build nur auf Windows |
| **S-B** | Rust (siehe Phase-1-Dev-Abschnitt unten) |
| **S-C / S-E** | Go 1.22+ bzw. Python 3.11+ |
| **S-D** | Node.js 20+ |

### Phase 1 (Dev auf Crostini/Debian)
Voraussetzungen: Go 1.24+, Node.js 24+ und Rust/Cargo 1.85+ (`sudo apt-get install golang-go cargo`). Der Bus-Port ist standardmäßig `49152` und über `-port` bzw. `NEXUS_WS_PORT` änderbar.

1. **Bus starten** (separates Terminal):

   ```sh
   go run ./cmd/bus
   ```

2. **Services starten** (je ein Terminal):

   ```sh
   go run ./services/s-c-automation
   go run ./services/s-e-monitor
   cd services/s-d-integrations && npm install && npm run build && npm start
   cd services/s-b-macro-launchpad && cargo run
   ```

3. **Verifikation** — erwartet 5 `event.system.hello` (Bus + S-B bis S-E) in 15 s:

   ```sh
   go run ./scripts/hello-check
   ```

Jeder Service sendet nach dem Verbinden `event.system.hello` und wiederholt es alle 5 s, damit `hello-check` die Services unabhängig vom Verbindungszeitpunkt erkennt. Details zum Handshake: [ARCHITECTURE.md](./ARCHITECTURE.md) Abschnitt 3.3.

**Live-HUD (optional)** — zeigt Metriken, Git-Status und Build-Ergebnis an:

```sh
go run ./scripts/prototype-hud
```

Mit `-test` und `-expect` dient der Prototype-HUD als End-to-End-Prüfung
(wird in der CI automatisch ausgeführt), siehe `scripts/prototype-hud/README.md`.

### Repo-Setup (Windows-Dev)
1. Repository klonen.
2. Je Modul Toolchain installieren (siehe Tabelle oben).
3. IPC-Protokoll-Schema aus `ARCHITECTURE.md` als JSON-Schema übernehmen.

### Dokumente
* **README.md** — Projektübersicht, Module, Roadmap (diese Datei).
* **ARCHITECTURE.md** — Technische Spec: IPC, Event-Schema, Security, RAM-Budget.
* **CONTRIBUTING.md** — Regeln für Beiträge (Issues, PRs, CI-Pflichten).
* **SECURITY.md** — Sicherheitsrichtlinie und Meldeprozess.
* **CHANGELOG.md** — Änderungshistorie.
* **LICENSE** — MIT-Lizenz.
