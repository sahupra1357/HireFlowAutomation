# Application Tracker

The master status board. `/job status` maintains it; `/job apply` appends rows.
`submitted` is only ever set after the **user** confirms they clicked Submit.

## Active

| Job ID | Role | Company | Fit | JD | Resume | Form | Status | Applied | Last update | Next action |
|---|---|---|---|---|---|---|---|---|---|---|
| _(none yet)_ | | | | | | | | | | |

## Closed

| Job ID | Role | Company | Outcome | Closed | Notes |
|---|---|---|---|---|---|
| _(none yet)_ | | | | | |

## Stats
- Found: 0 · Shortlisted: 0 · JD captured: 0 · Tailored: 0 · Filled: 0 · Submitted: 0 ·
  Interviewing: 0 · Offers: 0
- Awaiting manual JD download: 0
- Response rate: n/a

## Stage columns
Which artifact exists, independent of the linear Status. Mirrors `jobs/output/jobs.md`.

| Column | Written by | Values |
|---|---|---|
| **JD** | `/job jd` | `✓ <date>` · `⏳ manual` (user must download the posting) · `—` |
| **Resume** | `/job tailor` | `✓ <date>` · `—` |
| **Form** | `/job apply` | `✓ <date>` · `⏳ in-progress` · `—` |

## Status meanings
`found` → in a search file, not yet reviewed
`shortlisted` → passed triage, queued to apply
`jd-captured` → full JD on disk at `jobs/output/jds/<job-id>.md`
`tailored` → resume/cover letter written, form not started
`filled-awaiting-user` → **form complete, waiting on the user to review and submit**
`submitted` → user confirmed they submitted
`interviewing` / `offer` / `rejected` / `ghosted` (no reply 30+ days) / `skipped`
