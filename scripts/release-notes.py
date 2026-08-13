#!/usr/bin/env python3
# Zweck: Extrahiert die Release-Notizen einer Version aus CHANGELOG.md.
# Gibt die Sektion "## [version] - datum" (ohne Überschrift) auf stdout aus.
# Aufruf: python3 scripts/release-notes.py <version>
import re
import sys
from pathlib import Path

CHANGELOG = Path(__file__).resolve().parent.parent / "CHANGELOG.md"


def main() -> int:
    if len(sys.argv) != 2:
        print("Aufruf: release-notes.py <version>", file=sys.stderr)
        return 2
    version = sys.argv[1]
    changelog = CHANGELOG.read_text()
    m = re.search(rf"^## \[{re.escape(version)}\] - .*$", changelog, re.M)
    if not m:
        print(f"CHANGELOG-Sektion für {version} fehlt", file=sys.stderr)
        return 1
    rest = changelog[m.end():]
    end = re.search(r"^## \[", rest, re.M)
    section = rest[: end.start()] if end else rest
    print(section.strip())
    return 0


if __name__ == "__main__":
    sys.exit(main())
