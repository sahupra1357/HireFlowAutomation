# Job Application Workspace

Claude Code **is** the agent for this workspace. It finds jobs across configured
sites, records them as reviewable Markdown, tailors the user's resume to a specific job
description, and drives the browser to fill out the application form — stopping short of
submitting so the user reviews and clicks Submit themselves.

**The workspace has no field of its own.** What counts as a relevant job comes entirely from
the role families in `jobs/input/config/search-profile.md`, which `/job setup` derives from
the user's resume. No skill names a role, an industry, or a title — if you find yourself
assuming this is an engineering workspace (or any other kind), re-read the search profile.

The user is **whoever dropped their resume into `jobs/input/profile/source-resumes/`**. Their name,
contact details, and history come from that file via `/job setup` and live in
`jobs/input/profile/master-resume.md` and `jobs/input/config/application-answers.md` — never hardcoded anywhere
in this repo. Everything in `jobs/input/profile/` is their real personal data.

---

## Hard rules — never break these

1. **NEVER submit an application.** (This binds the generated `refill.js` too — it clicks no
   buttons at all.) Fill every field, upload every file, pass every
   validation — then STOP at the final review screen and hand off to the user. Do not click
   "Submit", "Send Application", "Finish", or any equivalent. This is the single most
   important rule in this workspace. If a form auto-submits on Enter, use `click` on a
   non-submit target instead of `press Enter`.
2. **NEVER fabricate.** Tailoring means re-ordering, re-weighting, and re-phrasing facts that
   already exist in `jobs/input/profile/master-resume.md`. It never means inventing employers, dates,
   titles, degrees, certifications, clearances, metrics, or years of experience. If the job
   asks for something the user does not have, leave it out and flag the gap in the
   application log — do not paper over it.
3. **NEVER guess a screening answer.** Work authorization, sponsorship, salary, notice
   period, veteran/disability/EEO, criminal history, relocation, references — answer only
   from `jobs/input/config/application-answers.md`. If the answer is not there, stop and ask the user,
   then persist the new answer back into that file.
4. **NEVER create accounts, accept terms, or pay** on the user's behalf without explicit
   per-instance approval. Many boards require an account to apply; ask first.
5. **`/job apply --batch` runs headed and leaves the browser open.** A human clicks Submit,
   so the run must end with a real window showing the filled form. Every `$B` call in that
   mode carries `--headed`, and **never `$B disconnect` at the end of a batch** — that throws
   away every filled form.
6. **Hand off, don't fight.** CAPTCHA, MFA, SSO, bot walls, or 3 failed attempts at the same
   interaction → `$B handoff "<reason>"`, tell the user what to do, wait, then `$B resume`.
7. **Never send `jobs/input/profile/` contents to a third-party service** (no pasting a resume into a
   random "AI resume scorer", no uploading to unrelated APIs). It goes to the employer's own
   application form and nowhere else.
8. **Never apply twice.** Check `jobs/output/tracker.md` for the job ID before starting.
9. **Respect the site.** Use the official/API path when one exists (see
   `jobs/input/config/job-sites.md`), keep request volume human-scale, and stop if a site explicitly
   blocks automated access rather than working around the block.
10. **NEVER invent a job description.** Tailoring and form-filling read the JD from
   `jobs/output/jds/<job-id>.md`. If that file is missing, or is still a
   `manual-required` stub, **stop and ask the user for the manual download** — do not
   reconstruct a JD from the job title, the company's marketing site, or a similar posting
   elsewhere. A resume aimed at a guessed JD is worse than no resume.

---

## Directory map

Two functional areas, each split into what the **user provides** (`input/`) and what the
**agent produces** (`output/`). Skills read from `input/`, write to `output/`, and never the
reverse.

### `jobs/` — tasks, plus profile setup → search → tailor → apply

`jobs/` has three parts, not two: `tasks/` is what the agent *does*, `input/` is what the
user provides, `output/` is what the agent produces.

