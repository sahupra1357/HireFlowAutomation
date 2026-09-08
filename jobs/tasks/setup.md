# Task: `setup`

Run via `/job setup`. Formerly the standalone `profile-setup` skill.

One-time (or refresh) setup of the job seeker's profile — ingest an existing resume PDF/DOCX into jobs/input/profile/master-resume.md, pull the LinkedIn profile, and fill the screening-question answer bank. Run this before the first job search, or whenever the resume changes.

**Tool budget for this task:** Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion, WebFetch

May drive `$B` only to fetch the user's own LinkedIn profile. Never spawns subagents.

Anything outside that budget means you are running the wrong task — return to
`SKILL.md` and route again rather than reaching for the tool.

---

Turn the user's existing materials into the structured profile the rest of this workspace
reads. Everything downstream — search fit scores, resume tailoring, form filling — is only
as good as what this step produces, so be thorough and do not guess.

## Step 1 — See what's already there

The three personal files are gitignored, so a fresh clone has none of them. Seed any that
are missing from the committed templates before anything else:

```bash
[ -f jobs/input/profile/master-resume.md ]        || cp jobs/input/templates/master-resume.md        jobs/input/profile/master-resume.md
[ -f jobs/input/profile/links.md ]                || cp jobs/input/templates/links.md                jobs/input/profile/links.md
[ -f jobs/input/config/application-answers.md ]   || cp jobs/input/templates/application-answers.md jobs/input/config/application-answers.md
[ -f jobs/input/config/search-profile.md ]        || cp jobs/input/templates/search-profile.md      jobs/input/config/search-profile.md
ls -la jobs/input/profile/source-resumes/ 2>/dev/null
grep -c 'TODO' jobs/input/profile/master-resume.md jobs/input/profile/links.md jobs/input/config/application-answers.md 2>/dev/null
```

If `source-resumes/` is empty, ask the user to drop their resume in
(`jobs/input/profile/source-resumes/`) and stop. Nothing useful happens without it.

**Nothing about the user is hardcoded anywhere in this repo.** Name, email, phone,
location, links — every identity fact comes from the file they dropped in. If you find
yourself typing a name that didn't come from `source-resumes/`, stop.

If `master-resume.md` already has few TODOs, this is a **refresh**: ask what changed rather
than rebuilding from scratch.

## Step 2 — Extract the resume

Read whatever is in `jobs/input/profile/source-resumes/`:

- **PDF** → `Read` handles it directly (use the `pages` param for long files).
- **DOCX** → `textutil -convert txt -stdout file.docx` on macOS, or
  `unzip -p file.docx word/document.xml | sed -e 's/<[^>]*>//g'` as a fallback.
- **Markdown / txt** → read it.

Extract *everything*, including things that seem irrelevant. `master-resume.md` is a
superset; tailoring cuts down from it later. Losing a fact here means it can never appear
in any application.

## Step 3 — Write `jobs/input/profile/master-resume.md`

Fill every section of the existing template. Rules:

- **Contact first.** Replace `{{Full name}}` in the title and the `TODO`s under *Contact*
  with the name, email, phone, and location exactly as they appear on the resume. This is
  the one place the user's name is written down; every tailored resume, cover letter, PDF
  filename, and form field reads it from here.
- **Copy facts exactly.** Dates, titles, company names, degrees, and numbers transfer
  verbatim. Do not round, upgrade, or "clean up" a title.
- **Populate the metrics bank.** Pull every number in the resume into that table with its
  context. This is what tailored bullets are allowed to draw from.
- **Fill "Things I have NOT done"** by asking the user about common JD requirements their
  resume doesn't evidence (Kubernetes? on-call? managing people? a specific cloud?).
- Leave `TODO` anywhere the source doesn't say. Never fill a gap with a plausible guess.

Then show the user a short diff-style summary: roles found, years covered, skill categories,
count of remaining TODOs.

## Step 4 — Links and LinkedIn

