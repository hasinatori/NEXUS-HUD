# tests/s-a-ui-shell

<!-- Zweck: Headless-Tests (xunit) für die Linux-kompilierbaren C#-Teile der
     UI-Shell (S-A): BusClient (Verbindung/Reconnect/Keepalive), Protocol, MainViewModel. -->
<!-- Zuletzt geändert: 2026-08-16 -->

Headless-Tests ohne UI — laufen auf Linux (Crostini/CI) und Windows.

Ausführen:

```bash
cd tests/s-a-ui-shell
dotnet test
```

Abgedeckt:
- `BusClient`: Verbindungsaufbau + Hello, Auto-Reconnect nach Bus-Ausfall (neues Hello je
  Verbindung), `Disconnected`-Event nach Verbindungsende, `SendAsync` ohne Verbindung wirft,
  periodisches Heartbeat-Senden (`event.system.heartbeat`), Watchdog erkennt harte Netzabriffe
  ohne Close-Frame und reconnectet (per Abort).
- `Protocol`: Hello-/Heartbeat-Builder liefern schema-konforme Felder (`source`, `service_id`,
  `version`, `ts`).
- `MainViewModel`: Initialzustand, Zustandswechsel (`PropertyChanged`), Zähler/Text-Ableitungen.
- `WidgetLayout`/`DashboardViewModel`: JSON-Parsing (snake_case, Enum-Strings), Grid-Validierung
  (Rastergrenzen, doppelte IDs), Standard-Layout, `SetStatus`-Routing mit `PropertyChanged`.

Hinweise:
- `TestWsServer` simuliert den Bus über `HttpListener` (Close-Frame = Bus-Ausfall). Harte
  Netzabriffe (ohne Close-Frame) werden mit einem stillen Server getestet: Die Verbindung
  bleibt offen, sendet aber nichts — der Watchdog des Clients muss den Abriss erkennen und
  reconnecten.
- Das Projekt targetiert `net8.0` mit `<RollForward>LatestMajor</RollForward>` — lokal genügt
  eine neuere SDK-Version, in der CI wird `dotnet 8.0.x` installiert.
- UI-Build/Tests (XAML, Frameless, Overlay) laufen nur unter Windows.

Geplant:
- Widget-Rendering, Grid-Konfiguration, Overlay-/Snapping-Verhalten (Windows-only).
