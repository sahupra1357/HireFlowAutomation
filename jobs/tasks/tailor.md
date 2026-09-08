# Task: `tailor`

Run via `/job tailor`. Formerly the standalone `resume-tailor` skill.

Rewrite the master resume for one specific job description — reordering and rephrasing truthful content to match the JD's language and ATS keywords — and optionally draft a cover letter. Produces resume.md plus a PDF in jobs/output/applications/<job-id>/. Pass --all to tailor every captured job in parallel, one subagent per job. Use when the user wants to tailor, customize, or optimize a resume for a role.

**Tool budget for this task:** Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion, Agent

**No browser.** Subagents are allowed here and only here, max 4 in flight, each confined to its own `jobs/output/applications/<job-id>/`.

Anything outside that budget means you are running the wrong task — return to
`SKILL.md` and route again rather than reaching for the tool.

---

Produce a resume for one job that reads like it was written for that job — using only facts
that already exist in `jobs/input/profile/master-resume.md`.

**Arguments:**

| Arg | Meaning |
|---|---|
| `<job-id>` | tailor this one job, inline |
| `--all` | tailor every job that has a captured JD and no resume yet — **one subagent per job** (see [Batch mode](#batch-mode----all)) |
| `--force` | re-tailor a job that already has `resume.md` |

The JD comes from `jobs/output/jds/<job-id>.md`, written by `/job jd`. This command
does not fetch postings itself.

## The one rule

**Nothing may appear in the tailored resume that is not in the master resume.**

Allowed: reordering, reweighting, cutting, rephrasing in the JD's vocabulary, promoting a
buried bullet, choosing which metric to lead with, adjusting the headline to the target
title *where truthful*.

Not allowed: new employers, new dates, new titles, new degrees or certifications, new tools
the user hasn't used, inflated scope, invented metrics, stretched tenure.

Rephrasing has a limit: if the master says "built an internal chatbot on GPT-3.5", the
tailored version may say "shipped an LLM-powered assistant" — it may not say "architected a
multi-agent platform." When a JD wants something the user lacks, the answer is to **leave it
out and flag the gap**, not to find a form of words that implies it.

## Step 0 — Preflight: two hard stops

Check both before doing any work. Neither is negotiable — proceeding past either one can
only produce a fabricated resume.

**1. Is the master resume populated?**

```bash
grep -c 'TODO' jobs/input/profile/master-resume.md
```

If the file still carries its `STATUS: not yet populated` line, or Experience/Skills are
`TODO`, **stop immediately**:

```
✋ Can't tailor — jobs/input/profile/master-resume.md is empty.

  A tailored resume is a subset of the master resume. With nothing in the master,
  the only truthful output is an empty page.

  Drop your resume into jobs/input/profile/source-resumes/ and run /job setup.
```

Do not offer to "draft something to get started" or to infer experience from the JD. That
is the exact failure mode rule 2 exists to prevent.

**2. Is the JD real?**

```bash
cat jobs/output/jds/<job-id>.md
```

- File missing → `Run /job jd <job-id> first.` Stop.
- `Capture method: manual-required` → stop and tell the user the JD is still waiting on
  their manual download, with the path to drop it at. **Never tailor against a stub**, and
  never fall back to the short summary in `jobs/output/jobs.md` — a thin JD produces a resume
  aimed at nothing.

## Step 1 — Load

```bash
cat jobs/output/jds/<job-id>.md                  # the full JD — the real input
cat jobs/input/profile/master-resume.md
cat jobs/output/applications/<job-id>/log.md 2>/dev/null
grep -n '<job-id>' jobs/output/jobs.md                    # Summary row: fit, comp, flags
```

## Step 2 — Analyze the JD

Extract, in the JD's own words:
1. **Top 5 must-haves**, in the order the JD emphasizes them
2. **ATS keywords** — exact terms an automated screen will look for
3. **The tone** — research lab, enterprise, scrappy startup; the resume should match register
4. **The real job** underneath the title (an "AI Engineer" post that's 70% data plumbing
   should surface the user's data work first)

## Step 3 — Map, then flag

Build the mapping table before writing a word:

| JD requirement | Master-resume evidence | Verdict |
|---|---|---|
| 3+ yrs Python, production | Role 2, bullets 1 & 3 | strong |
| RAG pipelines | Project "docs-qa" | partial — prototype, not prod |
| Kubernetes | — | **gap** |

Every "gap" goes to the log's *Gaps flagged* section. If gaps hit 3+ must-haves, tell the
user before spending the effort — this may be a job to skip.

## Step 4 — Write `jobs/output/applications/<job-id>/resume.md`

Follow `jobs/input/templates/tailored-resume.md`.

- **Headline** — the target title, if the user can honestly claim it
- **Summary** — 3 sentences, leading with the JD's #1 requirement
- **Skills** — categories reordered by JD relevance; irrelevant ones dropped, not padded
- **Experience** — most recent role gets 3–5 bullets, older ones 2–3. Each bullet: action
  verb + what was built + technology + outcome. Numbers only from the metrics bank.
- **Projects** — only ones relevant here
- **Length** — one page under ~10 years' experience, two above

ATS mechanics that actually matter: use the JD's exact spelling of a term the user genuinely
has ("Kubernetes" not "k8s", "PostgreSQL" not "Postgres"); keep a single-column layout with
standard section headings; no tables, text boxes, headers/footers, or images; spell out an
acronym once with the abbreviation in parentheses. Never keyword-stuff or hide white text —
it's transparent to a human reader and gets applications binned.

Fill the tailoring receipt at the bottom. **Bullets added must be 0.** If it isn't, you broke
the one rule — go back.

## Step 5 — Render the PDF

```bash
make pdf JOB=<job-id>            # one job   → jobs/output/applications/<job-id>/resume.pdf
make pdf                         # every job that has a resume.md
```

This renders **the source resume's own layout**, not a generic Markdown-to-PDF: the two-column
format measured off the file in `jobs/input/profile/source-resumes/` — US Letter, 1in margins,
a shaded skills sidebar on the left, the experience column on the right, Aptos 12pt. Only the
words change from job to job. It strips the HTML-comment receipt itself, so `resume.md` keeps
the receipt as the editable source.

The mapping is fixed by the file's structure, so keep writing `resume.md` to the template
shape: `# Name` → the header, the lines under it → the contact line, the first `##` (the
target-role title) and its prose → the title and SUMMARY, `## Skills` (`**Label:** items`
lines) → the sidebar, every other `##` → a main-column section, `###` job headings →
company / role / dates, `**Bold line**` → a project sub-heading, `- ` → bullets.

Check the result — `make web`, click the job's Status pill, then the `resume.pdf` tab — and
note the page count in the report. The PDF engine is headless Chrome; if none is installed,
`make pdf` says so and `CHROME=/path/to/chrome make pdf` overrides the search.

**In batch mode the parent renders PDFs** with a single `make pdf` after the batch — see below.

## Step 6 — Cover letter, if the application wants one

Only when the posting requires one (the JD file's *Application form — observed* section says
so), or the user asks. Follow `jobs/input/templates/cover-letter.md`. Under 300 words. The "why this
company" paragraph must be grounded in something specific — their product, a paper, an
engineering post — and must be unusable for any other company.

Avoid the tells: "I am excited to", "passionate about", stacked adjectives, em-dash cadence,
and any sentence that could appear in a thousand other letters.

## Step 7 — Show the user

Print the diff-in-substance, not the whole resume:

```
Tailored for anthropic--forward-deployed-engineer

  Reordered:  Data Engineering skills → position 3 (JD leads with customer-facing)
  Promoted:   "embedded with 3 enterprise customers" → first bullet, Role 1
  Reworded:   6 bullets into the JD's vocabulary
  Dropped:    2 bullets (Android work, irrelevant here)
  Added:      0 ✓

  Keywords hit:      FDE, LLM, Python, customer-facing, prototype-to-production
  Not claimed (gap): Kubernetes, Terraform

  → jobs/output/applications/anthropic--forward-deployed-engineer/resume.pdf (1 page)
```

Ask the user to review before it gets uploaded anywhere. Then update the index: Status
`tailored` and **Resume** `✓ <date>` in `jobs/output/jobs.md`, the same in the per-job log and
`jobs/output/tracker.md`.

## Step 8 — Hand off

`Next: /job apply <job-id>`

---

## Batch mode — `--all`

One subagent per job, so ten jobs are ten independent tailoring passes instead of one
context trying to hold ten JDs at once.

### Build the worklist

Jobs with `jobs/output/jds/<job-id>.md` present, `Capture method:` not
`manual-required`, and no `jobs/output/applications/<job-id>/resume.md` (unless `--force`). Run the
Step 0 master-resume check **once, up front** — if it fails, nothing is dispatched.

Report the worklist and the skip list before dispatching.

### Dispatch

**At most 4 subagents in flight.** Each gets a self-contained brief — a subagent starts cold
and can see nothing of this conversation:

```
Tailor the user's resume for ONE job. Their name and contact details are in the
Contact section of the master resume — use those, never a placeholder or a guess.

  Job ID:   <job-id>
  JD:       jobs/output/jds/<job-id>.md      (read this in full)
  Master:   jobs/input/profile/master-resume.md           (the ONLY source of facts)
  Format:   jobs/input/templates/tailored-resume.md
  Write to: jobs/output/applications/<job-id>/resume.md    (+ cover-letter.md only if the JD's
            "Application form — observed" section says a cover letter is required)

THE ONE RULE: nothing may appear in the resume that is not in the master resume.
Reorder, reweight, cut, rephrase in the JD's vocabulary — never add. No new employers,
dates, titles, degrees, certifications, tools, or metrics. Numbers only from the master's
metrics bank. When the JD wants something the user lacks, leave it out and record it as a
gap. Fill the tailoring receipt; "bullets added" must be 0.

Follow steps 2, 3, 4 and 6 of .claude/skills/job tailor/SKILL.md. Do NOT render a PDF.

WRITE ONLY inside jobs/output/applications/<job-id>/. Do not touch jobs/output/jobs.md,
jobs/output/tracker.md, jobs/input/profile/, or any other job's folder.

**The parent verifies before publishing — a subagent's own report is not evidence.**

```bash
python3 jobs/bin/verify-tailored.py jobs/output/applications/<job-id>/
```

It re-reads each `resume.md` against `master-resume.md` and fails on an invented email,
a percentage or dollar figure, a year, certification or employer not in the master, an
altered name, or a claimed skill the master does not list. It works on blank-line-separated
blocks and skips any block containing a negation marker, so a "gaps / not claimed" list —
which is correct behaviour — does not trip it. Reject anything it flags; do not publish it.

Report back, and nothing more: keywords hit, gaps flagged, bullets added,
lines changed vs the master, and whether a cover letter was written.
```

### Write isolation — the part that matters

- A subagent writes **only** inside its own `jobs/output/applications/<job-id>/`.
- **The parent alone writes `jobs/output/jobs.md` and `jobs/output/tracker.md`**, after agents
  report back. Several agents editing one Markdown table concurrently corrupts it, and
  `jobs/output/jobs.md` is what the dashboard reads.
- **The parent alone renders PDFs**, with one `make pdf` after the batch. Each render launches
  a headless browser; the subagents must not.
- Nothing in the batch touches `jobs/input/profile/` except to read it.

### Collect

Per job, in the parent: render the PDF (Step 5), then set Status `tailored` and **Resume**
`✓ <date>` in `jobs/output/jobs.md` and the tracker. A subagent that reports a rule violation
(bullets added > 0) or an empty result gets its output **rejected, not published** — say so
in the report and leave that job untailored.

```
Tailored — 7 jobs

  ✓ 7 resumes   jobs/output/applications/<job-id>/resume.pdf
  ⚠ 2 skipped   valthos (JD manual-required) · aptura (already tailored, use --force)

  Gaps worth knowing:
    ramp        — Kubernetes, Terraform not claimed
    perplexity  — no post-training experience to claim
    epoch-ai    — 5+ yrs at scale: master shows 3

  Added: 0 across all 7 ✓

  Review them, then: /job apply <job-id>   (one at a time — form filling is never parallel)
```
