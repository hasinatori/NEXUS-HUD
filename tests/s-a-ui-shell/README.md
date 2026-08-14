# tests/s-a-ui-shell

<!-- Zweck: Headless-Tests (xunit) für die Linux-kompilierbaren C#-Teile der
     UI-Shell (S-A): BusClient (Verbindung/Reconnect), Protocol, MainViewModel. -->
<!-- Zuletzt geändert: 2026-08-14 -->

Headless-Tests ohne UI — laufen auf Linux (Crostini/CI) und Windows.

Ausführen:
```bash
cd tests/s-a-ui-shell
dotnet test
```

Abgedeckt:
- `BusClient`: Verbindungsaufbau + Hello, Auto-Reconnect nach Bus-Ausfall (neues Hello je
  Verbindung), `Disconnected`-Event nach Verbindungsende, `SendAsync` ohne Verbindung wirft.
- `Protocol`: Hello-Builder liefert schema-konforme Felder (`source`, `service_id`, `version`, `ts`).
- `MainViewModel`: Initialzustand, Zustandswechsel (`PropertyChanged`), Zähler/Text-Ableitungen.

Hinweise:
- `TestWsServer` simuliert den Bus über `HttpListener` (Close-Frame = Bus-Ausfall). Auf Linux
  erkennt der Client echte `Abort()`/Netzabrisse nicht prompt — der Reconnect-Weg wird daher
  über den Close-Frame getestet; Keepalive/Heartbeat folgt mit Phase 2.
- Das Projekt targetiert `net8.0` mit `<RollForward>LatestMajor</RollForward>` — lokal genügt
  eine neuere SDK-Version, in der CI wird `dotnet 8.0.x` installiert.
- UI-Build/Tests (XAML, Frameless, Overlay) laufen nur unter Windows.

Geplant:
- Widget-Rendering, Grid-Konfiguration, Overlay-/Snapping-Verhalten (Windows-only).
