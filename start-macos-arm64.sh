#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")"
./dist/ashen-crown-macos-arm64 -web ./web -data ./data &
PID=$!
sleep 1
open http://localhost:8080 >/dev/null 2>&1 || true
printf 'Ashen Crown is running at http://localhost:8080 (PID %s)\n' "$PID"
wait "$PID"
