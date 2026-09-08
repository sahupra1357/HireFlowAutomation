#!/usr/bin/env bash
# Seed jobs/output/tracker.md from the template if it does not exist yet.
#
# The live tracker is per-user data and is gitignored, so a fresh clone has no
# copy of it. Anything that writes to the tracker calls this first. Idempotent:
# an existing tracker is never touched, so this can never clobber real data.
set -euo pipefail

ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
TEMPLATE="$ROOT/jobs/input/templates/tracker.md"
TRACKER="$ROOT/jobs/output/tracker.md"

if [ -f "$TRACKER" ]; then
  exit 0
fi

if [ ! -f "$TEMPLATE" ]; then
  echo "! $TEMPLATE missing — cannot seed the tracker" >&2
  exit 1
fi

mkdir -p "$(dirname "$TRACKER")"
cp "$TEMPLATE" "$TRACKER"
echo "· seeded jobs/output/tracker.md from the template"