| Path | What it holds |
|---|---|
| `jobs/tasks/*.md` | **The nine task files** `/job` routes to: `auto`, `setup`, `search`, `triage`, `jd`, `tailor`, `apply`, `status`, `interviews`. The agent's procedures, kept in the workspace so they can be edited without touching the skill. Each opens with its own tool budget. |
| `jobs/input/profile/source-resumes/` | Original resume files the user dropped in (PDF/DOCX/MD). **The only source of identity.** |
| `jobs/input/profile/master-resume.md` | Single source of truth for the user's experience **and identity** (name, email, phone in its Contact section). Gitignored; seeded from `jobs/input/templates/master-resume.md` by `/job setup`. |
| `jobs/input/profile/links.md` | LinkedIn, GitHub, portfolio, references. Gitignored. |
| `jobs/input/config/job-sites.md` | Sites to search, how to reach each one, per-site notes. **User-editable.** |
| `jobs/input/config/search-profile.md` | **The role families** the search targets, plus keywords, locations, comp floor, hard filters. User-defined, any number of families, no domain baked in. Seeded from `jobs/input/templates/search-profile.md` by `/job setup`. |
| `jobs/input/config/application-answers.md` | The screening-question answer bank. Grows over time. Gitignored. |
| `jobs/input/templates/` | The file formats the skills write, plus blank copies of the personal/config files (`master-resume.md`, `links.md`, `application-answers.md`, `search-profile.md`). Read the template before writing. |
| `jobs/output/jobs.md` | **The living index.** Summary + Unverified backlog + Excluded + Details. `/job search` merges into it; every later skill updates its **Status** and **JD / Resume / Form** columns in place. Never rewritten from scratch. |
| `jobs/output/tracker.md` | Master status board across all applications. |
| `jobs/output/run-log.md` | One entry per `/job` run, newest first — what was searched, filters in force, sites unswept, what the run learned. Narrative detail lives here, not in the `jobs.md` header. |
| `jobs/output/history/jobs-<timestamp>.md` | Timestamped snapshots of `jobs.md`, written by `jobs/bin/snapshot.sh` before any run edits the index. Last 50 kept. The dashboard's version picker renders any of them. |
| `jobs/bin/snapshot.sh` | Archives `jobs.md`. Run once per run, before the first edit. |
| `jobs/bin/fill-form.py` | Replays a job's `form-fill.json` into the **visible** browser — text, comboboxes, checkboxes, and a real `resume.pdf` upload. The mechanical half of `/job apply --batch`, and what `jobs/bin/morning-run.sh` calls for every mapped job each morning. Clicks no submit control. |
| `jobs/bin/morning-run.sh` | `make morning` — re-opens every mapped, unsubmitted application filled in its own tab, re-using a tab already on that posting. No agent, no tokens. |
| `jobs/bin/reset.sh` | `make reset` — puts the workspace back to a fresh start: deletes everything under `jobs/output/` and `interviews/output/`, restores `tracker.md` from `jobs/input/templates/tracker.md`, and leaves `jobs/input/` alone. `--profile` also wipes the resume, links and answer bank (then `/job setup` is required again). Archives whatever it removes to `.resets/` first, and refuses to run unattended without `--yes`. **Never run it on the user's behalf without them asking for it by name.** |
| `jobs/bin/daily-run.sh` | `make daily` — the whole loop in one command: `/job` (search → JD → tailor), then `/job apply --batch` (map + fill), then `morning-run.sh` (open every mapped application). Locked so a manual run and the scheduled one can't both edit `jobs/output/jobs.md`. Stages 1–2 are agent runs; stage 3 is free. |
| `jobs/output/logs/daily-<date>.log` | One log per daily run, last 30 kept. |
| `jobs/bin/mark-submitted.sh` | `make submitted JOB=<id>` — the user's own confirmation that they clicked Submit. The **only** thing that may set Status `submitted`, and what stops `make morning` re-opening a sent application every day. |
| `jobs/bin/refill-engine.js` · `jobs/bin/make-refill.py` | The replay half of `/job apply --batch`: the generator turns a job's `form-fill.json` into a `refill.js` the user pastes into the application page's console to refill the form later. The engine never clicks a button, so it cannot submit. |
| `jobs/output/jds/<job-id>.md` | **The full job description**, one file per job, written by `/job jd` — automatic and manual capture both land here. Everything downstream reads the JD from here, never from a live page. |
| `jobs/output/jds/<job-id>-source.*` | Raw posting the **user** downloaded by hand when auto-capture failed (`.pdf`, `.html`, `.txt`, `.md`, `.docx`, screenshot). The one user-supplied file under `output/`, because it belongs next to the JD it becomes. `/job jd` picks it up on the next run. |
| `jobs/output/applications/<job-id>/` | Per-job: tailored resume (`resume.md`, `resume.pdf`), cover letter, log, screenshots — plus, after `/job apply --batch`, the field map `form-fill.json` and the replay script `refill.js` generated from it by `jobs/bin/make-refill.py`. |
| `jobs/dashboard/` | Go dashboard for `jobs/output/jobs.md` — `make web` → http://localhost:8080. Also the resume PDF renderer (`make pdf`) and the document viewer behind the table's Status / Resume links. |

