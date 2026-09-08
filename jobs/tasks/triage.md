# Task: `triage`

Run via `/job triage`. Formerly the standalone `job-triage` skill.

Review the jobs in a search results file with the user, re-score them against the resume, and shortlist which ones to apply to. Use when the user wants to go through search results, pick jobs, or decide what to apply for.

**Tool budget for this task:** Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion

**No browser, no web search.** This task reads what is already on disk and decides; if you need a live page, the answer is to run `jd` instead.

Anything outside that budget means you are running the wrong task — return to
`SKILL.md` and route again rather than reaching for the tool.

---

Turn a raw search file into a ranked, decided shortlist. This is the checkpoint between
"the agent found things" and "the user spends real effort applying."

**Optional argument:** a fit floor (`/job triage 70`). Reads `jobs/output/jobs.md`.

## Step 1 — Load

```bash
cat jobs/output/jobs.md
cat jobs/output/tracker.md
```

Read `jobs/output/jobs.md` (the **Summary** table plus the matching `## Details` blocks), the tracker,
and `jobs/input/profile/master-resume.md` — fit re-scoring needs the real skills, not the summary.
Triage only rows whose **Status** is `found`; anything further along is already in flight.

## Step 2 — Re-score against the actual resume

`/job search` scores from JD text and a quick resume read. Now do it properly: for each job
with fit ≥ 40, compare the JD's must-haves line by line against the master resume.

For each job produce:
- **Match:** requirements the user demonstrably meets, with the resume line as evidence
- **Gap:** requirements they don't — and whether it's *disqualifying* (5 years of a language
  they've never used) or *bridgeable* (a tool that's a week's ramp)
- **Adjusted fit**, with a one-line reason when it moves more than 10 points

Down-weight aggressively for: title inflation vs. actual JD scope, "senior" roles wanting
10+ years, and postings whose requirements read like three jobs stapled together.

## Step 3 — Present the decision

Show a compact ranked table — job, company, fit, the single biggest gap, comp — then use
`AskUserQuestion` to work through the top candidates in batches of 3–4:

- **Shortlist** — queue it for `/job jd`, then tailoring and applying
- **Skip** — record why; it will not resurface in future searches
- **Maybe** — keep in the file, no action

For anything failing a hard filter but scoring 75+, surface it explicitly with the conflict
named ("no sponsorship, but you said sponsorship required") and let the user decide. Don't
silently drop a strong match.

Respect the queue cap in `search-profile.md` (default 10). If the user shortlists more, say
so and suggest an order rather than refusing.

## Step 4 — Order the queue

Rank the shortlist by what to do first, weighing: fit, posting age (older = closer to
filled), application effort (Lever/Ashby ~15 min, Workday ~45), and whether a referral path
exists. Say the reasoning in one line per job.

## Step 5 — Persist

- Update the **Status** column in `jobs/output/jobs.md` → `shortlisted` or `skipped`. Edit that cell in
  place and change nothing else on the row — in particular, leave the **JD / Resume / Form**
  columns alone; they belong to the later stages. For `skipped`, record the reason in the job's
  `## Details` block so it is never re-surfaced without context.
- Add each shortlisted job to `jobs/output/tracker.md` Active, status `shortlisted`
- Create `jobs/output/applications/<job-id>/` and write `log.md` from `jobs/input/templates/application-log.md`,
  seeded with what's known so far
- Update the tracker Stats line

## Step 6 — Report

```
Triaged 22 jobs → 6 shortlisted, 12 skipped, 4 maybe

Queue (recommended order):
  1. anthropic--forward-deployed-engineer   91  Greenhouse ~15min  posted 3d ago
  2. sierra--ai-engineer                    86  Ashby      ~15min  posted 6d ago
  3. databricks--data-engineer              78  Workday    ~45min  account needed

  Next: /job jd --shortlisted     (grabs the full JD for all 6)
```

Then the pipeline is `/job jd` → `/job tailor --all` → `/job apply <job-id>`.
