#!/usr/bin/env bash
# Put the workspace back to a fresh start.
#
#   jobs/bin/reset.sh --dry-run     # list exactly what would go, delete nothing
#   jobs/bin/reset.sh               # wipe agent output, after you type the confirmation
#   jobs/bin/reset.sh --profile     # ALSO wipe your resume, links and answer bank
#   jobs/bin/reset.sh --all         # --profile plus build artifacts
#   make reset                      # same as the plain form
#
# What goes by default — everything the agent generated:
#   jobs/output/jobs.md · run-log.md · history/ · jds/ · applications/ · logs/
#   interviews/output/interviews.md · reports/
#   jobs/output/tracker.md is restored to the blank board, not deleted.
#
# What stays by default — everything YOU provided:
#   jobs/input/profile/  (source resumes, master-resume.md, links.md)
#   jobs/input/config/   (job-sites.md, search-profile.md, application-answers.md)
#   jobs/tasks/ · jobs/input/templates/ · the dashboard · this script
#
# `--profile` takes the personal files too, which means the next run needs `/job setup`
# and a resume dropped back into jobs/input/profile/source-resumes/. It asks twice.
#
# Everything removed is archived first to .resets/reset-<timestamp>.tar.gz unless you pass
# --no-backup, so a reset you regret is one `tar -xzf` away.

set -uo pipefail
cd "$(dirname "$0")/../.." || exit 1
ROOT=$(pwd)

DRY=0; WIPE_PROFILE=0; WIPE_BUILD=0; ASSUME_YES=0; BACKUP=1
for a in "$@"; do
  case "$a" in
    --dry-run|-n) DRY=1 ;;
    --profile)    WIPE_PROFILE=1 ;;
    --all)        WIPE_PROFILE=1; WIPE_BUILD=1 ;;
    --yes|-y)     ASSUME_YES=1 ;;
    --no-backup)  BACKUP=0 ;;
    -h|--help)    sed -n '2,27p' "$0" | sed 's/^# \{0,1\}//'; exit 0 ;;
    *) echo "unknown flag: $a  (try --help)"; exit 1 ;;
  esac
done

# ── what is on the chopping block ────────────────────────────────────────────
TARGETS=(
  "jobs/output/jobs.md"
  "jobs/output/run-log.md"
  "jobs/output/history"
  "jobs/output/jds"
  "jobs/output/applications"
  "jobs/output/logs"
  "jobs/output/.daily-run.lock"
  "interviews/output/interviews.md"
  "interviews/output/reports"
)
if [ "$WIPE_PROFILE" = "1" ]; then
  TARGETS+=(
    "jobs/input/profile/source-resumes"
    "jobs/input/profile/master-resume.md"
    "jobs/input/profile/links.md"
    "jobs/input/config/application-answers.md"
    "jobs/input/config/search-profile.md"
  )
fi
[ "$WIPE_BUILD" = "1" ] && TARGETS+=("bin")

present=()
for t in "${TARGETS[@]}"; do [ -e "$ROOT/$t" ] && present+=("$t"); done

size_of() { du -sh "$ROOT/$1" 2>/dev/null | cut -f1 | tr -d ' '; }
count_of() { [ -d "$ROOT/$1" ] && printf '%s files' "$(find "$ROOT/$1" -type f | wc -l | tr -d ' ')" || printf 'file'; }

