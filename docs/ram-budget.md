<!-- Zuletzt geändert: 2026-08-29 -->

# RAM-Budget — Messungen (Phase 3)

Zweck: Die Schätzungen aus `ARCHITECTURE.md` (§8) mit echten Messungen auf
Crostini/Debian (Linux) verifizieren und Einsparpotenziale dokumentieren.

## Messumgebung & Methode

- Host: ChromeOS Crostini (Debian), Linux, kein Windows-UI.
- Gemessen: residenter Speicher (RSS aus `/proc/<pid>/status`, VmRSS) nach
  10 s Laufzeit gegen den lokal laufenden Bus (Test-Port `49199`).
- Alle Services laufen wie in der Entwicklung: Bus + S-C + S-E (Go, je eigenes
  Binary) + S-D (Node.js, `dist/index.js` ohne Spotify-Daten).
- Reproduzierbar per `scripts/measure-ram.sh` (baut Go-Binaries, startet alles,
  räumt danach auf).

## Ergebnisse (29.08.2026)

| Modul | RSS | Bemerkung |
| :--- | :--- | :--- |
| `bus` (Go) | ~11 MB | Basis-Bus, wenig Code |
| `s-c-automation` (Go) | ~10 MB | Watcher + Rule-Engine |
| `s-e-monitor` (Go) | ~10 MB | Metriken + Git + Build-Log optional |
| `s-d-integrations` (Node) | ~59 MB | Node-Laufzeit ist der größte Block |
| **Summe (Linux-Runtime)** | **~90 MB** | unter dem <150-MB-Richtwert |

Nicht messbar auf Crostini: S-A (WinUI 3, ~40–80 MB geschätzt) und S-B
(Rust, ~10–30 MB geschätzt) — Schätzwerte aus ARCHITECTURE.md §8 bleiben.

## Einordnung

- Der Gesamt-Richtwert von **~150 MB** wird von der Linux-Runtime (~90 MB)
  deutlich unterschritten; selbst mit S-A + S-B liegt die Summe (der Schätzung
  nach) im Budget.
- Größter Einzelposten ist wie erwartet **S-D (Node.js)**. Hebel (nicht
  freigegeben): `NODE_OPTIONS`-Limitierungen, Lazy-Import der Integrationen.
- Bereits umgesetzt als On-Demand-Optimierung: Metriken sind per
  `cmd.metrics.set_interval` (0 = aus) abschaltbar, Git-Repos nur auf
  Anforderung (`cmd.git.watch`) — nichts pollt unnötig dauerhaft.

## Nächste Schritte (nur wenn Budget relevant wird)

- S-D-Pakete reduzieren bzw. Integrationen erst bei Aktivierung importieren.
- Go-Services bei Bedarf mit `GOGC`/`GOMEMLIMIT` begrenzen (aktuell unnötig).
- Auf Windows erneut messen (S-A/S-B einbeziehen), Werte hier nachführen.
