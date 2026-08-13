# S-A — Frontend / UI Shell

<!-- Zweck: Das Hauptfenster (frameless Overlay auf dem 2. Monitor) und die komplette
     Widget-Oberfläche. Reine Darstellung — keine Geschäftslogik, keine OS-/API-Zugriffe. -->
<!-- Zuletzt geändert: 2026-08-13 -->

**Stack:** `C# (.NET 8 / WinUI 3)` — Build/Test nur auf Windows (siehe `docs/dev-env.md`)

Hauptaufgaben:
- [ ] Hauptfenster (Frameless Overlay, Snapping für 2. Monitor).
- [ ] Dark-Mode / Cyberpunk UI Design & Grid Layout.
- [ ] Widget-Bibliothek (Gauges, Knöpfe, Status-Badges).
- [ ] Anbindung der lokalen WebSocket/Named-Pipe-Schnittstelle (Empfang von Events).

## Konstrukt (Stand 2026-08-13)

Unpackaged WinUI-3-App (`.NET 8`, Windows App SDK 1.6) als `services/s-a-ui-shell`:

- `NexusHud.UI.csproj` — Projektdatei, `WindowsPackageType=None`, explizites `Program.Main`.
- `App.xaml[.cs]` — Theme (Cyberpunk Dark, Brushes unter `Nexus.*`) + App-Bootstrap.
- `MainWindow.xaml[.cs]` — Fenster: Kopfzeile (Status + Hello-Zähler), Widget-Platzhalter.
- `Bus/Protocol.cs` — JSON-RPC-2.0-Hello-Builder (Feldnamen exakt nach Schema, `ts` = RFC3339).
- `Bus/BusClient.cs` — WebSocket-Client (`ws://127.0.0.1:49152/`), Text-Frames.
- `ViewModels/MainViewModel.cs` — `INotifyPropertyChanged` für `x:Bind` (Status, Zähler, letztes Event).

Verhalten: Beim Start Verbindung zum Bus, danach Hello (`event.system.hello`, Quelle `S-A`),
anschließend Zähler + letztes Event in der UI. Reine C#-Teile (`Bus/`, `ViewModels/`)
sind Linux-kompilierbar; XAML/UI nur unter Windows baubar.

## Build & Run (Windows)

Voraussetzungen: Windows 10 1903+, Visual Studio 2022 (Workload „Desktopentwicklung mit C++“ +
`.NET Desktop Development`) oder `dotnet SDK 8`, sowie das Windows App SDK 1.6 Runtime.

```powershell
# Bus + alle Stubs starten (separate Terminals), dann:
cd services/s-a-ui-shell
dotnet build -p:Platform=x64 -p:RuntimeIdentifier=win-x64
dotnet run --no-build
```

Noch offen (nächste Schritte auf Windows): Frameless/Transparenz/Always-on-top,
Dauerhaft-Verbindung mit Reconnect, Widget-Layout, Package/Publish-Profil.

IPC:
- Sendet: `cmd.*` (Hotkeys, Launch, Automation, Media).
- Empfängt: alle `event.*`-Notifications.

Deliverable: Eine voll bedienbare UI, die Daten via JSON entgegennimmt und darstellt.
