# Task: `evaluate`

Run via `/job evaluate <job-id>`, `/job evaluate --all`, or as part of `/job auto`.

Write a readable assessment of one role against the user's real experience — what the job
is, which requirements the master resume actually evidences, the honest gaps, how to
position for the level, what it pays, what tailoring should emphasise, the interview
stories to prepare, and a verdict: apply, apply with caveats, or skip. Produces
`jobs/output/applications/<job-id>/evaluation.md`. Use when the user asks "is this job
worth it", "evaluate", "assess", "should I apply", or wants interview prep for one role.

**Tool budget for this task:** Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion, WebSearch, WebFetch

**No browser, no subagents, no form.** The JD is already on disk; the only thing fetched
from the web is public compensation data, and **the only thing sent in a query is the role
title, company, level and location** — never a line of the resume (hard rule 7).

Anything outside that budget means you are running the wrong task — return to
`SKILL.md` and route again rather than reaching for the tool.

---

The fit score in `jobs/output/jobs.md` is one number; this is the reasoning behind it,
written down so the user can decide in two minutes instead of re-reading the JD. It sits
between `jd` and `tailor`: it needs a real JD, and its tailoring plan is what `tailor` reads
first.

**It is advice.** It never changes Status, never skips a job by itself, and never touches
the JD / Resume / Form columns. The user (or `auto`, under its own rule) decides.

**Arguments:**

| Arg | Scope |
|---|---|
| `<job-id>` | that job |
| `--all` | every Summary row with **JD ✓** and no `evaluation.md` yet |
| `--refresh` | re-evaluate even if `evaluation.md` exists (e.g. after `/job setup` changed the master) |
| `--no-comp` | skip the web research in section D; write "not researched" |

---

## Step 0 — Preflight: the same two hard stops as `tailor`

An evaluation built on a missing resume or a missing JD can only be invented.

1. **Master resume populated?** If `jobs/input/profile/master-resume.md` is missing, still
   says `STATUS: not yet populated`, or has TODO Experience/Skills → stop:
   `Can't evaluate without your resume — drop it into jobs/input/profile/source-resumes/ and run /job setup.`
2. **JD real?** `jobs/output/jds/<job-id>.md` missing → `Run /job jd <job-id> first.`
   `Capture method: manual-required` → stop and give the path to drop the posting at.
   **Never evaluate against the short summary in `jobs/output/jobs.md`.**

For `--all`, a job failing either check is listed in the report and skipped; the batch goes on.

## Step 1 — Load

```bash
cat jobs/output/jds/<job-id>.md                 # the real input
cat jobs/input/profile/master-resume.md         # the only evidence allowed
cat jobs/input/templates/evaluation.md          # the output shape
grep -n '<job-id>' jobs/output/jobs.md          # Summary row: fit, profile, comp, verified
sed -n '/^## Compensation/,/^## /p;/^## Fit scoring/,/^## Volume/p' jobs/input/config/search-profile.md
```

Also read the job's `## Details` block in `jobs/output/jobs.md` (its *Gaps* line was written
at search time and is a starting point, not a conclusion) and the active family list per
`SKILL.md`.

## Step 2 — Write each section

Follow `jobs/input/templates/evaluation.md` section by section. The rules that make it
worth reading:

**A) Role summary.** What the JD says, in its own words where they matter. "Seniority" cites
the phrase that implies it ("owns the roadmap for…", "8+ years"). Unknown is `—`.

**B) Requirement match — the core of the report.** One row per requirement in the JD's
*Requirements — extracted* section (must-haves first). The *Evidence* cell **quotes the
master resume** and names where the line is. Scoring:
- `✓ strong` — the master states it directly, at the level asked.
- `~ partial` — related, smaller scale, older, or a prototype where production is asked.
  Say which.
- `✗ none` — not in the master. **Do not hunt for a form of words that implies it.**

