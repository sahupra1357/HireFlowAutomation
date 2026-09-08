---
name: job
description: The complete job-search agent — one skill, routed by subcommand. Sets up the profile from a resume, searches the configured sites for jobs, triages and shortlists them, collects full job descriptions, tailors the resume and cover letter to a specific role, fills out application forms in the browser (stopping before submit), tracks application status and follow-ups, and finds real interview experiences. Use whenever the user mentions job search, finding roles, resumes or CVs, tailoring, job descriptions, applying to a job, application status, follow-ups, or interview prep — and for "/job ..." in any form.
allowed-tools: Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion, WebSearch, WebFetch, Agent
---

# /job

One skill, eight tasks. This file is the **router** and holds no procedure of its own: work
out which task the request is, read that task's file from `jobs/tasks/`, and follow it.

**The task files live in the workspace, not in this skill** — `jobs/tasks/*.md`, alongside
`jobs/input/` and `jobs/output/`. They are editable workspace content: the user can tune a
task's steps without touching the skill, and a task file changed on disk takes effect on the
next run. Always read the file fresh; never work from memory of what a task used to say.

```
/job                      # the auto pipeline: resume → families → search → JDs → tailored
/job <task> [args]        # run one task
/job status               # read-only board, starts nothing
```

**Bare `/job` does work.** With no argument it runs `jobs/tasks/auto.md` — the unattended
pipeline that carries every job as far as it can go and never stops the batch for one stuck
job. It stops at `tailored`; it cannot fill a form. For a read-only look, `/job status`.

## Routing table

| Task | Aliases | Instructions | Args |
|---|---|---|---|
| `setup` | `profile`, `resume-setup` | `jobs/tasks/setup.md` | — |
| `search` | `find` | `jobs/tasks/search.md` | focus, `--limit N`, `--country "<list>"` |
| `triage` | `shortlist`, `pick` | `jobs/tasks/triage.md` | fit floor |
| `jd` | `jds`, `collect`, `describe` | `jobs/tasks/jd.md` | job-id or `--all` |
| `tailor` | `resume`, `cv` | `jobs/tasks/tailor.md` | job-id or `--all` |
| `apply` | `fill` | `jobs/tasks/apply.md` | job-id **(required)**, or `--batch [job-id ...]` |
| `status` | `track`, `tracker`, `board` | `jobs/tasks/status.md` | — |
| `interviews` | `ix`, `prep` | `jobs/tasks/interviews.md` | company |
| `auto` | *(no argument)*, `run`, `all`, `pipeline` | `jobs/tasks/auto.md` | `--fresh`, `--limit N`, `--country "<list>"` |

Paths are relative to the workspace root. Resolve them from the repo root, not from the
skill directory:

```bash
_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
ls "$_ROOT"/jobs/tasks/
```

**Read the task file before doing anything.** If the file named in the table is missing, say
so and stop — do not improvise the task from this router's one-line summary of it.

### How to route

1. **Explicit task word** → match case-insensitively against names and aliases.
2. **A bare job ID** (`<company-slug>--<role-slug>`) → "advance this job one stage." Read its
   row in `jobs/output/jobs.md` and pick by what's missing: no JD → `jd`, JD but no resume →
   `tailor`, resume but no form → `apply`.
3. **A bare role phrase** — no task word, no job ID, just a kind of job
   (`/job java full stack developer`, `/job remote staff SRE`) → run `auto` with that as the
   **focus**. This is the common first-time invocation and it must work with an empty
   workspace: `auto` builds a role family out of the phrase and searches on it, with or
   without a resume on file.
4. **Other natural language** → read it as intent and pick the task it describes
   ("what's still live?" → `jd`, "who's ghosting me?" → `status`, "make my resume fit this"
   → `tailor`).
5. **Genuinely ambiguous** → ask, don't guess. Picking `apply` when they meant `tailor`
   wastes a form; picking `search` when they meant `triage` wastes ten minutes.

Say which task you chose in one line before you start. **Pass arguments through verbatim** —
`/job search staff engineer --limit 25` runs `jobs/tasks/search.md` with
`staff engineer --limit 25`. Never reinterpret a flag.

## Task isolation — read this before every task

These eight used to be eight separate skills, each with its own `allowed-tools`. They are one
skill now, so **the tool list above is the union of all eight and no longer constrains any
single task.** The separation is real and still mandatory; it is just enforced by you rather
than by the harness. Each task file in `jobs/tasks/` opens with its own **tool budget** — treat
it as if it were the only tool list you had.

