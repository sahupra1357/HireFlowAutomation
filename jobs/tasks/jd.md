# Task: `jd`

Run via `/job jd`. Formerly the standalone `jd-collect` skill.

Visit each job's Careers URL to confirm the role is still active, capture the full job description into jobs/output/jds/<job-id>.md, and flag the ones that need a manual download. Use when the user wants job descriptions collected, wants to check which postings are still live, or before tailoring resumes.

**Tool budget for this task:** Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion

Drives `$B` read-only — `goto`, `text`, `snapshot`. **Never fills or submits a form.** Never spawns subagents for browser work.

Anything outside that budget means you are running the wrong task — return to
`SKILL.md` and route again rather than reaching for the tool.

---

Turn a row in `jobs/output/jobs.md` into a real job description on disk.

Two jobs at once: **prove the posting is still active** (the careers page only lists open
reqs — that is the whole point of going there) and **capture the full JD** so
`/job tailor` has something honest to work against.

**Arguments** (default `--shortlisted`):

| Arg | Scope |
|---|---|
| `<job-id>` | that one job |
| `--all` | every row in the **Summary** table of `jobs/output/jobs.md` |
| `--shortlisted` | Summary rows with Status `shortlisted` (falls back to `--all` if none) |
| `--backlog` | the **Unverified backlog** table — these have no Job ID yet, so mint one first |
| `--refresh` | re-capture jobs that already have a JD file (default is to skip them) |

---

## Output — one file per job, one place

```
jobs/output/jds/<job-id>.md          ← the JD.  Format: jobs/input/templates/job-description.md
jobs/output/jds/<job-id>-source.pdf  ← optional raw drop from the user (manual path)
```

Both the automatic and the manual path write the **same** `jobs/output/jds/<job-id>.md`.
Nothing downstream needs to know which path produced it — it reads `Capture method:` if it
cares.

`jobs/output/jobs.md` keeps only its short Details summary. The full text lives here.

---

## Step 0 — Setup

```bash
_ROOT=$(git rev-parse --show-toplevel 2>/dev/null)
B=""; [ -n "$_ROOT" ] && [ -x "$_ROOT/.claude/skills/gstack/browse/dist/browse" ] && B="$_ROOT/.claude/skills/gstack/browse/dist/browse"
[ -z "$B" ] && B="$HOME/.claude/skills/gstack/browse/dist/browse"
[ -x "$B" ] && echo "READY: $B" || echo "NEEDS_SETUP"

mkdir -p jobs/output/jds
sed -n '/^## Summary/,/^## Unverified/p' jobs/output/jobs.md      # the worklist
ls jobs/output/jds/                                       # what already exists
```

Build the worklist: job ID, Role, Company, Apply URL, Careers URL, Source/ATS — from the
Summary row. Skip any job that already has `jobs/output/jds/<job-id>.md` with a
`Capture method:` that is not `manual-required`, unless `--refresh`.

**Before touching the network, check for user drops.** For every job in scope, if
`jobs/output/jds/<job-id>-source.*` exists, that job goes down the **manual pickup** path
(Step 3) instead of the ladder — the user has already done the work, don't redo it.

---

## Step 1 — The escalation ladder

Per job, in order. Stop at the first rung that yields a real JD.

### Rung 1 — the ATS API (preferred, no browser)

Official endpoints, per `jobs/input/config/job-sites.md` and the "use the official path" rule. These
are plain HTTP, so **several jobs can be fetched in parallel** — the browser cannot.

```bash
# greenhouse — job-boards.greenhouse.io/<board>/jobs/<id>
curl -sS "https://boards-api.greenhouse.io/v1/boards/<board>/jobs/<id>?questions=true" | jq .

# lever — jobs.lever.co/<company>/<posting-id>
curl -sS "https://api.lever.co/v0/postings/<company>/<posting-id>" | jq .

# ashby — jobs.ashbyhq.com/<org>/<uuid>   (org slug is CASE-SENSITIVE: "Reka", "Edison Scientific")
curl -sS "https://api.ashbyhq.com/posting-api/job-board/<org>?includeCompensation=true" | jq .
```

The board-level endpoints (Greenhouse `/jobs`, the Ashby board) are also the **liveness
check**: a job present and listed there is `active` — that is the API equivalent of reading
the careers page, and it counts.

Do not trust the endpoint blindly. **Verify the payload before accepting it:**

- a description field is actually present and non-empty (Ashby's board response does not
  always carry one — if it doesn't, fall through to rung 2, don't fabricate a JD from the
  title and location),
- the decoded text is longer than ~400 characters,
- it reads like a job description (responsibilities/requirements language), not a cookie
  banner, an error body, or a "job not found" page served with HTTP 200 — `jobs/output/jobs.md`
  already has four jobs in **Excluded** that did exactly that.

Strip HTML to text before writing (`textutil -convert txt -stdin -stdout` handles it).

### Rung 2 — the Careers URL in the browser

```bash
$B goto "<careers-url>"
$B wait --networkidle
$B text
```

Find the role by title on the employer's own board. Found → follow it to the JD, capture,
`Liveness: active`, `Proved by:` the careers URL.

**Not found is not automatically dead.** Distinguish:

- the board rendered a list of jobs and this role is absent → genuinely **`dead`**
- the board rendered **nothing at all** (lazy-loaded React, infinite spinner, bot wall) →
  **inconclusive**, fall to rung 3. `jobs/output/jobs.md` records exactly this for Anthropic. An
  empty render is not evidence of absence.