Every `~` and `✗` gets a *Gaps* row. **Blocker?** is judged from the JD's language: a
licence, clearance, degree or work-authorisation requirement stated as mandatory is a
*hard blocker*; a must-have skill with nothing adjacent is a *likely screen-out*; the rest is
*nice-to-have*. *Honest mitigation* is only what the user can truthfully say — adjacent
experience that is really in the master, or "none — leave it out".

**C) Level & positioning.** Compare the JD's level with what the master shows (years, scope,
titles held). *Lead with* is the 2–3 strongest true facts. *Don't claim* lists what the JD
rewards that the master doesn't support — `tailor` and `apply` read this to stay honest.

**D) Comp & demand.** The posted figure first. Then, unless `--no-comp`, at most **two**
WebSearch queries — role title + level + location (+ company), nothing else — for public
figures (Levels.fyi, Glassdoor, Blind, the posting's own pay-transparency range). Every
figure has its source and date. **No data → "no public data found".** Never estimate a
number. Compare with the user's floor from `search-profile.md`.

**E) Tailoring plan.** Concrete instructions for `tailor`: which bullets to promote, which
section leads, which JD vocabulary to adopt — each one re-weighting something already in the
master. *Keywords to hit* are only those the master supports; *Keywords to leave out* are
the ones it doesn't, so `tailor` doesn't stuff them.

**F) Interview prep.** 3–6 stories drawn **only** from the master, each mapped to a JD
requirement. S/T/A/R come from the master's own bullet; a part it doesn't record is `—` for
the user to fill in. *Reflection* is `—` unless the master or the user said it. Then the
questions the gaps will provoke, each with a truthful answer.

**G) Verdict.**

| Verdict | When |
|---|---|
| `skip` | any **hard blocker**, or re-scored fit < 50, or ≥ half the must-haves `✗` |
| `apply-with-caveats` | a *likely screen-out* gap, comp below the floor, liveness `inconclusive`, or a level mismatch of one step |
| `apply` | none of the above |

Say `skip` plainly — "this is a weak match; applying is unlikely to get a response" — and
give the reason. An application a recruiter reads and rejects in six seconds costs the user
a slot that could have gone to a real match. The user can still apply; it's their call.

Re-score fit against the full JD with the `search-profile.md` weights and put it in the
header next to the index's number. A difference of 10+ points gets one line explaining why.

**Never fabricate, anywhere in the report.** Same rule as tailoring: nothing about the user
appears that is not in `master-resume.md`. An evaluation that flatters is worse than none —
it sends a tailored resume and an application at a job that was never a match.

## Step 3 — Write and record

1. Write `jobs/output/applications/<job-id>/evaluation.md` (create the folder if needed).
2. `jobs/bin/snapshot.sh` once per run, then append to the job's `## Details` block in
   `jobs/output/jobs.md`:
   `- **Evaluation:** <verdict> · fit <N> · <YYYY-MM-DD> → applications/<job-id>/evaluation.md`
   (replace an existing *Evaluation* line rather than adding a second).
   **Nothing else in `jobs.md` changes** — not Status, not Fit, not the stage columns.
3. Append one line to `jobs/output/applications/<job-id>/log.md`:
   `<date> evaluated — <verdict>, fit <N>, gaps: <short list>`.

The dashboard's Status pill for the job now opens the evaluation (and the resume, once one
exists).

## Step 4 — Report

```
Evaluated 4

  apply               acme--platform-engineer        fit 86  (index 84)  7/8 must-haves
  apply-with-caveats  globex--data-engineer          fit 71  (index 78)  comp below floor
  skip                initech--staff-analyst         fit 44  (index 70)  requires CPA licence
  —                   hooli--ml-engineer             JD still ⏳ manual — drop the posting first

  Reports: jobs/output/applications/<job-id>/evaluation.md  (or click the Status pill in make web)
  Next: /job tailor acme--platform-engineer · /job tailor globex--data-engineer
```

Keep it to this. The reasoning is in the files.