The three that matter most, because the harness can no longer stop you:

- **`triage` and `status` must not touch the browser or the web.** They decide from what is
  on disk. Wanting a live page means you should be running `jd`.
- **`tailor` is the only task that may spawn subagents**, max 4, each confined to its own
  `jobs/output/applications/<job-id>/`. `apply` never spawns one — a form is filled with the
  user watching.
- **`apply` is the only task that may fill a form**, one job at a time, and it stops at the
  review screen. Every other task drives `$B` read-only.

Reaching for a tool outside the current task's budget means you have drifted into another
task. Stop, come back here, and route again — do not carry the tool across.

## Sequencing gates

Tasks refuse to start until the previous stage's artifact exists on disk. Surface the refusal
to the user; never work around it.

| Task | Requires |
|---|---|
| `search` | `jobs/input/profile/master-resume.md` filled (else run `setup`) |
| `triage` | rows in `jobs/output/jobs.md` |
| `jd` | a shortlisted row |
| `tailor` | `jobs/output/jds/<job-id>.md` present and not a `manual-required` stub |
| `apply` | `jobs/output/applications/<job-id>/resume.*` present |

A failure at one stage must never half-produce the next one.

## No argument — run the pipeline

`/job` with nothing after it means `auto`. Read `jobs/tasks/auto.md` and follow it.

It is safe to run repeatedly and on a workspace in any state: it reconciles rather than
sequences, so it skips finished work, picks up jobs the user unblocked since the last run,
and does nothing much on a settled workspace. It never fills or submits a form.

**A missing resume is not a reason to refuse.** It blocks tailoring only. Given a focus —
`/job java full stack developer` — `auto` searches and captures JDs on that phrase alone and
stops before tailoring, saying what it needs. Bare `/job` on a workspace with neither a
resume nor a focus is the only case that stops early: there is nothing to search for.

Only when the user asks for a *read* — "where do things stand", `/job status` — do this
instead of starting anything:

```bash
grep -c '^| [0-9]' jobs/output/jobs.md 2>/dev/null
grep -oE '\|\s*(found|shortlisted|jd-captured|tailored|filled-awaiting-user|submitted|interviewing|offer|rejected|ghosted|skipped)\s*\|' jobs/output/jobs.md 2>/dev/null | tr -d '| ' | sort | uniq -c
grep -c '⏳ manual' jobs/output/jobs.md 2>/dev/null
```

## The pipeline lives in `auto.md`, not here

Do not re-implement it. `jobs/tasks/auto.md` owns the pass order, the fit-band thresholds
that stand in for attended triage, the non-blocking JD failure path, and the resume-from-
manual-JD behaviour that makes a second run finish what the first one couldn't.

The one rule the router enforces on top: **`auto` never reaches `apply`.** Applying is
attended by default — one job, `/job apply <job-id>`, user watching, user submits. The
unattended form (`/job apply --batch`) is still never reached by `auto`: it has to be asked
for by name, and it stops at a filled form like every other path. If a pipeline run
seems to want to fill a form, it has misread its own task file.

## Snapshot before writing the index

Any task that edits `jobs/output/jobs.md` — `search`, `triage`, `jd`, `tailor`, `apply`,
`auto` — runs `jobs/bin/snapshot.sh` **once, before its first edit**. It archives the index
to `jobs/output/history/jobs-<timestamp>.md`, no-ops when nothing has changed, and keeps the
last 50. `make web` renders any of them from its version picker.

The index is a single Markdown file that every stage merges into; a snapshot is the only
thing standing between a bad merge and a lost job hunt.

## Rules that outrank every task file

1. **Never submit an application.** Fill everything, then stop at the review screen.
2. **Never fabricate.** Tailoring re-orders and re-phrases what is already in
   `master-resume.md`. It never invents an employer, date, title, degree, or metric.
3. **Never guess a screening answer.** Only from `jobs/input/config/application-answers.md`;
   if it isn't there, ask, then persist the answer back into that file.
4. **Never invent a job description.** If `jobs/output/jds/<job-id>.md` is missing or a stub,
   stop and ask for the manual download.
5. **Never apply twice.** Check `jobs/output/tracker.md` before starting.

`CLAUDE.md` carries these in full, plus the account/payment, handoff, and site-respect rules.