### `interviews/` — interview-experience intel (separate `ix` namespace)

| Path | What it holds |
|---|---|
| `interviews/input/config/ix-profile.md` | Run limits, credibility scoring, junk filter. |
| `interviews/input/config/ix-sources.md` | Subreddits, HN, and sites to search. |
| `interviews/input/templates/ix-results.md` | Section shapes for the results file. |
| `interviews/output/interviews.md` | The compiled experiences index. Reads `jobs/output/jobs.md` for company names, never writes to it. |
| `interviews/output/reports/<company-slug>/` | Per-company prep briefs. |
| `interviews/bin/ix-fetch.py` | Reddit / HN fetcher the skill shells out to. |

### Root

| Path | What it holds |
|---|---|
| `Makefile` | Dashboard tasks: `web`, `stats`, `ix-stats`, `build`, `check`, `clean`. `make` lists them. |
| `.claude/skills/job/SKILL.md` | **The one skill.** Router only — routing table, task isolation rules, sequencing gates, pipeline. Holds no procedure. |

## Task routing

**There is exactly one skill: `/job`.** It is a router and nothing else — it reads the
request, picks a task, loads `jobs/tasks/<task>.md`, and follows it.

The procedures live in `jobs/tasks/`, not in the skill, so they are workspace content the
user can edit directly; a change there takes effect on the next run. The router carries only
a one-line summary of each task, which is never enough to run it — **always read the file**.

| The user wants to… | Run |
|---|---|
| Just get on with it | **`/job`** with no argument — the auto pipeline |
| Anything, without remembering the task name | `/job <whatever they said>` — the router picks |
| See where every job stands, starting nothing | `/job status` |
| Set up / refresh their resume + profile | `/job setup` |
| Find jobs | `/job search` |
| Rank, dedupe, and shortlist what was found | `/job triage` |
| Check which postings are still live and grab the full JDs | `/job jd` |
| Tailor the resume (and cover letter) to one job | `/job tailor <job-id>` |
| Tailor resumes for every job that has a JD, in parallel | `/job tailor --all` |
| Actually fill out an application | `/job apply <job-id>` |
| Find real interview experiences (at a target company, or with an AI screener) | `/job interviews` |
| See status, log outcomes, chase follow-ups | `/job status` |

### The auto pipeline — `/job`

Bare `/job` runs `jobs/tasks/auto.md`: read the resume → infer 3–4 role families → search →
collect JDs → tailor. It stops at `tailored` and **can never fill or submit a form**.

**A missing resume degrades the run, it does not block it.** The resume feeds role-family
inference, the 25-point skill-overlap dimension of the fit score, and tailoring — nothing
else. So `/job java full stack developer` with an empty profile builds a role family out of
that phrase, searches, verifies, captures JDs, marks every fit score provisional, and stops
before tailoring with a note saying what it needs. Only bare `/job` with neither a resume nor
a focus stops early, because there is nothing to search for.

It is a **reconciler, not a sequence.** Each run looks at what state every job is in and does
the next thing possible for it. Three consequences that matter:

- **A stuck job never stops the batch.** JD capture fails → the row gets `⏳ manual`, the run
  moves to the next job, and every failure is reported once at the end with its URL.
- **The next run finishes what the last one couldn't.** Paste the JD into
  `jobs/output/jds/<job-id>.md` (or drop the posting in as `<job-id>-source.pdf`) and the
  next `/job` sees a JD where there wasn't one and carries that job through to a tailored
  resume. There is no separate resume mode — the loop is idempotent.
- **Finished work is never re-done.** `/job` on a settled workspace does almost nothing.

Unattended, it stands in for the triage checkpoint with the fit bands from
`search-profile.md`: 60+ auto-shortlists and carries through to `tailored`; 40–59 is left at
`found` for the user to promote; hard-filter failures are `skipped` as usual. Inferred role
families are marked provisional in `search-profile.md` and called out in the report — an
unreviewed guess about what job someone wants is the most expensive thing here to get wrong.

