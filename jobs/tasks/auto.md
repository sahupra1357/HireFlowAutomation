# Task: `auto`

Run via `/job` with no arguments, or `/job auto`. The unattended pipeline: read the resume,
work out what to search for, search, pick up the user's inbox, collect JDs, evaluate,
tailor — carrying every job as far as it
can go in one pass and **never stopping the batch because one job is stuck**.

**Tool budget for this task:** Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion, WebSearch, WebFetch, Agent

**This task never fills a form.** It stops at `tailored`. Applying is always attended —
`/job apply <job-id>`, one job, user watching, user submits. Nothing here may route into
`jobs/tasks/apply.md`.

Anything outside that budget means you are running the wrong task — return to `SKILL.md`
and route again rather than reaching for the tool.

---

## The model: reconcile, don't sequence

This is **not** a script that runs eight steps and exits. It is a reconciliation loop over
`jobs/output/jobs.md`: look at what state each job is in, do the next thing that is possible
for it, and leave the ones that can't move.

That single property is what makes the manual-JD loop work. A job that needs a hand-pasted
JD is marked and skipped this run; when the user pastes it, the *next* run sees a JD on disk
where there wasn't one and carries that job straight through to a tailored resume. No
special "resume from where I stopped" mode — every run is the same loop, and the loop is
idempotent.

**Corollary: never re-do finished work.** A job with `Resume: ✓` is done; don't re-tailor it.
A JD already on disk is not re-fetched. Re-running `/job` on a settled workspace should do
almost nothing and say so.

## Pre-pass — Health check, before anything is written

The same two checks as `make doctor`, run first — before the snapshot, before any search,
before a token is spent on the web:

```bash
python3 jobs/bin/check-index.py                 # exit 1 on an ERROR
python3 jobs/bin/check-setup.py --preflight     # never fails; reports only
```

**An index ERROR stops the run.** Every pass below merges into `jobs/output/jobs.md`;
searching into a table with a crooked row, a duplicate job ID, or a ✓ with no file behind it
spreads the damage and costs a search doing it. Print the errors, point at the fix (re-run
the task that owns the column, or roll back: `make history`, then copy the last good
snapshot over `jobs/output/jobs.md`), and stop. Do **not** repair `jobs.md` yourself
here — that is an edit the user should see first. Index WARNs don't stop anything; carry
them into the report.

**Setup findings never stop the run.** No resume, an empty master, no confirmed families —
those are the degraded modes this task already handles below, which is why `--preflight`
prints them as `DEGRADED` and exits 0. Hold every `ERROR` / `DEGRADED` / `WARN` line and
list them once under *Needs you* in the Pass 5 report — a committed-personal-data `ERROR`
first, because it is the only one that gets worse the longer it sits. Don't repeat them
mid-run.

Both scripts take under a second, so run them even when `jobs/bin/daily-run.sh` already
did (it passes `--skip-search`); the state may have changed since.

## Pass 0 — Snapshot, then bootstrap

**Snapshot first, before anything is written.** Passes 1–4 all merge into `jobs/output/jobs.md`;
back it up before the first of them touches it:

```bash
jobs/bin/snapshot.sh
```

It writes `jobs/output/history/jobs-<YYYY-MM-DDTHH-MM-SS>.md` and prints the path. It skips
writing when nothing changed since the last snapshot, and keeps the most recent 50.

**Once per run, here — not per pass and not per job**, or a ten-job run leaves ten
near-identical files. Hold the path it prints and name it in the Pass 5 report, so the user
knows what to roll back to.

This is the undo for the whole run. A bad merge, a truncated table, or a status you got
wrong is one `cp` away from being recovered, and `make web` renders any snapshot from its
version picker. A snapshot taken *after* the damage is worthless — that is the entire reason
this sits in Pass 0 rather than next to the write it protects.

### Then read the state

