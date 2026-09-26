# How HireFlow works

A step-by-step walk through what happens from "I dropped my resume in" to "a filled
application is waiting for me to click Submit." For the file-by-file reference, see
`CLAUDE.md`; for install and quick start, see `README.md`.

---

## The big picture

HireFlow is not an app with its own AI. **Claude Code is the agent.** The repo holds:

- **One skill** — `/job` (`.claude/skills/job/SKILL.md`). It is only a router: it reads your
  request, picks a task, and loads that task's instructions.
- **Nine task files** — `jobs/tasks/*.md`. Each is a plain-Markdown procedure the agent
  follows (`setup`, `search`, `triage`, `jd`, `tailor`, `apply`, `status`, `interviews`,
  `auto`). You can edit them; changes apply on the next run.
- **Helper scripts** — `jobs/bin/` (snapshots, form filling, PDF rendering, daily loop).
- **A dashboard** — `jobs/dashboard/` (Go), started with `make web`.

Everything is split into **what you provide** (`jobs/input/`) and **what the agent produces**
(`jobs/output/`). Skills read `input/` and write `output/`, never the reverse.

```
 you drop a resume
        │
        ▼
 ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌──────┐   ┌─────────┐   ┌─────────┐
 │  setup  │──▶│ search  │──▶│ triage  │──▶│  jd  │──▶│ tailor  │──▶│  apply  │──▶ YOU click Submit
 └─────────┘   └─────────┘   └─────────┘   └──────┘   └─────────┘   └─────────┘
   profile      jobs.md       shortlist    jds/*.md   resume.pdf    filled form
                                                                    (never submitted)
```

Each stage **refuses to start until the previous stage's file exists on disk**, so a failure
never half-produces the next stage, and any stage can be re-run on its own.

---

## Step 1 — Setup: turn your resume into a profile (`/job setup`)

**You do:** put your resume (PDF / DOCX / MD) in `jobs/input/profile/source-resumes/` and run
`/job setup`.

**The agent does:**

1. Checks what already exists so it doesn't overwrite your edits.
2. Extracts the resume text.
3. Writes `jobs/input/profile/master-resume.md` — the **single source of truth** for your
   identity and experience. Every tailored resume later is a *subset* of this file.
4. Records links (LinkedIn, GitHub, portfolio) in `jobs/input/profile/links.md`.
5. Seeds the screening-answer bank `jobs/input/config/application-answers.md` (work
   authorization, sponsorship, notice period, salary…). Anything it can't source is left
   for you — it never guesses.
6. **Proposes 2–5 role families** from your history (titles + synonyms + JD "signals"), plus
   an *Adjacent* bucket, and asks you to confirm or edit them. Only the families you confirm
   go into `jobs/input/config/setup-families.md` — that is what every search uses.
7. Fills keywords, comp floor, locations, and employer-quality definition in
   `jobs/input/config/search-profile.md`.
8. Measures your resume's layout (`resume-format.md`) and, for a `.docx`, maps where each
   section sits (`resume-docx-map.json`) so tailored resumes come out in **your own format**.

The workspace has no built-in field — engineering, design, finance, whatever. What counts as
a relevant job comes entirely from the families you confirmed here.

---

## Step 2 — Search: find candidate jobs (`/job search`)

1. **Load context** — active role families, countries (a hard filter, default United
   States), keywords, sites from `jobs/input/config/job-sites.md`, and every job ID already
   known (for dedupe).
2. **Build queries** per role family.
3. **Collect candidates**, preferring official paths:
   - ATS APIs (Greenhouse / Lever / Ashby) — primary; they return the full JD and only list
     open reqs.
   - The Muse, Hacker News "Who is Hiring?", LinkedIn via Apify, web search, and the browser
     as fallbacks.
4. **Pre-rank** the pool cheaply (title match 40, employer quality 20, location 15,
   recency 15, source tier 10) so verification time goes to the best candidates first.
5. **Verify down the ranking** until the limit is hit: is the posting live, is it on the
   employer's own careers page, is the employer real, any ghost-job / scam red flags?
6. **Score 0–100** (the fit score), penalise anything unverified, **dedupe** by job ID
   (`<company-slug>--<role-slug>`), and apply hard filters — failures go to *Excluded* with
   a reason.
