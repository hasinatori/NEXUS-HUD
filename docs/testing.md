# Testing — Strategie NEXUS HUD

<!-- Zweck: Teststrategie über alle Module hinweg. Planungsstand; Details in tests/. -->

## Ebenen
1. **Unit-Tests** — je Modul, lokal im `tests/<modul>/`-Ordner.
2. **IPC-Tests** — Event-Format gegen `schema/events.schema.json` validieren (alle Module).
3. **Integrationstests** — End-to-End über den IPC-Bus (z. B. Hello-World-Handshake, Phase 1).
4. **Headless-UI-Tests** — Rendering ohne Desktop (offscreen).
5. **Dogfooding (Phase 4)** — das Team nutzt das HUD täglich selbst.

## Headless (Crostini)
- UI-/Rendering-Tests laufen headless; visuelle Prüfung numerisch (Pixel-Analyse), nicht per Blick.
- Kein Overlay-Betrieb unter Linux möglich → Overlay-Tests nur auf Windows-Zielgerät.

## CI
- CI-Workflows führen Tests je Modul aus, sobald Module Code besitzen (siehe `.github/workflows/ci.yml`).
- Secrets: niemals API-Credentials in Tests — Mock-Server verwenden (S-D).
