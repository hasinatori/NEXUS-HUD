#!/usr/bin/env python3
# Zweck: Prüft die Version-Konsistenz des Repos. Quell der Wahrheit ist
# VERSION.json; die daraus abgeleiteten Modulversionen müssen in
# shared/version/version.go, Cargo.toml (S-B) und package.json (S-D)
# exakt übereinstimmen. CHANGELOG.md muss im Keep-a-Changelog-Format sein.
# Lauf in CI (Job "Version-Konsistenz") und lokal: python3 scripts/check-version.py
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

SEMVER = re.compile(r"^\d+\.\d+\.\d+$")


def fail(msg: str) -> None:
    print(f"FEHLER: {msg}")


def go_const(text: str, name: str) -> str | None:
    m = re.search(rf"{name}\s*=\s*\"([^\"]+)\"", text)
    return m.group(1) if m else None


def main() -> int:
    ok = True

    vj = json.loads((ROOT / "VERSION.json").read_text())
    proj = vj.get("version")
    if not SEMVER.match(proj or ""):
        ok = False
        fail(f"VERSION.json: '{proj}' ist kein gültiges MAJOR.MINOR.PATCH")
        return 1
    modules = vj.get("modules", {})
    expected = {m: f"{proj}-{m}.{modules[m]}" for m in modules}

    checks = []

    go_text = (ROOT / "shared/version/version.go").read_text()
    checks.append(("shared/version: Project", go_const(go_text, "Project"), proj))
    for const_name, mod in (("Bus", "bus"), ("SCAutomation", "s-c"), ("SEMonitor", "s-e")):
        checks.append((f"shared/version: {const_name}", go_const(go_text, const_name), expected[mod]))

    cargo = (ROOT / "services" / "s-b-macro-launchpad" / "Cargo.toml").read_text()
    m = re.search(r"^version\s*=\s*\"([^\"]+)\"", cargo, re.M)
    checks.append(("Cargo.toml (S-B)", m.group(1) if m else None, expected["s-b"]))

    pj = json.loads((ROOT / "services" / "s-d-integrations" / "package.json").read_text())
    checks.append(("package.json (S-D)", pj.get("version"), expected["s-d"]))

    for name, actual, want in checks:
        if actual != want:
            ok = False
            fail(f"{name}: erwartet {want}, gefunden {actual}")

    changelog = (ROOT / "CHANGELOG.md").read_text()
    if "## [Unreleased]" not in changelog:
        ok = False
        fail("CHANGELOG.md enthält keine [Unreleased]-Sektion")

    if ok:
        modules_str = ", ".join(f"{k}.{v}" for k, v in modules.items())
        print(f"Version konsistent: {proj} (Module: {modules_str})")
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