7. **Merge** results into `jobs/output/jobs.md` (after `jobs/bin/snapshot.sh` archives the
   previous version), stamping each row's **Profile** column with the family that found it.

---

## Step 3 — Triage: pick what's worth pursuing (`/job triage`)

Works from disk only — no browser. Re-scores each `found` job against your actual resume,
presents the ranked list, and marks your picks `shortlisted`. Hard-filter failures become
`skipped`.

---

## Step 4 — JD capture: get the full job description (`/job jd`)

For each shortlisted job, climb an **escalation ladder** until one rung works:

1. **ATS API** (no browser).
2. **The employer's Careers page** in the browser — also the best liveness proof.
3. **The Apply URL** in the browser (proves the page exists, but liveness stays
   `inconclusive`).
4. **`manual-required`** — writes a stub to `jobs/output/jds/<job-id>.md` and marks the row
   `⏳ manual`.

The result is always one file: `jobs/output/jds/<job-id>.md`. Everything downstream reads the
JD from there, never from a live page.

**If a job lands on `⏳ manual`:** paste the JD into that stub, or drop the posting as
`jobs/output/jds/<job-id>-source.pdf` (or `.html`, `.txt`, `.docx`, screenshot). The next run
picks it up. The agent **never** invents a JD from the title or the company's website.

---

## Step 5 — Tailor: rewrite the resume for one job (`/job tailor <job-id>`)

**Preflight — two hard stops:** the master resume must be populated, and the JD must be real
(not missing, not a stub).

Then:

1. Analyse the JD — must-haves, nice-to-haves, the vocabulary it uses.
2. Map each requirement to evidence in `master-resume.md`; **flag gaps** instead of filling
   them.
3. Write `jobs/output/applications/<job-id>/resume.md` — re-ordered, re-weighted, rephrased
   in the JD's language. **Nothing new**: no invented employers, dates, titles, degrees,
   certifications, tools, or metrics.
4. `jobs/bin/verify-tailored.py` checks for invented emails, figures, years, or
   certifications not present in the master.
5. `make pdf` renders it: with a mapped `.docx` it edits a copy of **your own Word file** and
   exports `resume.docx` + `resume.pdf`; otherwise it renders in the layout measured from
   your source resume. Only the words change between jobs.
6. Writes a cover letter if the application wants one.

`/job tailor --all` does this for every job that has a JD, one subagent per job, at most 4 at
once. Each subagent writes only inside its own `applications/<job-id>/` folder; the main
session is the only writer of `jobs.md`.

---

## Step 6 — Apply: fill the form, stop before Submit (`/job apply <job-id>`)

**The single most important rule: the agent never submits.**

Attended mode (one job, you watching):

1. **Preflight** — tailored resume exists; job not already in `jobs/output/tracker.md`
   (never apply twice).
2. Re-read the posting.
3. **Map the form** — every field, its type, whether it's required, where its value comes
   from. Greenhouse / Lever / Workday often use iframes and react-select comboboxes.
4. **Authentication** — if an account is needed, it asks first. CAPTCHA / MFA / SSO → hands
   the browser to you (`$B handoff`) and waits.
5. **Fill** from `master-resume.md` and `application-answers.md` only. A screening question
   with no stored answer → it asks you, then saves your answer back to the bank.
6. **Verify** — no validation errors, resume attached, screenshot saved.
7. **Hand off** — Status becomes `filled-awaiting-user`. You review and click Submit.
8. After you submit, run `make submitted JOB=<job-id>` — the only thing that sets `submitted`.

### Batch mode (`/job apply --batch`)

Fills every tailored job serially, **each in its own tab of a visible browser that stays
open**. Differences from attended mode: it never asks (unknown answers stay blank and are
listed as gaps), it skips jobs that need login/CAPTCHA instead of waiting, and it saves:

- `form-fill.json` — the field map, so the form never has to be mapped twice;
- `refill.js` — a console script to refill the form if the window was closed.

`jobs/bin/fill-form.py` replays `form-fill.json` into the browser (including the real
`resume.pdf` upload). Neither script clicks a submit control.

---

## Step 7 — Track (`/job status`)

Read-only board of every application: status, follow-ups due, patterns (e.g. who's
ghosting). You log outcomes here — `interviewing`, `offer`, `rejected`, `ghosted`.