```bash
ls jobs/input/profile/source-resumes/ 2>/dev/null
grep -c 'TODO' jobs/input/profile/master-resume.md 2>/dev/null
grep -c '^### [0-9]' jobs/input/config/search-profile.md 2>/dev/null
```

### With a resume

- **`master-resume.md` missing or TODO-heavy** → run `jobs/tasks/setup.md` first. It extracts
  the resume and derives the role families.
- **Family list:** `jobs/input/config/setup-families.md` if it exists (user-confirmed at
  setup — use it as is), else the families in `search-profile.md`.
- **Role families still unwritten** (neither file has any) → derive **3–4** from the resume
  per `setup.md` Step 6.
  In unattended mode, write your best inference and **mark it provisional** in
  `search-profile.md` rather than blocking on confirmation:

  ```
  > ⚠️ Role families inferred from the resume on <date>, not yet confirmed by the user.
  > Run `/job setup` to review them.
  ```

  Report at the end that search ran against inferred families — an unreviewed guess about
  what job someone wants is the most expensive thing in this pipeline to get wrong.

### Without a resume — run degraded, don't stop

**A missing resume does not block searching.** It blocks *tailoring*, and only tailoring.
The resume feeds three things: role-family inference, the 25-point skill-overlap dimension of
the fit score, and the tailored resume itself. Everything else — finding postings, verifying
they're live, capturing JDs — needs nothing from it.

So branch on whether the user said what they're looking for:

| `source-resumes/` | Focus argument | What happens |
|---|---|---|
| empty | given (`/job java full stack developer`) | **Run degraded.** Search + JD capture proceed; stop before tailoring. |
| empty | none | **Stop.** There is nothing to infer a search from — a resume or a focus, one of the two. |

**Degraded run, step by step:**

1. **Build one role family from the focus string itself.** Expand it the way `setup.md`
   Step 6 expands a resume-derived family — the literal phrase plus the synonyms and
   seniority variants the market actually uses for it, plus JD-body signals that confirm the
   role is really that work. `java full stack developer` becomes titles like *Full Stack
   Engineer (Java)*, *Java Developer*, *Senior Software Engineer — Java*, with signals like
   *Spring Boot*, *React*, *REST*, *microservices*.

   Write it into `search-profile.md` as a provisional family so the next run reuses it, and
   mark it:

   ```
   > ⚠️ Role family inferred from the search phrase "<focus>" on <date>, with no resume on
   > file. Drop a resume into jobs/input/profile/source-resumes/ and run `/job setup`.
   ```

2. **Score without the skill-overlap dimension.** 25 of the 100 points are unscorable. Do not
   redistribute them and do not pretend the score is complete — scale to 75 and label every
   fit score provisional. Write the warning into the `jobs/output/jobs.md` preamble as a
   blockquote (`make web` renders these in the header):

   ```
   > ⚠️ **Fit scores are provisional.** No resume on file, so the 25-point skill-overlap
   > dimension is unscored. Run `/job setup`, then `/job triage`.
   ```

3. **Run Pass 1 (search) and Pass 3 (JDs) normally.** Both are fully functional. Every hard
   filter that has a value still applies; list the inactive ones as usual.

4. **Stop at Pass 4.** Tailoring without a master resume would mean inventing content, which
   is the one thing this workspace never does. Leave those jobs at `jd-captured` and say so:

   ```
   Found 8 · JDs captured 6 · Tailored 0

   Blocked on your resume
     Drop a resume into jobs/input/profile/source-resumes/ and run /job — the 6 jobs
     with JDs will tailor on the next pass. Nothing found so far is lost.
   ```

Never soften this into "I'll tailor from the job description" or "give me your details and
I'll draft one." No resume means no tailoring.

**Screening answers are not a blocker here.** An empty answer bank only blocks `apply`,
which this task never reaches. Note the gap in the report and carry on.

## Pass 1 — Search

**Country first.** Read **Countries** from *Location & work mode* in `search-profile.md`
(ships as `United States`) and pass it into the search, along with any `--country` override
from this run's arguments. It is a hard filter — see `search.md`. State it in the run header:
`Searching United States · remote, hybrid.`

