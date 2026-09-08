#!/usr/bin/env bash
# Record that YOU submitted an application.
#
#   jobs/bin/mark-submitted.sh <job-id> [<job-id> ...]     ·   make submitted JOB=<job-id>
#
# The agent can never set `submitted` — it stops at a filled form, so only you know whether
# you clicked the button. This is how you tell the workspace, and it is what stops
# `make morning` from re-opening a job you already sent, every morning, forever.
#
# Sets Status `submitted` in jobs/output/jobs.md and the tracker, stamps today's date, and
# appends a line to the job's log with a +14-day follow-up.

set -uo pipefail
cd "$(dirname "$0")/../.." || exit 1

[ $# -gt 0 ] || { echo "usage: $0 <job-id> [<job-id> ...]"; exit 1; }

JOBS=jobs/output/jobs.md
TRACKER=jobs/output/tracker.md
TODAY=$(date +%F)
FOLLOWUP=$(date -v+14d +%F 2>/dev/null || date -d "+14 days" +%F)

for id in "$@"; do
  if ! grep -q "\`$id\`" "$JOBS"; then
    echo "✗ $id — no such job in $JOBS"
    continue
  fi
  if grep -q "\`$id\`.*| submitted |" "$JOBS"; then
    echo "· $id — already marked submitted"
    continue
  fi

  python3 - "$id" "$TODAY" <<'PY'
import re, sys
job_id, today = sys.argv[1], sys.argv[2]
for path in ("jobs/output/jobs.md", "jobs/output/tracker.md"):
    try:
        src = open(path, encoding="utf-8").read()
    except FileNotFoundError:
        continue
    out = []
    for line in src.split("\n"):
        if line.startswith("|") and f"`{job_id}`" in line:
            # only the last-but-one cell is Status in jobs.md; in the tracker it is named.
            cells = line.split("|")
            for i, c in enumerate(cells):
                if c.strip() in ("filled-awaiting-user", "tailored", "jd-captured",
                                 "shortlisted", "found"):
                    cells[i] = " submitted "
                    break
            line = "|".join(cells)
        out.append(line)
    open(path, "w", encoding="utf-8").write("\n".join(out))
PY

  log="jobs/output/applications/$id/log.md"
  if [ -f "$log" ]; then
    {
      echo ""
      echo "## $TODAY — submitted by the user"
      echo ""
      echo "- Submitted: $TODAY"
      echo "- Follow up if no reply by: $FOLLOWUP"
    } >> "$log"
  fi
  echo "✓ $id — submitted $TODAY · follow up $FOLLOWUP"
done
