#!/usr/bin/env bash
# Morning run — open every mapped application, filled, in a visible browser.
#
#   jobs/bin/morning-run.sh              # every job that has a field map and isn't submitted
#   jobs/bin/morning-run.sh <job-id> ... # just these
#
# No agent, no tokens, no thinking: it replays the field maps `/job apply --batch` already
# wrote. Each job gets its own tab; the window stays open when this exits. You look through
# the tabs, finish the blanks, and submit the ones you want.
#
# A job with no form-fill.json is listed at the end, not filled — mapping a new posting is
# the agent's job: /job apply --batch <job-id>.
#
# It never submits. fill-form.py clicks form controls only.

set -uo pipefail
cd "$(dirname "$0")/../.." || exit 1

B="$HOME/.claude/skills/gstack/browse/dist/browse"
[ -x .claude/skills/gstack/browse/dist/browse ] && B=".claude/skills/gstack/browse/dist/browse"
[ -x "$B" ] || { echo "browse not found — install the gstack browse skill"; exit 1; }

APPS=jobs/output/applications
JOBS=jobs/output/jobs.md

# Which jobs? Named ones, or every mapped job that the user hasn't submitted yet.
if [ $# -gt 0 ]; then
  ids=("$@")
else
  ids=()
  for f in "$APPS"/*/form-fill.json; do
    [ -e "$f" ] || continue
    id=$(basename "$(dirname "$f")")
    # skip anything already sent — never re-open a submitted application
    if grep -q "\`$id\`.*| submitted |" "$JOBS" 2>/dev/null; then continue; fi
    ids+=("$id")
  done
fi

if [ ${#ids[@]} -eq 0 ]; then
  mapped=$(ls -d "$APPS"/*/form-fill.json 2>/dev/null | wc -l | tr -d ' ')
  if [ "$mapped" = "0" ]; then
    echo "Nothing mapped yet. Run:  /job apply --batch"
  else
    echo "Nothing to open — all $mapped mapped application(s) are already submitted."
    echo "Map a new one with:  /job apply --batch <job-id>"
  fi
  exit 0
fi

# The daemon must be headed, and a headless one has to go first — browse refuses to mix.
# Only a daemon we KNOW is headless gets disconnected. Anything else — already headed, or
# not running at all — is left alone: `--headed` starts a cold daemon headed by itself, and
# disconnecting a live headed one would close every filled tab on screen.
st=$("$B" --headed status 2>&1)
mode=$(printf '%s' "$st" | awk '/^Mode:/{print $2}')
if [ "$mode" = "launched" ]; then
  echo "→ switching the browser from headless to a visible window"
  "$B" disconnect >/dev/null 2>&1; sleep 1
elif ! printf '%s' "$st" | grep -q '^Status: healthy'; then
  # Closing the window leaves the daemon alive with no page, and every command then fails
  # with "No active page". Nothing is lost by restarting it — there are no tabs left.
  echo "→ browser session is wedged (window closed?) — restarting it"
  "$B" disconnect >/dev/null 2>&1; sleep 1
fi

echo "Opening ${#ids[@]} application(s) — nothing will be submitted."
echo

ok=0; bad=0
for id in "${ids[@]}"; do
  if [ ! -f "$APPS/$id/form-fill.json" ]; then
    echo "▢ $id — not mapped yet, run: /job apply --batch $id"
    bad=$((bad+1)); continue
  fi
  echo "── $id"
  url=$(python3 -c "import json,sys;print(json.load(open('$APPS/$id/form-fill.json')).get('apply_url',''))")
  # Re-use a tab already open on this posting. A run that happens every morning must be
  # idempotent, or a week of it buries the desktop in duplicate tabs of the same job.
  # NB: headed mode lists one real tab twice (page + extension target). Matching by URL and
  # taking the first listing is therefore correct — and stops a daily run stacking tabs.
  tab=$("$B" --headed tabs 2>/dev/null | grep -F "$url" | grep -oE '\[[0-9]+\]' | head -1 | tr -d '[]')
  if [ -n "$tab" ]; then
    echo "  (reusing tab $tab)"
  else
    # newtab prints "Opened tab N → url"; older builds print "[N]". Take either.
    tab=$("$B" --headed newtab "$url" 2>/dev/null | grep -oE 'tab [0-9]+|\[[0-9]+\]' | head -1 | grep -oE '[0-9]+')
    sleep 2
  fi
  if [ -n "$tab" ]; then
    python3 jobs/bin/fill-form.py "$id" --tab "$tab" && ok=$((ok+1)) || bad=$((bad+1))
  else
    python3 jobs/bin/fill-form.py "$id" && ok=$((ok+1)) || bad=$((bad+1))
  fi
  echo
done

echo "════════════════════════════════════════════"
echo "  $ok filled · $bad need attention"
echo "  The browser is open. Review each tab, fill the blanks, submit yourself."
echo "  Nothing was submitted."