### Rung 3 — the Apply URL in the browser

```bash
$B goto "<apply-url>"
$B wait --networkidle
$B text
$B forms          # cheap, and fills in the "Application form — observed" section
```

This gets the JD but **proves nothing about liveness**. Write `Liveness: inconclusive` and
`Proved by: not established`. Never round this up to `active`.

If the apply URL 404s, or serves a "Job not found" body under HTTP 200, the posting is
**`dead`** — that is real evidence, unlike an empty careers render.

### Rung 4 — `manual-required`

Everything above failed: bot wall, login wall, CAPTCHA, Workday tenant that needs an
account, or a JD published only as a PDF. **Do not fight it and do not retry a third time**
— CLAUDE.md rule 5. Write the stub (Step 2) and move to the next job.

Never create an account or accept terms to reach a JD.

---

## Step 2 — The `manual-required` stub

Write `jobs/output/jds/<job-id>.md` from the template with the header filled in as far as
it is known, `Capture method: manual-required`, and a **Capture notes** section that tells
the user exactly what to do:

```markdown
- **Capture method:** manual-required
- **Liveness:** inconclusive

## Full job description

_Not captured automatically._

## Capture notes

Auto-capture failed at every rung:
- ATS API: no Greenhouse/Lever/Ashby endpoint for this board
- Careers page (https://www.anthropic.com/careers): rendered 0 listings — lazy-loaded
- Apply URL (https://job-boards.greenhouse.io/anthropic/jobs/5390966008): Cloudflare check

**To finish this one:** open the Apply URL above, save the posting into
`jobs/output/jds/` as `anthropic--ai-engineer-gtm-claudification-source.pdf`
(`.html`, `.txt`, `.md`, `.docx` and `.png`/`.jpg` screenshots all work too),
then run `/job jd anthropic--ai-engineer-gtm-claudification`.
```

Set the **JD** column in `jobs/output/jobs.md` to `⏳ manual`. Leave **Status** alone — a job
waiting on a manual JD has not progressed.

---

## Step 3 — Manual pickup

When `jobs/output/jds/<job-id>-source.*` exists, parse it into the JD file:

| Drop | How to read it |
|---|---|
| `.pdf` | `Read` with `pages` — it renders PDFs directly |
| `.png` / `.jpg` | `Read` — it reads images; transcribe the posting text |
| `.html` / `.docx` / `.rtf` | `textutil -convert txt -stdout <file>` |
| `.txt` / `.md` | `cat` |

Then write the full `jobs/output/jds/<job-id>.md`: `Capture method: manual`,
`Liveness: active` **only if the user's capture shows the posting live** (a saved careers
page listing, a live apply form) — otherwise `inconclusive`. Extract requirements and ATS
keywords the same as any other rung.

Keep the source file. It is the evidence for what was captured.

Set **JD** to `✓ <date>` and Status to `jd-captured`.

---

## Step 4 — Write the index (parent only, one writer)

After each job — not in a batch at the end, so an interrupted run keeps what it earned:

**`jobs/output/jobs.md`**, Summary row:

| Column | Value |
|---|---|
| **JD** | `✓ YYYY-MM-DD` · `⏳ manual` · `—` |
| **Status** | `jd-captured` on success · unchanged on `manual-required` |
| **Verified** | `verified-live` only when liveness is `active`. Never downgrade a stronger existing value on an inconclusive read. |

A job that came back **`dead`** moves out of Summary into the **Excluded** table with the
verdict and the evidence, exactly as `/job search` does it. Delete its JD file if one was
stubbed.

`jobs/output/tracker.md` gets the same JD state. Nothing else writes these files during
this run.

---

## Step 5 — Report

```
JD collect — 10 jobs

  ✓ captured    7   (5 via ATS API, 2 via careers page)
  ⏳ manual      2   ← need you
  ✗ dead        1   (aptura--mts-applied-ai → Excluded: role absent from live board)

  Liveness:  6 active · 1 inconclusive (JD read off the apply URL only)

  Manual — save the posting into jobs/output/jds/ as <job-id>-source.pdf
  (.html / .txt / .md / .docx / screenshot all fine):

    anthropic--ai-engineer-gtm-claudification
      https://job-boards.greenhouse.io/anthropic/jobs/5390966008
    valthos--mts-applied-ai-engineer
      https://jobs.ashbyhq.com/valthos/ace82983-1564-47d0-a328-74c3bd9d03c8

  Then: /job jd --all      (picks up whatever you dropped in)
  Next: /job tailor --all   (7 ready)
```

Keep it to this. The detail is in the files.

---

## Rules

- **Verbatim JD.** Capture the posting's own words in the "Full job description" section.
  Summarizing here is how tailoring quietly drifts into invention two steps later.
- **Never upgrade liveness.** `active` requires the employer's own board or the ATS live
  index. Reading a JD off an apply URL is `inconclusive`, permanently.
- **An empty render is not a dead job.** Only a board that listed other jobs and not this
  one, or a real 404 / "not found" body, is evidence of death.
- **Human-scale request volume.** Rung 1 fetches can go in parallel; browser work is serial
  and unhurried. If a site blocks automated access, stop — `manual-required` exists for
  exactly this.
- **Never send anything from `jobs/input/profile/` anywhere.** This command only reads.
