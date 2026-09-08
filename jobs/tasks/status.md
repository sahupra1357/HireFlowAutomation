# Task: `status`

Run via `/job status`. Formerly the standalone `application-tracker` skill.

Show and maintain the job application status board — what's in flight, what needs follow-up, response rates, and logging outcomes like interviews and rejections. Use when the user asks about application status, wants to log a response, or asks what to follow up on.

**Tool budget for this task:** Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion

**No browser, no web search.** Reads and updates status files only.

Anything outside that budget means you are running the wrong task — return to
`SKILL.md` and route again rather than reaching for the tool.

---

The status board across everything in flight, plus the nudges that keep applications from
going stale.

**Arguments (all optional):**
- *(none)* — show the board and anything needing attention
- `<job-id> <status>` — log an outcome (`/job status sierra--ai-engineer rejected`)
- `follow-ups` — just what's due
- `stats` — funnel numbers

## Step 1 — Load

```bash
cat jobs/output/tracker.md
sed -n '/^## Summary/,/^## Unverified/p' jobs/output/jobs.md   # the JD / Resume / Form columns
ls jobs/output/jds/*.md 2>/dev/null | head -50        # captured JDs
ls jobs/output/applications/*/resume.pdf 2>/dev/null | head -50     # tailored resumes
ls jobs/output/applications/*/log.md 2>/dev/null | head -50
date +%F
```

Reconcile: every `jobs/output/applications/<job-id>/log.md` should have a tracker row, and every active
tracker row should have a log. Fix drift silently — a per-job log is the source of truth for
its own status; the tracker is the index.

**Reconcile the stage columns against what's actually on disk.** `jobs/output/jobs.md` is what the
dashboard shows, so a stale cell there is the board lying to the user:

| Column | True when |
|---|---|
| **JD** `✓` | `jobs/output/jds/<job-id>.md` exists with `Capture method:` ≠ `manual-required` |
| **JD** `⏳ manual` | that file exists but is still a `manual-required` stub |
| **Resume** `✓` | `jobs/output/applications/<job-id>/resume.pdf` exists |
| **Form** `✓` | the log records the form filled and Status is `filled-awaiting-user` or later |

Correct any cell that disagrees with the filesystem, and say how many you fixed.

## Step 2 — Show the board

```
Applications — 2026-08-30

  IN FLIGHT (4)
  anthropic--forward-deployed-engineer   submitted 8-25   5d  no response
  sierra--ai-engineer                    interviewing     screen Thu 9-04
  glean--data-engineer                   submitted 8-18  12d  no response
  scale--ai-engineer                     filled-awaiting-user  ⚠️ 3d — needs your submit

  QUEUED (3)
  databricks--data-engineer   78  Workday ~45min
  ...

  PIPELINE          JD    Resume   Form
  ramp                ✓        ✓      —   → /job apply ramp--applied-ai-engineer
  perplexity          ✓        —      —   → /job tailor perplexity--mts-...
  anthropic     ⏳ manual       —      —   → drop the JD in, then /job jd

  NEEDS ATTENTION
  ⚠️  scale--ai-engineer sitting unsubmitted for 3 days
  📮  glean--data-engineer — 12 days, no reply. Follow up?
  💀  langchain--ai-engineer — 34 days silent → mark ghosted?

  FUNNEL   found 34 → shortlisted 9 → submitted 6 → responses 2 → interviews 1
           response rate 33% · median time to response 6d
```

Flag anything at `filled-awaiting-user` for more than 2 days — the user has a finished
application sitting there. Postings close.

Flag anything at `⏳ manual` for more than 3 days too — a JD nobody downloaded blocks that
job's entire pipeline, and the user is the only one who can unblock it. List the URL again
rather than making them go find it.

## Step 3 — Handle updates

When the user logs an outcome, update both the per-job `log.md` (status + timeline entry)
and the tracker row, then move the row to Closed if terminal. Recompute the Stats line.

Auto-suggestions, always as a question, never automatic:
- No response after 14 days → offer to draft a follow-up email
- No response after 30 days → offer to mark `ghosted` and move to Closed
- `interviewing` → offer to pull `jobs/output/jds/<job-id>.md` and the master resume into
  a prep sheet

## Step 4 — Follow-up drafts

When asked, write to `jobs/output/applications/<job-id>/follow-up-<date>.md`. Short: reference the role
and date applied, add one genuinely new and relevant thing (a shipped project, a relevant
release), reiterate interest in a sentence, close. Under 150 words. The user sends it — the
agent does not send email.

## Step 5 — Patterns worth surfacing

Once there are ~10 applications, look for what the data says and mention it in one or two
lines: which role family gets the best response rate, whether Greenhouse-hosted roles
outperform Workday ones, whether the fit score actually predicts responses, whether comp
band correlates with silence. Use it to suggest tuning `jobs/input/config/search-profile.md`.

Only surface a pattern the numbers support. With 4 data points, say nothing.