Ask for LinkedIn / GitHub / portfolio URLs if `jobs/input/profile/links.md` still has TODOs.

With a LinkedIn URL, fetch the public profile to cross-check:

```bash
_ROOT=$(git rev-parse --show-toplevel 2>/dev/null)
B=""; [ -n "$_ROOT" ] && [ -x "$_ROOT/.claude/skills/gstack/browse/dist/browse" ] && B="$_ROOT/.claude/skills/gstack/browse/dist/browse"
[ -z "$B" ] && B="$HOME/.claude/skills/gstack/browse/dist/browse"
$B goto "<linkedin-url>"
$B text
```

LinkedIn often shows a login wall to headless. If so, `$B handoff "LinkedIn wants a login to
show your profile"`, let the user log in, then `$B resume`. If they'd rather skip it, skip
it — it's a nice-to-have.

Mirror the highlights into `jobs/input/profile/links.md`. **If LinkedIn and the resume disagree** on a
date, title, or employer, surface the conflict to the user and let them pick. Never silently
reconcile — a mismatch between a submitted resume and a public profile is a real problem for
the user.

## Step 5 — Fill the answer bank

First copy the *Identity & contact* rows (full name, email, phone, address if the resume
shows one) from what Step 2 extracted — do not ask the user for things their resume already
says. Then walk `jobs/input/config/application-answers.md` for the remaining `TODO`s. Batch them with
`AskUserQuestion` — a few at a time, most-blocking first:

1. **Work authorization + sponsorship** (blocks the most applications)
2. Phone + address (nearly every form requires them)
3. Relocation, travel willingness, work mode, notice period
4. Salary expectation
5. Years-of-experience numbers
6. EEO self-identification — offer "decline to self-identify" as the first option and make
   clear these are always optional

For anything the user declines to answer, write `ASK-EACH-TIME` rather than `TODO`, so
`/job apply` knows to prompt instead of treating it as unfinished setup.

## Step 6 — Derive the search profile from the resume

`jobs/input/config/search-profile.md` is **not preloaded with a field**. Its role families are
whatever this user is actually looking for, and this step is where they get written. Do not
assume the workspace is for engineering, or for any other domain — read what Step 2 extracted.

1. **Propose role families from the resume.** From the titles, tools, and responsibilities in
   `master-resume.md`, draft 2–5 families. Each needs:
   - **Titles** — the exact title plus every synonym and seniority variant the market uses
     for it. Include the ones the user has *not* held but is qualified for; a family is a
     search target, not a work history.
   - **Signals** — phrases that would appear in the JD body and confirm the role is really
     that work. These are what disambiguate a title that means different things in different
     industries.
   Always include a final **Adjacent** open bucket so an emerging title isn't dropped.

2. **Show the draft and let the user edit it** with `AskUserQuestion` — which families to
   keep, which to drop, what's missing. They know their target market better than the resume
   does; someone pivoting will want families their history doesn't evidence yet. Take what
   they say over what you inferred.

3. **Fill the keyword lists** from the same source: *boost* keywords are the skills and tools
   the resume actually evidences; *exclude* keywords are the disqualifiers for this user
   specifically (a credential they lack, a clearance they don't hold, a years-of-experience
   bar beyond them).

4. **Define employer quality.** Ask what a good employer looks like in their field and write
   it into the one-line definition under the fit-scoring table. Left blank, the agent guesses.

5. **Fill the remaining `{{...}}` markers** with real numbers: metro area, comp floor,
   seniority band, time zone.

Leave no `{{placeholder}}` behind — every one is a dimension the search will otherwise skip.

## Step 7 — Report

```
Profile ready.
  Resume:     N roles, YYYY–YYYY, N skill categories
  Links:      LinkedIn ✓  GitHub ✓  Portfolio —
  Answers:    N of M filled, N marked ask-each-time
  Remaining:  <list the TODOs that still matter>

Next: /job search
```

If meaningful TODOs remain, say which ones and what they'll block. Don't declare the profile
ready when work authorization is still blank.
