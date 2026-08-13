#!/usr/bin/env python3
# Zweck: Projekt- und Modulversionen erhöhen. Quell der Wahrheit ist
# VERSION.json; abgeleitete Dateien werden synchronisiert:
#   - shared/version/version.go (Bus, S-C, S-E)
#   - services/s-b-macro-launchpad/Cargo.toml
#   - services/s-d-integrations/package.json
# CHANGELOG.md: [Unreleased]-Einträge werden in eine neue Version-Sektion
# übernommen (Keep a Changelog).
#
# Aufruf:
#   python3 scripts/bump-version.py --next <patch|minor|major>
#       gibt nur die nächste Projektversion aus (schreibt nichts)
#   python3 scripts/bump-version.py <patch|minor|major> [module ...]
#       führt den Bump aus; module aus VERSION.json (bus, s-b, s-c, s-d, s-e)
#       werden jeweils um 1 erhöht. Läuft NICHT auf main.
import json
import re
import subprocess
import sys
from datetime import date
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

VERSION_JSON = ROOT / "VERSION.json"
VERSION_GO = ROOT / "shared/version" / "version.go"
CARGO_TOML = ROOT / "services" / "s-b-macro-launchpad" / "Cargo.toml"
PACKAGE_JSON = ROOT / "services" / "s-d-integrations" / "package.json"
CHANGELOG = ROOT / "CHANGELOG.md"

GO_TEMPLATE = """// Package version hält die zentrale Projekt- und Modulversion.
// Quell der Wahrheit ist die VERSION.json im Repo-Root; diese Datei wird
// von scripts/bump-version.py synchronisiert und von check-version.py geprüft.
package version

const (
\tProject = "{proj}"

\tBus          = "{proj}-bus.{bus}"
\tSCAutomation = "{proj}-s-c.{sc}"
\tSEMonitor    = "{proj}-s-e.{se}"
)
"""


def die(msg: str) -> None:
    print(f"FEHLER: {msg}", file=sys.stderr)
    sys.exit(1)


def next_version(current: str, kind: str) -> str:
    maj, mid, pat = (int(p) for p in current.split("."))
    if kind == "major":
        return f"{maj + 1}.0.0"
    if kind == "minor":
        return f"{maj}.{mid + 1}.0"
    if kind == "patch":
        return f"{maj}.{mid}.{pat + 1}"
    die(f"unbekannter Bump-Typ '{kind}' (patch|minor|major)")


def main() -> int:
    if len(sys.argv) < 2:
        print("Aufruf: bump-version.py --next <patch|minor|major> | bump-version.py <patch|minor|major> [module ...]")
        return 2

    vj = json.loads(VERSION_JSON.read_text())
    current = vj.get("version")
    if not re.match(r"^\d+\.\d+\.\d+$", current or ""):
        die(f"VERSION.json: '{current}' ist kein gültiges MAJOR.MINOR.PATCH")
    modules = vj.get("modules", {})
    known = set(modules)

    if sys.argv[1] == "--next":
        if len(sys.argv) != 3:
            die("--next braucht genau einen Bump-Typ")
        print(next_version(current, sys.argv[2]))
        return 0

    kind = sys.argv[1]
    bump_mods = sys.argv[2:]
    unknown = [m for m in bump_mods if m not in known]
    if unknown:
        die(f"unbekannte Module: {', '.join(unknown)} (bekannt: {', '.join(sorted(known))})")

    branch = subprocess.run(["git", "branch", "--show-current"], capture_output=True, text=True).stdout.strip()
    if branch == "main":
        die("Bump nur auf einem Release-Branch ausführen (main ist geschützt)")

    new_proj = next_version(current, kind)
    new_modules = dict(modules)
    for m in bump_mods:
        new_modules[m] += 1

    changelog = CHANGELOG.read_text()
    if "## [Unreleased]" not in changelog:
        die("CHANGELOG.md enthält keine [Unreleased]-Sektion")
    header, _sep, rest = changelog.partition("## [Unreleased]")
    content, _, after = rest.lstrip("\n").partition("\n## [")
    if content.startswith("## ["):
        content, after = "", rest.lstrip("\n")
    m = re.search(r"^### Planned \(Roadmap\)\s*$", content, re.M)
    if m:
        release_content, roadmap = content[: m.start()].strip(), content[m.start() :].rstrip()
    else:
        release_content, roadmap = content.strip(), ""
    if not release_content:
        die("CHANGELOG.md: [Unreleased]-Sektion enthält keine Änderungen")
    today = date.today().isoformat()
    new_section = f"## [{new_proj}] - {today}\n\n{release_content}\n"
    blocks = ["## [Unreleased]"]
    if roadmap:
        blocks.append(roadmap)
    blocks.append(new_section.strip())
    if after:
        blocks.append(("## [" + after).strip())
    new_changelog = header + "\n\n".join(blocks) + "\n"

    VERSION_JSON.write_text(json.dumps({"version": new_proj, "modules": new_modules}, indent=2) + "\n")
    VERSION_GO.write_text(
        GO_TEMPLATE.format(proj=new_proj, bus=new_modules["bus"], sc=new_modules["s-c"], se=new_modules["s-e"])
    )
    cargo = CARGO_TOML.read_text()
    CARGO_TOML.write_text(re.sub(r"^version\s*=\s*\"[^\"]+\"", f'version = "{new_proj}-s-b.{new_modules["s-b"]}"', cargo, count=1, flags=re.M))
    pj = json.loads(PACKAGE_JSON.read_text())
    pj["version"] = f"{new_proj}-s-d.{new_modules['s-d']}"
    PACKAGE_JSON.write_text(json.dumps(pj, indent=2) + "\n")
    CHANGELOG.write_text(new_changelog)

    mods_str = ", ".join(f"{m}.{new_modules[m]}" for m in sorted(new_modules))
    print(f"Neue Projektversion: {new_proj} (Module: {mods_str})")
    print("CHANGELOG und abgeleitete Versionen aktualisiert. Nun: Commit, PR erstellen, merge.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
