#!/usr/bin/env bash
# Boots mbsecli against a scratch copy of the example model, so e2e tests
# (in particular the live-reload test) can freely edit the file on disk
# without touching the tracked examples/drone.sysml.
set -euo pipefail

PORT="${1:-4199}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ ! -x "$ROOT_DIR/mbsecli" ]; then
  echo "mbsecli binary not found at $ROOT_DIR/mbsecli — run 'make build' first" >&2
  exit 1
fi

WORK_DIR="$(mktemp -d)"
cp "$ROOT_DIR/examples/drone.sysml" "$WORK_DIR/drone.sysml"
echo "$WORK_DIR" > "$SCRIPT_DIR/.e2e-workdir"

exec "$ROOT_DIR/mbsecli" start "$WORK_DIR/drone.sysml" --port "$PORT"
