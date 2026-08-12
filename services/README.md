# services/ — Hintergrund-Dienste

<!-- Zweck: Alle eigenständigen Services (Module) des NEXUS HUD als getrennte Unterordner. -->

| Modul | Ordner | Aufgabe | Stack (laut README) |
| :--- | :--- | :--- | :--- |
| **S-A** | `s-a-ui-shell` | Frontend / UI Shell | C# (.NET 8 / WinUI 3) oder Tauri (Rust + React) |
| **S-B** | `s-b-macro-launchpad` | Macro- & Launchpad-System | C# / C++ / Rust |
| **S-C** | `s-c-automation` | Automation Engine | Go oder Node.js / Python |
| **S-D** | `s-d-integrations` | Spotify, Discord, WhatsApp | Node.js / TypeScript |
| **S-E** | `s-e-monitor` | Coding- & Build-Monitoring | Go oder Python |

Konventionen:
- Jeder Service startet unabhängig und verbindet sich über den IPC-Bus (siehe `../ARCHITECTURE.md`, Abschnitt 3).
- Kein Service kennt einen anderen direkt — Kommunikation nur über Events.
- Der konkrete Tech-Stack eines Moduls wird festgelegt, bevor der erste Quellcode entsteht.
