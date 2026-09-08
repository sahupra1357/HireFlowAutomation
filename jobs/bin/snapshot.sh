#!/usr/bin/env bash
# Snapshot jobs.md before a run rewrites it.
#
# Every /job run that touches the index calls this first, so a bad merge, a truncated
# table, or a status the agent got wrong is always one file away from being recovered.
#
#   jobs/bin/snapshot.sh                    # snapshot the default index
#   jobs/bin/snapshot.sh path/to/jobs.md    # snapshot a specific file
#
# Writes jobs/output/history/jobs-YYYY-MM-DDTHH-MM-SS.md and prints the path.
set -euo pipefail

ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
SRC=${1:-"$ROOT/jobs/output/jobs.md"}
DIR="$ROOT/jobs/output/history"
KEEP=${SNAPSHOT_KEEP:-50}

[ -f "$SRC" ] || { echo "snapshot: no $SRC yet — nothing to back up" >&2; exit 0; }

mkdir -p "$DIR"
DEST="$DIR/jobs-$(date +%Y-%m-%dT%H-%M-%S).md"

# Don't write a duplicate when nothing changed since the last snapshot.
LAST=$(ls -1 "$DIR"/jobs-*.md 2>/dev/null | tail -1 || true)
if [ -n "$LAST" ] && cmp -s "$SRC" "$LAST"; then
  echo "$LAST"
  exit 0
fi

cp "$SRC" "$DEST"

# Keep the most recent $KEEP snapshots; drop the rest oldest-first.
# (BSD head has no negative -n, so count and slice instead.)
N=$(ls -1 "$DIR"/jobs-*.md 2>/dev/null | wc -l | tr -d ' ')
if [ "$N" -gt "$KEEP" ]; then
  ls -1 "$DIR"/jobs-*.md | sed -n "1,$((N - KEEP))p" | while read -r old; do rm -f "$old"; done
fi

echo "$DEST"
