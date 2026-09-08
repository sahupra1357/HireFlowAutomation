<!-- SECTION SHAPES for jobs/output/jobs.md — the single living index /job search merges into.
     Not a file to create: jobs/output/jobs.md already exists. Match these column sets exactly,
     because jobs/dashboard/ parses the tables by header name. -->

# Jobs

- **Run:** {{DATE}} {{TIME}}
- **Sites searched:** {{list}}
- **Sites skipped:** {{site — reason}}
- **Queries:** {{the actual search strings used}}
- **Found:** {{N}} · **New:** {{N}} · **Duplicates of earlier runs:** {{N}}
- **Excluded by verification:** {{N}} dead · {{N}} stale · {{N}} agency-repost · {{N}} suspicious/scam
- **Filters inactive this run:** {{list, or "none"}}
- **Result limit:** {{N | all}} — {{"reached, stopped early" | "pool exhausted at M" | "exhaustive, completed in M min" | "exhaustive, stopped at budget"}}
- **Sites unswept:** {{list, or "none"}} (next run starts here)

## Summary

| Fit | Profile | Verified | Job ID | Role | Company | Location | Comp | Source | Apply URL | Careers | JD | Resume | Form | Status |
|-----|---------|----------|--------|------|---------|----------|------|--------|-----------|---------|----|--------|------|--------|
| 87 | {{role family}} | verified-live | `anthropic--forward-deployed-engineer` | Forward Deployed Engineer | Anthropic | SF / Remote | $300–405k | greenhouse | https://job-boards.greenhouse.io/... | https://www.anthropic.com/careers | — | — | — | found |

Sorted by fit, descending. `Status` starts at `found`; `/job triage` moves it to
`shortlisted` or `skipped`.

**The three stage columns start at `—` and are written by later skills only** — `JD` by
`/job jd`, `Resume` by `/job tailor`, `Form` by `/job apply`. `/job search` always
emits them as `—`; the row must have all 15 cells or the dashboard's table goes crooked.

---

## 1. {{Role}} — {{Company}}   ·   Fit {{score}}/100

- **Job ID:** `{{job-id}}`
- **First seen:** {{YYYY-MM-DD}}
- **Apply URL:** {{direct link to the application form, not the aggregator}}
- **Careers page:** {{the employer's own careers page; falls back to the ATS board root}}
- **Posting URL:** {{where it was found}}
- **Source:** {{site name}}
- **ATS:** {{greenhouse | lever | ashby | workday | custom | unknown}}
- **Location / mode:** {{}}
- **Compensation:** {{band, or "not listed"}}
- **Posted:** {{YYYY-MM-DD, or "unknown"}}
- **Verified:** {{verified-live | likely-live | unverified | pipeline}} on {{YYYY-MM-DD}} — {{how: "listed on anthropic.com/careers, form reachable"}}
- **Apply path:** {{confirmed | unconfirmed}}
- **Flags:** {{unverified-filter: sponsorship · evergreen · none}}
- **Profile:** *(mirrors the Summary table's Profile column)*
- **Role family:** {{the family from `search-profile.md` this matched, or `Adjacent` for the open bucket}}
- **Status:** found

### Why it fits
2–4 bullets tying the JD to concrete lines in `jobs/input/profile/master-resume.md`.

### Requirements
- **Must-have:** {{extracted verbatim-ish from the JD}}
- **Nice-to-have:** {{}}
- **Gaps:** {{what the user does not have — be honest, this drives the tailoring}}

### Keywords for ATS
{{comma-separated terms lifted from the JD that the tailored resume should legitimately hit}}

### Application notes
- Account required: {{yes/no}}
- Cover letter: {{required | optional | none}}
- Known screening questions: {{}}
- Gotchas: {{iframe, multi-step wizard, file-type limits, …}}

---
*(repeat per job)*

---

## Pipeline / talent-pool postings
Real but not an open req — general-interest and "future opportunities" listings. Kept
separate so they never compete with live roles in the queue.

| Job ID | Role | Company | Note |
|---|---|---|---|

## Excluded by verification
Recorded so the same listing isn't re-surfaced next run.

| Job ID | Company | Verdict | Why |
|---|---|---|---|
| `example--role` | Example | dead | ATS returned 404 |

## Unverified backlog
Ranked candidates the run never reached, carried forward. The next `/job search` verifies
from this list before sweeping sites again.

| Provisional | Profile | Role | Company | Location | Source | Apply URL | Careers | Status |
|---|---|---|---|---|---|---|---|
