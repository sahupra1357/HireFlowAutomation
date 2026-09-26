# Task: `add`

Run via `/job add <url> [<url> ...]`, `/job add --inbox`, `/job add` followed by pasted JD
text, or just `/job <url>`.

Put a job the user found themselves on the list — a link someone sent, a posting on a site
the search doesn't cover, a company they want. Verifies the posting, mints the job ID, writes
the row as `shortlisted`, and hands the new IDs to `jd` for capture. Use whenever the user
pastes a job URL or job-description text, or says "add this job".

**Tool budget for this task:** Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion

Drives `$B` read-only — `goto`, `text`, `snapshot` — and only when the ATS API can't serve
the posting. **Never fills or submits a form.** No subagents, no web search: the user gave
the URL, there is nothing to search for.

Anything outside that budget means you are running the wrong task — return to
`SKILL.md` and route again rather than reaching for the tool.

---

`search` finds jobs; this takes the ones the user already found. The difference that shapes
every step: **the user chose this job.** So it skips triage (Status starts at
`shortlisted`), a hard-filter failure is a question for the user rather than a silent
exclusion, and a thin fit score is information, not a reason to drop it.

Everything else is the same discipline as `search`: one living index, dedupe by job ID,
verify before listing, never invent a JD.

**Arguments:**

