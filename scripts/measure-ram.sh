#!/usr/bin/env bash
# Zuletzt geändert: 2026-08-29
# Misst den RAM-Verbrauch (RSS) der auf Linux lauffaehigen NEXUS-Module
# (Phase 3, siehe docs/ram-budget.md). Baut die Go-Binaerien, startet alles
# auf einem Test-Port und beendet die Prozesse am Ende.

set -u

PORT="${NEXUS_RAM_PORT:-49199}"
OUT="$(mktemp -d)"
BIN="$OUT/bin"
mkdir -p "$BIN"

echo "==> Go-Builds"
go build -o "$BIN/" ./cmd/bus ./services/s-c-automation ./services/s-e-monitor ./scripts/prototype-hud

echo "==> Start auf Port $PORT"
"$BIN/bus" -port "$PORT" >"$OUT/bus.log" 2>&1 &
BUS=$!
sleep 1
"$BIN/s-c-automation" -port "$PORT" >"$OUT/sc.log" 2>&1 &
SC=$!
"$BIN/s-e-monitor" -port "$PORT" -git-dir "$(git rev-parse --show-toplevel)" >"$OUT/se.log" 2>&1 &
SE=$!
bash -c "cd services/s-d-integrations && exec env NEXUS_WS_PORT=$PORT node dist/index.js" >"$OUT/sd.log" 2>&1 &
SD=$!

echo "==> Aufwaermen (10 s)"
sleep 10

rss() { awk '/VmRSS/{print $2}' "/proc/$1/status" 2>/dev/null; }

printf "\n%-16s %10s\n" "Modul" "RSS (kB)"
printf "%-16s %10s\n" "-----" "--------"
for spec in "bus:$BUS" "s-c-automation:$SC" "s-e-monitor:$SE" "s-d (node):$SD"; do
	name="${spec%%:*}"
	pid="${spec##*:}"
	k="$(rss "$pid")"
	printf "%-16s %10s\n" "$name" "${k:-0}"
done
total=0
for pid in "$BUS" "$SC" "$SE" "$SD"; do
	k="$(rss "$pid")"
	total=$((total + ${k:-0}))
done
printf "%-16s %10s\n" "Summe" "$total"

kill "$BUS" "$SC" "$SE" "$SD" 2>/dev/null || true
rm -rf "$OUT"