`make web` surfaces the result: a **Pipeline** card (JD / resume / form counts) and a
**Needs you** panel listing every `⏳ manual` row with a link to its posting.

Normal end-to-end flow:

```
/job setup → /job search → /job triage → /job jd → /job tailor → /job apply → /job status
  (once)                                      ↑            ↑             ↑
                                       careers page,  one subagent   one job at a
                                       or the user    per job        time, never
                                       drops the JD                  parallel
```

`auto` is the only task that chains others, and the router enforces one rule on top of it:
**`auto` never reaches `apply`.** Applying is attended, always — one job, user watching, user
submits.

**The tasks stay separate on purpose.** `tailor` produces a resume and touches no browser;
`apply` fills a form and does no tailoring. Each refuses to start until the previous stage's
artifact exists on disk, so a failure at one stage never half-produces the next one and any
task can be re-run alone.

**Task isolation is now enforced by you, not by the harness.** These eight were once eight
skills, each with its own `allowed-tools`. As one skill they share a single tool list that is
the union of all eight, so nothing mechanically stops `triage` from opening a browser or
`apply` from spawning a subagent. Every file in `jobs/tasks/` therefore opens with a **tool budget** —
read it and treat it as the only tool list you have. Reaching outside it means you have
drifted into another task: stop, re-route, and do not carry the tool across.

Load **one task file at a time.** Reading all eight defeats the point of the split and floods
the context with instructions for work you are not doing.

**Skills read from `input/`, write to `output/`, and read `jobs/tasks/` for their own
procedure.** A task file is never rewritten by the agent mid-run — if a task's steps are
wrong, say so and let the user edit it.

## Conventions

**Two URLs per job.** `Apply URL` is the exact application form (an ATS deep link, which
churns when a req is reposted). `Careers` is the employer's own careers page, falling back to
the ATS board root — the durable link that survives a closed req. Prefer the employer's own
apply path when both exist.

The `Careers` URL is also the **liveness test**: an employer's own board lists open reqs
only, so a role present there is still active and a role absent from a board that rendered
other jobs is dead. But an empty render is not evidence of absence — plenty of careers pages
lazy-load and show nothing headless. Reading a JD off the apply URL alone proves the page
exists, not that the req is open, and is recorded as `inconclusive` forever. Never round
that up.

**Countries** — `search-profile.md` → *Location & work mode* → `Countries:`, shipping as
`United States`. A **hard filter**, not a scoring dimension: a job outside it is excluded with
`reason: country`, never merely penalised, because a perfect-stack role in a country the user
can't work in should not outrank one they can take. Applied three times — in the query, at
collection, and again at verification, where the posting's own stated location beats the
aggregator's. `--country "<list>"` overrides for one run; `--country Anywhere` turns it off.

**Fit** stays in `jobs.md` but is **hidden from the table** — `/job auto` shortlists on its
bands, `/job triage` takes it as a floor, and the Dashboard tab charts it, so it is
load-bearing data rather than something to read row by row. There is no `#` column: sorting
replaced it, and it had to be renumbered on every merge.

**Stage** — the one-word position computed from Status + the JD/Resume/Form columns so it can
never disagree with them: `Searched` → `Needs your JD` → `JD captured` → `Tailored` →
`Filling form` → `Ready for submission` → `Submitted`. **`Ready for submission`** is the
agent's ceiling — `/job apply` filled every field and stopped. It is **not a table column**:
Status plus JD / Resume / Form say the same thing in the same row, so showing it was a fourth
restatement of three columns. It still drives the Dashboard tab and the "Needs you" panel;
`showStage` in `jobs/dashboard/main.go` puts the column back.

**Profile** — the role family from `search-profile.md` that surfaced a job (`AI Engineer`,
`Java Full Stack Developer`, `Adjacent`). Written by `/job search` into its own column in
`jobs/output/jobs.md`, using the family's heading verbatim. One index can hold several
parallel job hunts; this column is what keeps them apart, on the dashboard and in triage.

**Job ID** — the stable key used for folder names, tracker rows, and dedupe:
`<company-slug>--<role-slug>` (lowercase, non-alphanumerics → `-`), e.g.
`anthropic--forward-deployed-engineer`. Collisions get a `--<location-slug>` suffix.

**Snapshots** — every run that edits `jobs/output/jobs.md` calls `jobs/bin/snapshot.sh`
first, once. The index is one file that all stages merge into, so this is the only undo.