**Tier order is quality order.** `search.md` sweeps `job-sites.md` top-down: the ATS board
APIs first (live jobs, JD bodies included, nothing dead), then HN, then LinkedIn, and only
then `site:` web search. If a run's results look stale or full of agency reposts, it fell
through to Tier 4 — say so in the report rather than presenting them as equivalent.

Run `jobs/tasks/search.md` when the list is short of the configured **Result limit**, when
`--fresh` was passed, **or when the run carries a focus that no existing row matches** — a
new role family always needs a search, however full the list already is. Skip it only when
the list is full, the focus is one already covered, and nothing is stale — a run that only
needs to clear a JD backlog should not spend six minutes searching.

**`--skip-search` skips this pass unconditionally.** `jobs/bin/daily-run.sh` passes it: the
script has already run `/job search` once per role family on a cheaper model, so this run
starts at Pass 2 and works only on what is already in `jobs/output/jobs.md`. Say
`Pass 1 skipped (--skip-search)` in the report.

### Then the inbox — always, even with `--skip-search`

Jobs the user found themselves are not a search, so no flag skips them. If
`jobs/input/inbox.md` has links under `## Add` that are not in the index yet, run
`jobs/tasks/add.md` with `--inbox --no-jd`: it verifies each one and writes it as
`shortlisted`, and Pass 3 captures its JD with everything else. Pass 2 never demotes them —
the user picked them, which outranks any fit band. **Unattended, `add` asks nothing:** a
hard-filter failure is held back (not added, not excluded) and listed under *Needs you*,
and an unreadable link is listed with its URL.

## Pass 2 — Triage without stopping

`jobs/tasks/triage.md` is normally the checkpoint where the user picks. Unattended, apply the
fit bands from `search-profile.md` instead.

**Read `Auto-tailor threshold` from `search-profile.md`** (under *Fit scoring*). Call it `T`.
If the line is missing, use `75`. It is the one number that decides how many resumes an
unattended run writes, so it lives in config, not here.

| Fit | Unattended action |
|---|---|
| ≥ `T` | auto-shortlist → carry through to `tailored` |
| 40 – (`T`−1) | leave at `found`; list it for the user to promote |
| <40 | leave at `found`; do not spend a JD fetch on it |

Never lower a job that is already past `found`: a row shortlisted or tailored by an earlier
run under a lower threshold keeps its status.

Anything failing a **hard filter** is `skipped` with the reason logged, exactly as in
attended triage. Scam hard-drops never appear at all.

Auto-shortlisting is a scoring decision, not a judgment about whether the user wants the
job. Say so in the report and make the below-threshold list easy to scan, because that band is where
the user's own taste actually matters.

## Pass 3 — JDs, failures included

Run `jobs/tasks/jd.md` for every shortlisted job with `JD: —`. Attended, a failed capture is
a checkpoint. **Here it is not.** On failure:

1. Write the stub to `jobs/output/jds/<job-id>.md` with the posting URL and paste
   instructions, exactly as `jd.md` specifies for `manual-required`.
2. Set the row's **JD** column to `⏳ manual`.
3. Leave **Status** at `shortlisted`.
4. **Move to the next job.** Do not ask, do not retry a third time, do not reconstruct the
   JD from the title or the company's marketing site.

Collect every failure and report them **once, at the end**, with all the URLs together.

**Pick up what the user pasted.** Before fetching anything, re-check every `⏳ manual` row —
this is the other half of the loop:

```bash
ls jobs/output/jds/*-source.* 2>/dev/null
grep -Ll 'manual-required' jobs/output/jds/*.md 2>/dev/null
```

A stub that now has real content, or a `<job-id>-source.*` file the user dropped in, means
the block is cleared. Flip **JD** to `✓ <date>`, set Status to `jd-captured`, and let the job
flow into Pass 4 this run. **This is the behaviour that makes the second run finish what the
first one couldn't** — verify it explicitly rather than assuming it happened.

