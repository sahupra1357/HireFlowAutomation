#!/usr/bin/env bash
# The whole loop, once: search → JDs → tailored resumes → filled forms on screen.
#
#   jobs/bin/daily-run.sh              # run it now, watch it
#   jobs/bin/daily-run.sh --unattended # what launchd calls: no prompts, quieter
#   make daily                         # same as the first
#
# Three stages, in order. Any one of them can be skipped:
#
#   1. /job              the auto pipeline — search, verify, capture JDs, tailor resumes.
#                        Reconciles: finished work is skipped, duplicates are dropped by
#                        job ID, a stuck job never stops the batch. Stops at `tailored`.
#                        (--no-search skips this stage)
#   2. /job apply --batch    maps any newly tailored posting's form and fills it in a
#                        visible browser. Never asks, never waits, never submits.
#                        (--no-apply skips this stage)
#   3. morning-run.sh    re-opens/refills every mapped, unsubmitted application, so the
#                        tabs are all there even for jobs stage 2 had nothing new to do.
#
# Stages 1 and 2 are agent runs: they cost tokens. Stage 3 is free.
#
# It cannot submit anything. That is a workspace rule the agent follows, and the two
# fill scripts click no submit control.

set -uo pipefail
cd "$(dirname "$0")/../.." || exit 1
ROOT=$(pwd)

DO_SEARCH=1; DO_APPLY=1; UNATTENDED=0
for a in "$@"; do
  case "$a" in
    --no-search) DO_SEARCH=0 ;;
    --no-apply)  DO_APPLY=0 ;;
    --unattended) UNATTENDED=1 ;;
    -h|--help) sed -n '2,26p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *) echo "unknown flag: $a"; exit 1 ;;
  esac
done

# An unattended agent cannot answer a permission prompt, so the scheduled run bypasses them.
# Override with PERMISSION_MODE=acceptEdits (etc.) if you keep a tighter allowlist.
# NOTE: bypassing permission *prompts* does not loosen the workspace's own rules — never
# submit, never fabricate, never guess a screening answer — which are instructions the agent
# follows, not approval gates.
PERMISSION_MODE=${PERMISSION_MODE:-bypassPermissions}

LOGDIR="$ROOT/jobs/output/logs"
mkdir -p "$LOGDIR"
LOG="$LOGDIR/daily-$(date +%F).log"
LOCK="$ROOT/jobs/output/.daily-run.lock"

say() { printf '%s\n' "$*" | tee -a "$LOG"; }

# ── one at a time ────────────────────────────────────────────────────────────
# A manual run and the scheduled one both editing jobs.md would corrupt the index.
if ! mkdir "$LOCK" 2>/dev/null; then
  pid=$(cat "$LOCK/pid" 2>/dev/null || echo "")
  if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
    say "Another daily run is going (pid $pid). Nothing to do."
    exit 0
  fi
  say "Clearing a stale lock from a run that died."
  rm -rf "$LOCK"; mkdir "$LOCK" 2>/dev/null || { say "could not take the lock"; exit 1; }
fi
echo $$ > "$LOCK/pid"
trap 'rm -rf "$LOCK"' EXIT INT TERM

# ── preflight ────────────────────────────────────────────────────────────────
command -v claude >/dev/null || { say "claude CLI not on PATH — install it or run with --no-search --no-apply"; exit 1; }
B="$HOME/.claude/skills/gstack/browse/dist/browse"
[ -x "$ROOT/.claude/skills/gstack/browse/dist/browse" ] && B="$ROOT/.claude/skills/gstack/browse/dist/browse"
[ -x "$B" ] || { say "gstack browse not found — stage 2 and 3 need it"; exit 1; }

say ""
say "════════ daily run · $(date '+%F %H:%M') ════════"
say "  log: $LOG"

agent() {                     # $1 = prompt, $2 = label
  say ""
  say "── $2"
  local out
  out=$(claude -p "$1" --permission-mode "$PERMISSION_MODE" --add-dir "$ROOT" 2>&1)
  printf '%s\n' "$out" >> "$LOG"
  if [ "$UNATTENDED" = "1" ]; then
    printf '%s\n' "$out" | tail -n 25
  else
    printf '%s\n' "$out"
  fi
}

# ── stage 1: search → JD → tailor ────────────────────────────────────────────
if [ "$DO_SEARCH" = "1" ]; then
  agent "/job" "stage 1 · search, JDs, tailoring"
else
  say ""; say "── stage 1 skipped (--no-search)"
fi

# ── stage 2: map and fill any new forms ──────────────────────────────────────
if [ "$DO_APPLY" = "1" ]; then
  agent "/job apply --batch" "stage 2 · mapping and filling forms"
else
  say ""; say "── stage 2 skipped (--no-apply)"
fi

# ── stage 3: make sure every mapped job is open and filled ───────────────────
say ""
say "── stage 3 · opening every mapped application"
bash "$ROOT/jobs/bin/morning-run.sh" 2>&1 | tee -a "$LOG"

# ── where things stand ───────────────────────────────────────────────────────
JOBS="$ROOT/jobs/output/jobs.md"
count() { local n; n=$(grep -c "| $1 |" "$JOBS" 2>/dev/null); echo "${n:-0}"; }
say ""
say "════════ done · $(date '+%H:%M') ════════"
say "  found $(count found) · tailored $(count tailored) · filled $(count filled-awaiting-user) · submitted $(count submitted)"
manual=$(grep -c '⏳ manual' "$JOBS" 2>/dev/null); manual=${manual:-0}
[ "$manual" != "0" ] && say "  $manual job(s) need you to download the JD by hand — see the dashboard's Needs-you panel"
say "  Nothing was submitted."
say "  After you submit one:  make submitted JOB=<job-id>"

# keep a month of logs, no more
ls -1t "$LOGDIR"/daily-*.log 2>/dev/null | tail -n +31 | xargs -r rm -f