| Form | What it adds |
|---|---|
| `/job add <url> [<url> ...]` | those postings |
| `/job <url>` | same — the router sends a bare URL here |
| `/job add --inbox` | every link under **Add** in `jobs/input/inbox.md` not already in the index |
| `/job add` + pasted text | one job whose JD is the pasted text (ask for company and role if the text doesn't say) |
| `--no-jd` | stop after writing the rows; don't chain into `jd` |

---

## Step 0 — Setup

```bash
jobs/bin/snapshot.sh                     # once, before the first edit to jobs.md
bash jobs/bin/ensure-tracker.sh
test -f jobs/input/inbox.md || cp jobs/input/templates/inbox.md jobs/input/inbox.md
cat jobs/input/templates/job-results.md  # Summary row + Details block shapes
grep -oE '[a-z0-9-]+--[a-z0-9-]+' jobs/output/jobs.md jobs/output/tracker.md 2>/dev/null | sort -u
```

If `jobs/output/jobs.md` does not exist yet, create it with the template's header block and
its empty **Summary**, **Unverified backlog** and **Excluded by verification** tables, then
continue — `add` can be the very first thing a user ever runs.

**Build the worklist.**

- URLs from the arguments, or with `--inbox`, every `- <url>` line under `## Add` in
  `jobs/input/inbox.md`. Skip lines containing `<` (the template's examples). The optional
  `| company | role | note` after a URL is a hint, used when the page can't be read.
- **Never edit `jobs/input/inbox.md`.** It is the user's file; the index is the record of
  what was processed.
- **Drop URLs already in the index.** Grep `jobs/output/jobs.md` for the URL (and for it with
  query string stripped). A URL found in the Summary or backlog is reported as
  `already on the list as <job-id>`; one found in Excluded is reported with its reason.

## Step 1 — Read the posting

Per URL, cheapest reliable path first — exactly `jd.md`'s rungs 1–3, so a posting read here
reads the same as one captured there:

1. **ATS API** when the URL is Greenhouse, Lever or Ashby (including an employer page that
   embeds one — `?gh_jid=`, `jobs.lever.co`, `jobs.ashbyhq.com` in the page). The board-level
   listing is also the liveness proof.
2. **Careers page / the URL itself in the browser**, read-only: `$B goto`, `$B text`.
3. **Can't read it** — LinkedIn and Indeed usually need a login; some boards block
   automation. Do not work around a block. Use the inbox hint (company, role) if present;
   otherwise record the URL as unreadable and move on.

From what was read, extract: company, role title, location / work mode, comp as posted,
ATS, the **Apply URL** (the ATS form, not the aggregator), and the **Careers** URL (the
employer's own careers page, falling back to the ATS board root).

**Pasted text** skips this step: the text *is* the JD. Take company and role from it or from
the user; Apply and Careers URLs are `—` unless the user gives them.

## Step 2 — Mint the job ID and dedupe

`<company-slug>--<role-slug>` (lowercase, non-alphanumerics → `-`), with `--<location-slug>`
on a collision. If the ID already exists in any of the three tables or the tracker, it is
the same job reached by another link: report it and add nothing.

A job currently in **Unverified backlog** is **promoted**: move its row into Summary as
`shortlisted`, carrying what the backlog already knew.

## Step 3 — Verify, lightly

The user vouched for the job's appeal, not for the posting being real or open.

- **Liveness** — `verified-live` only if the role is on the employer's own board or the ATS
  live index; a page that merely rendered is `likely-live`; pasted text or an unreadable
  link is `unverified`. Same rule as everywhere: never round up.
- **Scam red flags** from `search-profile.md` → *Scam hard-drops* — a match is **not added**;
  tell the user why.
- **Hard filters** (country, and the rest of *Hard filters* in `search-profile.md`) — a
  failure is **not** silently excluded, because the user picked this job. Collect it and ask
  once, at the end (Step 6).

## Step 4 — Score

Score fit 0–100 with the weights in `search-profile.md`, against the JD text read in Step 1,
the same way `search` Step 5 does, with the same liveness penalties. No resume on file →
mark the score provisional. **The score does not gate anything here** — it is recorded so
the dashboard and `/job status` can compare this job with the rest.

**Profile** — the active family (per `SKILL.md`: `setup-families.md`, else
`search-profile.md`) whose titles and signals the job matches best, heading verbatim; none
matches → `Adjacent`.

## Step 5 — Write the index (one writer, after each job)

Following `jobs/input/templates/job-results.md` exactly:

- **Summary row** — all 15 cells. `Source: added`, `JD —`, `Resume —`, `Form —`,
  **`Status: shortlisted`**. Place it by fit so the table stays sorted.
- **Details block** — the template's block, with `Source: added by you (<date>)`,
  `Posting URL:` the link the user gave, the inbox note if any under *Application notes*,
  and honest *Why it fits* / *Gaps* lines from `master-resume.md` (or "no resume on file").
- **`jobs/output/tracker.md`** — an Active row with the same state.
- **Pasted text** — also write `jobs/output/jds/<job-id>.md` now, from
  `jobs/input/templates/job-description.md`, with the user's text verbatim under
  *Full job description*, `Capture method: manual`, `Liveness: inconclusive`; set **JD** to
  `✓ <date>` and Status to `jd-captured`. Nothing to chain for this one.

## Step 6 — Chain into `jd`, then report

Unless `--no-jd`: read `jobs/tasks/jd.md` and run it on exactly the job IDs this run added
(`/job jd <id>` each). `add` is allowed to chain `jd` and nothing else — it never evaluates,
tailors, or applies. The JD text from Step 1 is already in hand, so for an ATS-API job this
is a write, not a second fetch.

Then one report, with every question batched at the end:

```
Added 3 of 5 — all shortlisted

  ✓ acme--platform-engineer           fit 82  verified-live   JD ✓
  ✓ globex--data-engineer             fit 64  likely-live     JD ✓
  ✓ initech--staff-analyst            fit 71  unverified      JD ⏳ manual
  = hooli--ml-engineer                already on the list (found 2026-09-20)
  ✗ https://www.linkedin.com/jobs/view/…   needs a login — add `| Company | Role` to the
                                            inbox line, or paste the JD: /job add <text>

  ? umbrella--site-reliability-engineer   outside Countries (Germany) — add it anyway?

  Next: /job evaluate <id> for a written assessment · /job tailor <id>
        (a bare /job carries shortlisted jobs through on its next run)
```

For a hard-filter question the user answers yes to, add the row as above and note
`hard-filter override: <filter>` in its Details block. A no adds it to **Excluded** with
`reason: <filter> (declined by user)` so it is never asked again.

## Rules

- **Never invent a JD.** An unreadable posting gets a row (if the user gave company and
  role) and `⏳ manual` from `jd`, never a JD reconstructed from the title.
- **Never edit `jobs/input/inbox.md`.** Read-only; the index records what was done.
- **Never re-add.** A job ID or URL already in the index is reported, not duplicated.
- **Respect the site.** A login wall or bot block is a reason to ask the user, not to work
  around it.