**Status lifecycle:**

```
found → shortlisted → jd-captured → tailored → filled-awaiting-user → submitted
                                                     ▲ agent's ceiling   ▲ only you set this
      → interviewing → offer / rejected / ghosted          (or skipped at any point)
```

---

## Bonus — Interview intel (`/job interviews <company>`)

Pulls real interview experiences (Reddit, Hacker News, configured sites) for companies in
your index, scores them for credibility, and writes prep briefs to
`interviews/output/reports/<company-slug>/`. It reads `jobs.md` but never writes to it.

---

## Running it all at once

### `/job` — the auto pipeline

Bare `/job` runs `jobs/tasks/auto.md`: resume → role families → search → triage → JDs →
tailor. It **stops at `tailored`** and never fills a form.

It is a **reconciler, not a script**: each run looks at every job's current state and does
the next possible thing.

- Fit ≥ **Auto-tailor threshold** (default 75, set in `search-profile.md`) → shortlisted and
  carried to a tailored resume.
- Fit 40 – threshold → left at `found` for you to promote.
- A failed JD capture → `⏳ manual`, the run moves on, and all failures are listed once at
  the end.
- Next run: anything you unblocked (pasted a JD) continues; finished work is not redone.

`/job java full stack developer` (a role phrase) works even with no resume: it searches on
that phrase, captures JDs, marks fits provisional, and stops before tailoring.

### `make daily` — the whole loop, one command

| Stage | What | Model (override via env) |
|---|---|---|
| 1 | `/job search` once per confirmed family (or one generic search if setup hasn't run) | `SEARCH_MODEL` = sonnet |
| 2 | `/job auto --skip-search` — triage, JDs, tailoring | `REASON_MODEL` = opus |
| 3 | `/job apply --batch` — fill new forms in a visible browser | `APPLY_MODEL` = sonnet |
| 4 | `jobs/bin/morning-run.sh` — reopen every unsubmitted filled form | free, no agent |

Locked so two runs can't edit `jobs.md` at once; logs go to
`jobs/output/logs/daily-<date>.log`. Skip stages with `--no-search`, `--no-tailor`,
`--no-apply`.

### `make morning`

Reopens every mapped, unsubmitted application, filled, one tab each. No agent, no tokens.
You review the tabs and submit the ones you want.

---

## Seeing the results — `make web`

Starts the dashboard at http://localhost:8080 (stop with `make stop`). It reads
`jobs/output/jobs.md` and shows:

- the job table with **Status / JD / Resume / Form** columns (✓ links open the JD, resume
  PDF, and form screenshots);
- a **Pipeline** card (JD / resume / form counts);
- a **Needs you** panel — every `⏳ manual` JD with its posting link;
- a version picker over `jobs/output/history/` snapshots (the index's only undo).

---

## Safety rails, in one place

| Rule | How it's enforced |
|---|---|
| Never submit | Agent stops at review screen; `fill-form.py` / `refill.js` click no submit control; only `make submitted` sets `submitted` |
| Never fabricate | Tailoring is a subset of `master-resume.md`; `verify-tailored.py` flags new figures/years/certs |
| Never guess screening answers | Only from `application-answers.md`; otherwise ask (attended) or leave blank (batch) |
| Never invent a JD | Tailor refuses without a real `jds/<job-id>.md` |
| Never apply twice | `tracker.md` checked before every apply |
| No accounts / payments without asking | Attended apply asks; batch skips |
| CAPTCHA / MFA / bot walls | Hand the browser to you, don't fight |
| Your data stays local | `jobs/input/profile/` goes only to the employer's own form |
| Undo | `jobs/bin/snapshot.sh` before every edit to `jobs.md`; `make history` lists them |

---

## Typical first session

```bash
# 1. put your resume in jobs/input/profile/source-resumes/
# 2. in Claude Code:
/job setup                 # confirm your role families
/job                       # search → JDs → tailored resumes
# 3. in a terminal:
make web                   # review jobs, resumes, "Needs you"
# 4. back in Claude Code, for a job you like:
/job apply <job-id>        # fills the form, stops before Submit
# 5. you click Submit, then:
make submitted JOB=<job-id>
```

After that, `make daily` (or a scheduled run) keeps the loop going, and `make morning` puts
the filled forms in front of you.
