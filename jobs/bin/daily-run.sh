#!/usr/bin/env bash
# The whole loop, once: search → JDs → tailored resumes → filled forms on screen.
#
#   jobs/bin/daily-run.sh              # run it now, watch it
#   jobs/bin/daily-run.sh --unattended # what launchd calls: no prompts, quieter
#   make daily                         # same as the first
#
# What it searches depends on whether `/job setup` has run:
#
#   set up      — master-resume.md is populated AND jobs/input/config/setup-families.md
#                 exists. Searches ONLY the families the user confirmed in setup, one
#                 `/job search "<family>"` per family, each with its own Result limit.
#   not set up  — falls back to ONE `/job search --limit $FALLBACK_LIMIT` (default 50) across
#                 every family in search-profile.md's generic list, 50 jobs in total.
#                 Tailoring then degrades as auto.md describes (no resume → no tailoring).
#
# Four stages, in order. Any agent stage can be skipped:
#
#   1. search            as above. Mechanical work, so it runs on $SEARCH_MODEL
#                        (default sonnet). (--no-search skips this stage)
#   2. /job auto --skip-search   pick up jobs/input/inbox.md, triage by the Auto-tailor
#                        threshold, capture JDs, evaluate, tailor resumes (not the ones
#                        evaluated `skip`). Judgment-heavy (fit scoring, no-fabrication tailoring),
#                        so it runs on $REASON_MODEL (default opus). Stops at `tailored`.
#                        (--no-tailor skips this stage)
#   3. /job apply --batch    maps any newly tailored posting's form and fills it in a
#                        visible browser on $APPLY_MODEL (default sonnet). Never asks,
#                        never waits, never submits.
#                        (--no-apply skips this stage)
#   4. morning-run.sh    re-opens/refills every mapped, unsubmitted application, so the
#                        tabs are all there even for jobs stage 3 had nothing new to do.
#
# Before stage 1, a health check (jobs/bin/check-index.py + check-setup.py --preflight,
# the same two `make doctor` runs). An index ERROR stops the run before anything is
# searched — every stage merges into jobs.md, so searching into a broken index makes it
# worse and costs tokens doing it. Setup findings never stop it (no resume, no confirmed
# families are modes this script already runs degraded in); they are printed and logged.
# (--no-doctor skips the check)
#
# Stages 1–3 are agent runs: they cost tokens. Stage 4 is free.
#
# Models are any value `claude --model` accepts (alias or full ID), overridable per run:
#   SEARCH_MODEL=haiku REASON_MODEL=opus FALLBACK_LIMIT=25 make daily
#   SITES=linkedin make daily    # stage 1 searches only these sources (search.md `--sites`)
#
# It cannot submit anything. That is a workspace rule the agent follows, and the two
# fill scripts click no submit control.

set -uo pipefail
cd "$(dirname "$0")/../.." || exit 1
ROOT=$(pwd)

DO_SEARCH=1; DO_TAILOR=1; DO_APPLY=1; DO_DOCTOR=1; UNATTENDED=0
for a in "$@"; do
  case "$a" in
    --no-search) DO_SEARCH=0 ;;
    --no-tailor) DO_TAILOR=0 ;;
    --no-apply)  DO_APPLY=0 ;;
    --no-doctor) DO_DOCTOR=0 ;;
    --unattended) UNATTENDED=1 ;;
    -h|--help) sed -n '2,47p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *) echo "unknown flag: $a"; exit 1 ;;
  esac
done

SEARCH_MODEL=${SEARCH_MODEL:-sonnet}
REASON_MODEL=${REASON_MODEL:-opus}
APPLY_MODEL=${APPLY_MODEL:-sonnet}
FALLBACK_LIMIT=${FALLBACK_LIMIT:-50}
SITES=${SITES:-}
SITES_ARG=${SITES:+ --sites $SITES}

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
command -v claude >/dev/null || { say "claude CLI not on PATH — install it or run with --no-search --no-tailor --no-apply"; exit 1; }
B="$HOME/.claude/skills/gstack/browse/dist/browse"
[ -x "$ROOT/.claude/skills/gstack/browse/dist/browse" ] && B="$ROOT/.claude/skills/gstack/browse/dist/browse"
[ -x "$B" ] || { say "gstack browse not found — stages 3 and 4 need it"; exit 1; }

say ""
say "════════ daily run · $(date '+%F %H:%M') ════════"
say "  log: $LOG"
say "  models: search=$SEARCH_MODEL · triage/JD/tailor=$REASON_MODEL · apply=$APPLY_MODEL"