**Dates** — always absolute ISO (`2026-08-30`), never "today" or "last week".

**Status** lives in the `jobs/output/jobs.md` **Status** column and is mirrored in the per-job log and
`jobs/output/tracker.md`. `jobs/output/jobs.md` is what the dashboard reads, so it must never go stale.

**Status values:** `found` → `shortlisted` → `jd-captured` → `tailored` →
`filled-awaiting-user` → `submitted` → `interviewing` → `offer` / `rejected` / `ghosted` /
`skipped`. `submitted` is set **by the user**, never by the agent — the agent can only reach
`filled-awaiting-user`.

**Stage columns.** Status is the linear position; three more columns in `jobs/output/jobs.md` say
which artifact actually exists, so the user can see the resume half and the form half
separately:

| Column | Written by | Values |
|---|---|---|
| **JD** | `/job jd` | `✓ <date>` · `⏳ manual` · `—` |
| **Resume** | `/job tailor` | `✓ <date>` · `—` (the ✓ links to the resume in the dashboard) |
| **Form** | `/job apply` | `✓ <date>` · `⏳ in-progress` · `—` |

`⏳ manual` means auto-capture failed and **the user must download the posting themselves**
into `jobs/output/jds/`. That job is blocked until they do — never work around it by
tailoring against the thin summary in `jobs/output/jobs.md`.

## Browser usage

All web work goes through the gstack `/browse` skill (`$B`). Never use
`mcp__claude-in-chrome__*`. Standard setup before any browse command:

```bash
_ROOT=$(git rev-parse --show-toplevel 2>/dev/null)
B=""
[ -n "$_ROOT" ] && [ -x "$_ROOT/.claude/skills/gstack/browse/dist/browse" ] && B="$_ROOT/.claude/skills/gstack/browse/dist/browse"
[ -z "$B" ] && B="$HOME/.claude/skills/gstack/browse/dist/browse"
[ -x "$B" ] && echo "READY: $B" || echo "NEEDS_SETUP"
```

Workhorse commands for this workspace: `goto`, `snapshot -i` (get `@e` refs), `fill`,
`select`, `click`, `upload`, `forms` (dump fields as JSON), `is visible`, `screenshot`,
`frame` (Greenhouse/Lever/Workday embed forms in iframes), `handoff` / `resume`.

## Subagents

Fan-out is allowed **only** for per-job resume tailoring (`/job tailor --all`), one
subagent per job, at most 4 in flight. Everything else runs in the main session.

Rules, because these agents write to a shared workspace:

- **A subagent writes only inside its own `jobs/output/applications/<job-id>/`.** Never `jobs/output/jobs.md`,
  never `jobs/output/tracker.md`, never `jobs/output/jds/`, never another job's folder.
- **The parent is the only writer of shared files**, after the agents report back. Several
  agents editing one Markdown table at once corrupts it, and `jobs/output/jobs.md` is what the
  dashboard reads.
- **Never parallelize anything that uses the browser.** `$B` is a single shared page — two
  agents driving it collide. JD capture over HTTP/ATS APIs can go in parallel; browser work,
  PDF rendering that may fall back to the browser, and all of `/job apply` are serial.
- **Never parallelize `/job apply`.** One form at a time, with the user in the loop.
- A subagent starts cold: give it the job ID, the JD path, the master-resume path, the
  template path, and the no-fabrication rule in its brief. It can see nothing of the parent
  conversation.
- **`jobs/input/profile/` is read-only to subagents.** They never send it anywhere (rule 6).
- Reject, don't publish, any subagent result that reports a rule violation (bullets added
  > 0) or comes back empty. Say so in the report and leave that job untailored.

## When to ask the user vs. decide

**Decide yourself:** search phrasing, which of the configured sites to hit, how to word a
bullet, ranking order, whether two postings are duplicates, which ladder rung captured a JD,
how many tailoring subagents to run.

**Ask the user:** any screening answer not in the answer bank, creating an account,
anything that costs money, a job that fails a hard filter but looks strong anyway, **a JD
that needs a manual download**, and always at the end of `/job apply` before they submit.

Batch the asks. A `/job jd` run that needs three manual downloads says so once at the
end with all three URLs — it does not stop three times.

## Output style

Terminal output stays short: what was searched, what was found, what needs a decision.
The detail lives in the Markdown files — link to them by path rather than reprinting them.