## Pass 3b — Evaluate

Run `jobs/tasks/evaluate.md` for every job with `JD: ✓`, `Resume: —` and no
`jobs/output/applications/<job-id>/evaluation.md` yet — the same set Pass 4 is about to
tailor. **Serially, in this session** — it is not one of the fan-out tasks, and its only
web use is two public pay-data queries per job. Skipped entirely with `--no-evaluate`, and
for every job when there is no master resume (evaluate has the same preflight as tailor).

This is the one place the unattended run takes the evaluation's advice: a job whose verdict
is **`skip`** is **not tailored** this run. Its Status stays where it is, and it is listed
under *Needs you* with the deciding reason, so the user can overrule it with
`/job tailor <job-id>`. `apply-with-caveats` is tailored as normal, with the caveat in the
report. A job that already has `evaluation.md` is not re-evaluated — the loop never re-does
finished work.

## Pass 4 — Tailor

Run `jobs/tasks/tailor.md` for every job with `JD: ✓` and `Resume: —`, except those whose
evaluation says `skip` (Pass 3b). Fan out one subagent
per job, **max 4 in flight**, each confined to its own `jobs/output/applications/<job-id>/`.
The parent is the only writer of `jobs/output/jobs.md`.

Reject any subagent result that reports a rule violation or comes back empty: leave that job
untailored, say so, and keep going. One bad tailoring does not stop the others.

## Pass 5 — Write the state and report

Update `jobs/output/jobs.md` — Status plus the JD / Resume / Form columns — and mirror into
`jobs/output/tracker.md`. This file is what `make web` reads, so it must be accurate before
you print anything.

**Then prepend an entry to `jobs/output/run-log.md`** — newest first — with the snapshot
name, what was searched, the country filter in force, counts, sites unswept, filters
inactive, and anything durable the run learned. `jobs.md` carries *state*; the run log
carries *how that state came to be*, and it is where the narrative detail belongs so the
dashboard can stay a board rather than a changelog. Keep `jobs.md`'s own preamble to the few
lines that are still true right now.

Then report, short:

```
Pipeline run <date>   (snapshot: jobs-2026-09-07T09-42-25.md)

  Searched      3 families (inferred) · 12 found · 4 new
  Inbox         2 added from jobs/input/inbox.md · 1 unreadable
  Shortlisted   7 auto (fit 75+) · 5 left at found (40–74, your call)
  JDs           5 captured · 2 need you
  Evaluated     6 — 4 apply · 1 with caveats · 1 skip (not tailored)
  Tailored      5 resumes written · 0 failed

  Needs you
    Setup: 1 work-authorization answer still TODO (make doctor for the full list)
    2 JDs to paste:
      acme--data-engineer   https://...
      globex--ai-engineer   https://...
    5 jobs in the 40–74 band — /job triage to promote any
    1 evaluated skip: initech--staff-analyst — requires CPA licence
      (applications/initech--staff-analyst/evaluation.md · /job tailor it to overrule)
    Answer bank incomplete — /job setup before your first apply

  Ready to apply: 5     /job apply <job-id>     (one at a time, you submit)
```

Lead with what needs the user, not with what succeeded. **Never imply anything was
submitted** — this task cannot submit and cannot fill a form.

## Guardrails

- **Never apply.** Not even for a single high-fit job. `filled-awaiting-user` is unreachable
  from here.
- **Never invent a JD.** A missing JD is a `⏳ manual` row, always.
- **Never fabricate resume content**, and never relax that because the run is unattended and
  nobody is watching the individual bullets.
- **Never re-run a completed stage** for a job that already has the artifact.
- **Never tailor without `master-resume.md`.** A degraded run ends at `jd-captured`.
- **Stay inside the time budget** in `search-profile.md`. On hitting it, write everything
  done so far and report what was left — the next run picks it up.