# ── health check, before anything costs a token ─────────────────────────────
if [ "$DO_DOCTOR" = "1" ]; then
  say ""
  out=$(HIREFLOW_DAILY_RUN=1 python3 "$ROOT/jobs/bin/check-index.py" 2>&1); rc=$?
  printf '%s\n' "$out" | sed 's/^/  /' | tee -a "$LOG"
  if [ "$rc" != "0" ]; then
    say ""
    say "  ✋ jobs/output/jobs.md has errors (above). Nothing was searched."
    say "     Fix them, or roll back: make history → cp jobs/output/history/<snapshot> jobs/output/jobs.md"
    say "     Then re-run. (--no-doctor skips this check.)"
    exit 1
  fi
  python3 "$ROOT/jobs/bin/check-setup.py" --preflight 2>&1 | sed 's/^/  /' | tee -a "$LOG"
fi

# ── has /job setup run? ──────────────────────────────────────────────────────
# Setup is interactive (it asks which families to keep), so an unattended run can't do it.
# It can only check whether it happened: setup writes the confirmed families — and only
# those — to setup-families.md. That file present means "search exactly these".
MASTER="$ROOT/jobs/input/profile/master-resume.md"
PROFILE="$ROOT/jobs/input/config/search-profile.md"
CONFIRMED="$ROOT/jobs/input/config/setup-families.md"

# "### " headings under "## Families": strip the "N. " prefix and any trailing italic note
# like "*(thin — confirm …)*". The rest is passed verbatim as the search focus, so it matches
# the Profile column search.md stamps on each row.
families() {
  awk '/^## Families/{on=1; next} /^## /{on=0} on && /^### /' "$CONFIRMED" 2>/dev/null \
    | sed -E 's/^### +//; s/^[0-9]+\. +//; s/ *\*\(.*$//; s/[[:space:]]+$//' \
    | grep -v '{{'
}

MASTER_OK=1
if [ ! -s "$MASTER" ] || grep -q 'STATUS: not yet populated' "$MASTER" || ! grep -q '^### ' "$MASTER"; then
  MASTER_OK=0
fi
FAMILIES=()
while IFS= read -r f; do [ -n "$f" ] && FAMILIES+=("$f"); done < <(families)

if [ "$MASTER_OK" = "1" ] && [ "${#FAMILIES[@]}" -gt 0 ]; then
  MODE=setup
  say "  families: ${#FAMILIES[@]} confirmed in setup-families.md — $(printf '%s · ' "${FAMILIES[@]}" | sed 's/ · $//')"
else
  MODE=fallback
  [ "$MASTER_OK" = "0" ] && say "  ✋ master-resume.md is empty — tailoring will be skipped until /job setup runs."
  [ "${#FAMILIES[@]}" -eq 0 ] && say "  ✋ no setup-families.md — /job setup hasn't confirmed your families."
  say "  fallback: one search across search-profile.md's generic families, limit $FALLBACK_LIMIT total."
  say "  Run /job setup in Claude to search only the families you pick."
fi
T=$(sed -nE 's/^### Auto-tailor threshold: *([0-9]+).*/\1/p' "$PROFILE" | head -1)
say "  auto-tailor threshold: ${T:-75 (default — line missing from search-profile.md)}"

agent() {                     # $1 = prompt, $2 = label, $3 = model
  say ""
  say "── $2   [$3]"
  local out
  out=$(claude -p "$1" --model "$3" --permission-mode "$PERMISSION_MODE" --add-dir "$ROOT" 2>&1)
  printf '%s\n' "$out" >> "$LOG"
  if [ "$UNATTENDED" = "1" ]; then
    printf '%s\n' "$out" | tail -n 25
  else
    printf '%s\n' "$out"
  fi
}

# ── stage 1: search ──────────────────────────────────────────────────────────
# Serial on purpose: every search merges into the same jobs.md.
if [ "$DO_SEARCH" = "1" ] && [ "$MODE" = "setup" ]; then
  i=0
  for fam in "${FAMILIES[@]}"; do
    i=$((i+1))
    agent "/job search $fam$SITES_ARG" "stage 1 · search $i/${#FAMILIES[@]} · $fam${SITES:+ · $SITES}" "$SEARCH_MODEL"
  done
elif [ "$DO_SEARCH" = "1" ]; then
  agent "/job search --limit $FALLBACK_LIMIT$SITES_ARG" "stage 1 · search · all generic families, limit $FALLBACK_LIMIT${SITES:+ · $SITES}" "$SEARCH_MODEL"
else
  say ""; say "── stage 1 skipped"
fi

# ── stage 2: triage → JD → tailor ────────────────────────────────────────────
if [ "$DO_TAILOR" = "1" ]; then
  agent "/job auto --skip-search" "stage 2 · inbox, triage, JDs, evaluation, tailoring" "$REASON_MODEL"
else
  say ""; say "── stage 2 skipped"
fi

# ── stage 3: map and fill any new forms ──────────────────────────────────────
if [ "$DO_APPLY" = "1" ]; then
  agent "/job apply --batch" "stage 3 · mapping and filling forms" "$APPLY_MODEL"
else
  say ""; say "── stage 3 skipped"
fi

# ── stage 4: make sure every mapped job is open and filled ───────────────────
say ""
say "── stage 4 · opening every mapped application"
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