echo
echo "════════ reset · $(date '+%F %H:%M') ════════"
if [ ${#present[@]} -eq 0 ]; then
  echo "  Already clean — nothing to remove."
  exit 0
fi
echo
echo "  WOULD REMOVE:"
for t in "${present[@]}"; do printf '    %-42s %6s  %s\n' "$t" "$(size_of "$t")" "$(count_of "$t")"; done
echo
echo "  WOULD RESTORE:"
echo "    jobs/output/tracker.md                        → the blank board"
echo "    empty jobs/output/{history,jds,applications}, interviews/output/reports"
echo
if [ "$WIPE_PROFILE" = "1" ]; then
  echo "  ⚠ --profile: your resume, links and answer bank go too."
  echo "    Afterwards you must drop a resume into jobs/input/profile/source-resumes/"
  echo "    and run /job setup before anything else works."
else
  echo "  KEEPING your own files: jobs/input/profile/ and jobs/input/config/"
fi
echo

if [ "$DRY" = "1" ]; then
  echo "  --dry-run: nothing was touched."
  exit 0
fi

# ── confirm ──────────────────────────────────────────────────────────────────
if [ "$ASSUME_YES" != "1" ]; then
  if [ ! -t 0 ]; then
    echo "  Refusing to wipe without a terminal to confirm at. Pass --yes if you mean it."
    exit 1
  fi
  printf '  Type reset to continue (anything else aborts): '
  read -r reply
  [ "$reply" = "reset" ] || { echo "  Aborted. Nothing was touched."; exit 1; }
  if [ "$WIPE_PROFILE" = "1" ]; then
    printf '  This also deletes your resume and answer bank. Type profile to confirm: '
    read -r reply2
    [ "$reply2" = "profile" ] || { echo "  Aborted. Nothing was touched."; exit 1; }
  fi
fi

# ── stop anything holding these files open ───────────────────────────────────
# The dashboard serves jobs.md, and the headed browser may be sitting on filled forms that
# are about to stop existing in the tracker. Both come down before the delete.
pkill -f "jobs-dashboard -addr" >/dev/null 2>&1 && echo "  · stopped the dashboard"
B="$HOME/.claude/skills/gstack/browse/dist/browse"
[ -x "$ROOT/.claude/skills/gstack/browse/dist/browse" ] && B="$ROOT/.claude/skills/gstack/browse/dist/browse"
if [ -x "$B" ]; then
  "$B" disconnect >/dev/null 2>&1 && echo "  · closed the browser session"
fi

# ── archive, then delete ─────────────────────────────────────────────────────
if [ "$BACKUP" = "1" ]; then
  mkdir -p "$ROOT/.resets"
  ARCHIVE="$ROOT/.resets/reset-$(date +%Y-%m-%dT%H-%M-%S).tar.gz"
  if tar -czf "$ARCHIVE" -C "$ROOT" "${present[@]}" 2>/dev/null; then
    echo "  · backed up to ${ARCHIVE#$ROOT/}  ($(du -sh "$ARCHIVE" | cut -f1 | tr -d ' '))"
  else
    echo "  ✗ backup failed — stopping rather than deleting anything"
    exit 1
  fi
fi

for t in "${present[@]}"; do
  rm -rf "${ROOT:?}/$t"
  echo "  · removed $t"
done

# ── put the scaffolding back ─────────────────────────────────────────────────
for d in jobs/output/history jobs/output/jds jobs/output/applications interviews/output/reports; do
  mkdir -p "$ROOT/$d"
  touch "$ROOT/$d/.gitkeep"
done
[ "$WIPE_PROFILE" = "1" ] && { mkdir -p "$ROOT/jobs/input/profile/source-resumes"; touch "$ROOT/jobs/input/profile/source-resumes/.gitkeep"; }

if [ -f "$ROOT/jobs/input/templates/tracker.md" ]; then
  cp "$ROOT/jobs/input/templates/tracker.md" "$ROOT/jobs/output/tracker.md"
  echo "  · restored jobs/output/tracker.md from the template"
else
  echo "  ! jobs/input/templates/tracker.md missing — tracker not restored"
fi

echo
echo "════════ clean ════════"
if [ "$WIPE_PROFILE" = "1" ]; then
  echo "  Factory state. Next: drop a resume into jobs/input/profile/source-resumes/,"
  echo "  then run  /job setup"
else
  echo "  Your resume, links, answer bank and search profile are untouched."
  echo "  Next:  /job          (search → JDs → tailored resumes)"
fi
[ "$BACKUP" = "1" ] && echo "  Undo:  tar -xzf ${ARCHIVE#$ROOT/}"
echo
