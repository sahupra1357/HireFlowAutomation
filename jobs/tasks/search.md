# Task: `search`

Run via `/job search`. Formerly the standalone `job-search` skill.

Search the configured job sites for the role families defined in jobs/input/config/search-profile.md, then merge the verified results into the living jobs/output/jobs.md index. Use when the user wants to find jobs, run a search, or check for new postings.

**Tool budget for this task:** Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion, WebSearch, WebFetch

Writes only `jobs/output/jobs.md`. Never opens an application form, never tailors.

Anything outside that budget means you are running the wrong task — return to
`SKILL.md` and route again rather than reaching for the tool.

---

Search every enabled site in `jobs/input/config/job-sites.md` for roles matching
`jobs/input/config/search-profile.md`, and produce one reviewable Markdown file the user can walk
through later.

**Optional arguments:**
- A role family or free-text focus — `/job search <family name>`, `/job search remote <role>`.
  With no focus, search **every** role family listed in `search-profile.md`, however many
  there are. Never assume a fixed set or a fixed count; read the file.
- `--country "<list>"` — override the country filter for this run only. `--country Anywhere`
  disables it. Persistent default is **Countries** under *Location & work mode* in
  `search-profile.md`, which ships as `United States`.
- `--limit N` — override the result limit for this run only. Persistent default lives in
  `jobs/input/config/search-profile.md` under **Result limit**.
- `--limit all` (or `--no-limit`) — **exhaustive mode**: sweep every enabled site and verify
  everything found, bounded by the time budget instead of a count.

## Step 0 — Load context

```bash
cat jobs/input/config/search-profile.md
grep -A6 'Enabled: yes' jobs/input/config/job-sites.md | head -100
grep -c '^|' jobs/output/jobs.md 2>/dev/null
grep -oE '[a-z0-9-]+--[a-z0-9-]+' jobs/output/jobs.md jobs/output/tracker.md 2>/dev/null | sort -u | wc -l
```

Read the search profile fully, note which sites are enabled, and load every job ID already
seen — those are your dedupe set. `jobs/output/jobs.md` is the single living index: its **Summary**,
**Unverified backlog**, and **Excluded** tables together are the full dedupe set. A job ID in
any of the three has been seen and must not be re-surfaced.

**Read the country filter.** Take **Countries** from *Location & work mode*. It is a hard
filter and it applies at three points, not one:

1. **In the query** (Step 1) — append the country/region to every search string so the boards
   filter server-side rather than you filtering 200 results by hand.
2. **At collection** (Step 2) — drop a candidate whose listed location is outside it before
   spending any verification time on it.
3. **At verification** (Step 4) — the posting's own stated location wins over the
   aggregator's. A board that says "Remote" and a JD that says "Remote (India)" is India.

A job outside the list is **excluded** with `reason: country`, never merely penalised. Same
for **Language**: a posting demanding fluency in a language not listed is excluded with
`reason: language`. Both go in the Excluded table so they are not re-surfaced.

`Anywhere` disables the country filter; say so in the run header when it is off.

**Read the limit.** Take **Result limit** and **Candidate pool multiplier** from
`jobs/input/config/search-profile.md`. A `--limit N` argument overrides the former for this run only —
never write it back to the file. Compute:

```
LIMIT  = --limit argument, else Result limit from config   (default 10)
TARGET = LIMIT × Candidate pool multiplier                 (default 30 candidates)
```

State both up front: `Target: 10 verified jobs (collecting ~30 candidates).` Everything
downstream is budgeted against these two numbers.

**Exhaustive mode.** `--limit all`, `--no-limit`, `--limit 0`, or `Result limit: all` in
config all mean the same thing: no count cap. Then:

```
LIMIT    = ∞          — no early stop in Step 4
TARGET   = ∞          — no pool cap in Step 2; sweep every enabled site
DEADLINE = now + Exhaustive run budget    (default 30 minutes)
```

Record `DEADLINE` as a real clock time (`date -v+30M +%H:%M` on macOS) and announce it:
`Exhaustive run — sweeping all 10 sites, hard stop at 15:42.` The budget is what guarantees
the run ends; never treat exhaustive as unbounded.

If `jobs/input/profile/master-resume.md` is still full of TODOs, say so and suggest `/job setup`
first; fit scores are meaningless without it. Offer to search anyway — a search with weak
scoring still surfaces real jobs, and the file can be re-scored later by `/job triage`.

**Unresolved filter inputs.** A hard filter whose input is `TODO` (work authorization,
sponsorship, relocation) or an `*(edit)*` placeholder (comp floor, metro, seniority) cannot
be evaluated. Do not treat it as pass and do not treat it as fail. Instead:

- Leave the filter **inactive** for this run and keep the affected jobs in the results.
- Record the term on the job's detail block as `unverified-filter: sponsorship` so
  `/job triage` and `/job apply` know it was never checked.
- List every inactive filter in the run header and in the final report, so the user knows
  which results are unfiltered.

Never silently drop a job on a filter you had no value for, and never claim a job passed
one.

## Step 1 — Build the queries

For each role family in `search-profile.md`, construct queries from **its own title list** —
the file is the only source of what to search for. Combine each family's titles into a
handful of high-yield strings rather than one query per title, OR-ing the synonyms and
AND-ing a disambiguating signal where a title is ambiguous across industries:

```
"<title A>" OR "<title B>" OR "<title C>"                 <location>
"<broad title>" ("<signal 1>" OR "<signal 2>")            <location>
```

Same shape whatever the field — the titles and signals come from the config, not from here.

**A query means different things per method.** For `ats-api` there is no query string at all
— you fetch whole boards and filter the results in memory, so the "query" is the
title-list + signals-list + country predicate from Step 2. Build search strings only for
`websearch`, `apify` and `browse`, and append the country to those:
`"<title A>" OR "<title B>" United States remote`.

Say what you are matching on before you run — the families, the signals, and the country —
so a run that returns nothing is obviously a filter problem and not a bug.

## Step 2 — Collect candidates

Order: **Tier 1 → Tier 5**, exactly as `job-sites.md` lists them. Tier order is a quality
ranking, not a formality — Tier 1 returns only live jobs, Tier 4 returns a crawler's index.

**Verify the backlog first.** Before sweeping any site, take the **Unverified backlog** table
from `jobs/output/jobs.md` as your starting pool — those candidates are already ranked and
cost nothing to re-find. Only sweep sites if the backlog cannot fill `LIMIT`.

**Stop when the pool is full.** Once you hold `TARGET` de-duplicated candidates, stop
sweeping and go to Step 3, even with sites left unsearched. Record which ones you skipped;
the next run starts with them so coverage rotates.

### Method: `ats-api` — the primary path

One request per company returns its **entire live board plus every job description**. Fetch
the enabled companies in parallel; a board is one call, not one call per job.

```bash
# Greenhouse — content=true is required, it is what carries the JD body
curl -s "https://boards-api.greenhouse.io/v1/boards/<slug>/jobs?content=true"
# Lever — body already present in descriptionPlain + lists
curl -s "https://api.lever.co/v0/postings/<slug>?mode=json"
# Ashby
curl -s "https://api.ashbyhq.com/posting-api/job-board/<slug>"
```

Then filter what came back, in this order:

1. **Country** — from the job's own location field (`location.name`, `categories.location`,
   `location`). This is the posting's own statement of where it is, so it is authoritative;
   drop anything outside **Countries**.
2. **Title OR signals** — match the title against the role family's title list, **or** the JD
   body against its signals list. **Both, not just the title.** Product companies title a
   role "Software Engineer — Backend" and name the stack only in the body: a sweep on
   2026-09-07 found 826 live US engineering roles with **zero** "Java" titles, while one
   company alone had 44 whose bodies required Java/Spring Boot. Title-only matching finds
   none of them.
3. **Recency** — `updated_at` / `createdAt` where the API gives it.

**Jobs from this method are `verified-live` without further checking.** The board API only
serves open reqs, so Step 4's liveness pass is redundant — record
`Proved by: <the API URL>` and skip 4a–4c for them. Still run 4d (real employer) and
4e (scam/ghost signals); an API tells you a req is open, not that it is honest.

This is what makes a run fast: no fetch-per-candidate, no dead links, no verification queue.

### Method: `themuse` — broad coverage beyond your board list

Tier 1 only knows the employers you listed. This is how a run reaches the ones you didn't —
banks, insurers, retailers, large non-tech employers — which is where a lot of enterprise
work lives.

```bash
curl -s "https://www.themuse.com/api/public/jobs?page=0&location=United%20States"
```

- No key. **500 requests/hour**, 20 results per page — check `page_count` and page through
  only as far as the candidate pool needs.
- Pass the country from the filter as `location` (`United States` is a valid whole-country
  value); `Flexible / Remote` for remote-only.
- **Do not pass `category` unless you have confirmed the exact string.** An unrecognised
  value returns `total: 0` with no error, which is indistinguishable from "no jobs" — the
  valid list is in `job-sites.md`. When in doubt omit it and match on title + signals
  yourself.
- `contents` is the JD **as HTML**: strip tags before matching signals.
- `publication_date` gives you recency for free.
- **A job can list several locations** — 9 of 20 on a sample page did. `locations[0]` is not
  necessarily the US one, so check the whole array and record the location that actually
  matched the filter. Reading only the first entry mislabels a US-eligible role as foreign.


**These are candidates, not verified jobs.** `refs.landing_page` is a themuse.com redirect,
not the employer's ATS, so it proves the posting exists but is not an apply URL. Run the full
Step 4 on them and resolve the real apply URL there — that is the one thing this method does
not hand you.

### Method: `api` — Hacker News "Who is Hiring?"

```bash
curl -s "https://hn.algolia.com/api/v1/search_by_date?tags=story,author_whoishiring&query=hiring&hitsPerPage=3"
curl -s "https://hn.algolia.com/api/v1/items/<objectID>"
```

Date-sorted and author-scoped — a plain relevance search returns threads from 2016. Each
top-level comment is one employer; parse role, location, remote-ness and the apply link out
of free text. Keep only comments matching the role families and the country filter. Follow
each to its real apply URL, and verify those normally.

### Method: `apify` — LinkedIn

Pass the search terms, the **country** from the filter, and a `postedLimit` matching the
recency setting; retrieve with `mcp__apify__get-dataset-items`. If the Actor reports that it
needs permission approval, **say so once in the final report with the approval link from
`job-sites.md`, mark the site skipped, and carry on.** Never stall the run waiting on it.

### Method: `websearch` — last resort

Only when Tiers 1–3 cannot fill `LIMIT`. Everything it returns is unverified and **must**
go through the whole of Step 4. Expect roughly half to be closed reqs and a large share of
staffing reposts with no named employer.

### Method: `browse`

For a specific careers page that has no API. Budget ~90 seconds or ~5 page loads per site.

```bash
$B goto "<search-url>"; $B wait --networkidle; $B text; $B links
```

### Per-site discipline

- Take at most `TARGET ÷ 3` candidates from any one site, so a single board can't fill the
  whole pool. A 900-job board is one site, not nine.
- On a login wall, bot challenge, or CAPTCHA: try once, then `$B handoff "<site> wants a
  login"`. If the user isn't around, mark the site skipped and continue. Never let one site
  block the run.
- When you learn something durable about a site (an iframe, a rate limit, a filter that
  silently does nothing), append a dated line to its `Notes` in `job-sites.md`. If it turns
  out to be useless, move it to **Disabled** there with the evidence.

### Stamp the Profile column

Every row carries a **Profile** cell naming the role family in `search-profile.md` that
matched it — `AI Engineer`, `Java Full Stack Developer`, `Data Engineer`. Use the family's
heading verbatim so the values stay groupable; a job caught by the open bucket is `Adjacent`.
This is what lets one index hold several job hunts at once and lets the dashboard filter
between them, so never leave it blank and never invent a family that isn't in the config.

## Step 3 — Pre-rank the candidate pool

The limit means the user gets the **best** N jobs, not the first N encountered. So rank the
whole pool before spending any verification time on it.

Score each candidate 0–100 from cheap signals only — no page fetches:

| Weight | Signal | Source |
|---|---|---|
| 40 | Title match against the focus argument and role families | candidate title |
| 20 | Employer quality — per the definition in `search-profile.md` | company name |
| 15 | Location / work-mode match | candidate location |
| 15 | Recency | posted date |
| 10 | Source tier — Tier 1 boards carry less noise | which site found it |

Drop anything already in the dedupe set, then sort descending. This provisional score only
decides *verification order*; the real score comes in Step 5 once the JD has been read.

`ats-api` candidates arrive with their JD already in hand, so they need no verification pass
and can be scored properly at Step 5 immediately. Rank them anyway — the limit still decides
which ones make the list.

## Step 4 — Verify down the ranking until the limit is hit

Never put a job on the list you haven't confirmed still exists. A stale posting wastes a
tailoring pass; a fake one harvests the user's address, phone, and work history.

**Candidates from `ats-api` skip 4a–4c.** A board API serves open reqs only, so their
liveness is already established — mark them `verified-live` with the API URL as
`Proved by`, and run only 4d and 4e on them. Everything from `websearch`, `browse`, or HN
goes through the whole loop below.

**This step is a loop, not a batch.** Walk the ranked pool from Step 3 top-down. For each
candidate, in order:

1. Resolve the **canonical apply URL** — aggregator links are for humans, not applications.
   Follow through to the company's ATS (`job-boards.greenhouse.io/...`, `jobs.lever.co/...`,
   `jobs.ashbyhq.com/...`). Record posting URL, apply URL, and which ATS. `/job apply`
   depends on this. The employer's **Careers** URL is captured separately in 4c.
2. Fetch the full JD (`$B goto <apply-url>` + `$B text`). Read the **actual posting**, not
   the aggregator's summary — snippets are routinely wrong about location, comp, and
   seniority.
3. Run checks 4a–4e below.
4. If it passes, add it to the list and **increment the verified count**.

**Stop the moment the verified count reaches `LIMIT`.** Leave the rest of the pool
untouched — unverified candidates are cheap to keep and get carried into the results file
as backlog (Step 6), so the next run resumes from there instead of re-searching.

**In exhaustive mode** there is no count to stop at, so two other rules take over:

- **Stop at `DEADLINE`.** Check the clock every 5 verifications. On reaching it, stop
  immediately, write the file with everything verified so far, and report the remainder as
  backlog. Never run past the budget.
- **Checkpoint every 10 verified jobs.** Write the results file incrementally rather than
  only at the end, so an interrupted or timed-out run keeps its work. Say so:
  `Checkpoint — 20 verified, merged into jobs/output/jobs.md`

Ranking still matters most here. Because Step 3 ordered the pool best-first, a run that
times out at 40 of 90 candidates has verified the best 40, not a random 40.

Announce progress as you go so a long run isn't silent:
`Verified 6/10 — checking sierra--ai-engineer…`
In exhaustive mode, report against the pool and the clock instead:
`Verified 23 · 41 of 90 candidates checked · 18 min left`

**If the pool runs dry before the limit:** say so plainly, report how many verified, and
offer to widen — sweep the sites skipped in Step 2, relax recency, or drop the focus
argument. Do not pad the list with jobs that failed verification to reach the number.

### 4a. Cheap liveness pass

```bash
curl -sS -L -o /dev/null -w '%{http_code}  %{url_effective}\n' --max-time 15 "<apply-url>"
```

- `404` / `410` → **dead**, drop it.
- `200` but `url_effective` landed somewhere generic — the board root, `/careers`,
  `/jobs`, a search page — → the req closed and the ATS redirected. **Dead**, drop it.
- Anything else → survives to 4b.

When several candidates are queued, batch these first — they are fast and cull the pool
before any browser time is spent. Note that a `200` alone proves nothing, since most ATS platforms
serve a styled "no longer accepting applications" page with a success status.

### 4b. Confirm on the page

```bash
$B goto "<apply-url>"
$B wait --networkidle
$B url                      # did it redirect after JS ran?
$B text | head -60
```

Mark **dead** on any of: "no longer accepting applications", "this position has been
filled", "position/req is closed", "job not found", "this posting has expired", "we are no
longer hiring for this role".

Then confirm the application is actually reachable:

```bash
$B snapshot -i | grep -iE 'apply|submit|resume|first name|attach'
```

No apply control and no form fields → the posting is a **description page only**. Keep it,
but mark it `apply-path: unconfirmed` so `/job apply` knows to hunt for the real form.

### 4c. Cross-check against the employer's own board (the strongest signal)

Aggregators keep serving reqs for months after they close. The company's own careers page is
the authority. Because the run is capped at `LIMIT` jobs, do this for **every** candidate
that reaches 4c — the volume is small enough to afford it. In exhaustive mode the volume is
not small: fall back to cross-checking only candidates provisionally scoring 60+, and mark
the rest `likely-live` rather than claiming a check you skipped:

```bash
$B goto "<company careers URL>"
$B wait --networkidle
$B links | grep -iE '<role keywords>'
```

**Record the careers URL you land on** — it goes in the **Careers** column and is the one
link that outlives this particular req. Resolve it in this order:

1. The employer's own careers page (`company.com/careers`), if it exists and loads.
2. Otherwise the **ATS board root** — `jobs.ashbyhq.com/<company>`,
   `jobs.lever.co/<company>`, `job-boards.greenhouse.io/<company>`. For many startups this
   *is* their careers page: their site's "Careers" link points straight at it.
3. If neither resolves, `—`. Never invent a careers URL by guessing a path — check it.

Common traps worth one extra probe before giving up: `/career` singular, `/about/careers`,
and a `403` from `curl` that is bot-blocking rather than a missing page (confirm with `$B`).

- Present on the company's board → `verified-live`
- Absent, but the ATS page is live and has a form → `likely-live`
- Absent and the aggregator is the only source → `stale`, drop it

Prefer the company's own apply URL over the aggregator's whenever both exist.

### 4d. Is the employer real?

Drop or flag on these:

- **Unnamed employer** — "Confidential", "Our client", "A leading AI company". Staffing-agency
  repost. Drop unless the user has said they want agency listings.
- **Apply path off-ATS** — application by emailing a Gmail/Outlook/Proton address, a Google
  Form, or a domain unrelated to the company. Flag `suspicious`.
- **No web presence** — the company has no site, or the site is a single page created
  recently. Flag `suspicious`.
- **Comp wildly out of band** for the role and level — a classic bait pattern. Flag.

### 4e. Ghost-job and scam red flags

A **ghost job** is real-looking but not actually being hired for. Signals:

- The job has sat in `jobs/output/jobs.md` across runs while still advertising a fresh "posted" date.
  Every detail block records `First seen: YYYY-MM-DD` from the run that found it; compare it
  against the posting's currently claimed date:
  ```bash
  grep -A3 '<job-id>' jobs/output/jobs.md | grep 'First seen'
  ```
  First seen 60+ days ago and still claiming a recent posting date → flag
  `evergreen — possibly a ghost posting`.
- "Always accepting applications", "building a pipeline", "future opportunities", talent-pool
  and general-interest reqs → mark `pipeline`, not a live opening.
- Posted 90+ days ago and still open with no edits.

**Scam signals — these are a hard drop, never surface them:**

- Asks for SSN, date of birth, passport, driver's licence, or bank/routing details *in the
  application itself* (legitimate employers collect these after an offer, never at apply)
- Asks for any payment — equipment, training, background-check fees
- Interviews conducted only over Telegram, WhatsApp, or Signal
- An offer with no interview

Log a hard-dropped scam posting in the results file's skipped list with the reason, so the
same listing isn't re-surfaced next run.

### 4f. Record the verdict

Every job carries one of these into the results file:

| Verdict | Meaning |
|---|---|
| `verified-live` | On the employer's own board, form reachable, checked today |
| `likely-live` | ATS page live with a form, not cross-checked |
| `unverified` | Reachable but liveness could not be established — say why |
| `pipeline` | Real, but a talent pool rather than an open req |
| `stale` / `dead` / `suspicious` / `scam` | Excluded, with the reason recorded |

Only `verified-live`, `likely-live`, and `unverified` reach the summary table. `pipeline`
goes in a separate section at the bottom. The rest appear only as a count plus reason in
the run header.

**Verification budget:** ~20 seconds per job. If a site is slow or rate-limiting, mark the
affected jobs `unverified` with the reason and move on — never let verification stall the
run, and never upgrade a verdict you didn't actually establish.

## Step 5 — Score and dedupe

Score each job 0–100 with the weights in `search-profile.md`. Be honest — an inflated score
wastes the user's time later.

Apply the Step 4 verdict to the score: `verified-live` unchanged, `likely-live` −5,
`unverified` −15, `evergreen`/ghost-flagged −20. A job the agent could not confirm should
not outrank one it did.

Dedupe by job ID (`<company-slug>--<role-slug>`). The same job posted on four boards is one
row; keep the best apply URL and note the other sources. If a job ID already exists in an
any of `jobs/output/jobs.md`'s three tables, or in `jobs/output/tracker.md`, exclude it and count it as
a duplicate in the header. Apply the hard filters — a job failing one is excluded, with the
reason recorded.

## Step 6 — Write the results file

```bash
cat jobs/input/templates/job-results.md    # the section shapes jobs/output/jobs.md uses
```

**Everything goes into `jobs/output/jobs.md` — one living file, never a dated one.** Merge into it; do
not rewrite it from scratch. The **Status** column is the point of the file and must survive
every run.

Merge rules:

- **New verified job** → append a row to **Summary** with `Status: found` and `—` in all
  three stage columns (**JD**, **Resume**, **Form**), then add its detail block under
  `## Details`. Keep the row's column count identical to the header — a short row breaks the dashboard.
- **Already in Summary** → leave the row alone. Refresh `Verified` and `Comp` if they
  changed, but **never reset a Status** that has moved past `found`, and **never touch the
  JD / Resume / Form columns** — those belong to `/job jd`, `/job tailor` and
  `/job apply`.
- **Promoted from backlog** → move the row out of **Unverified backlog** into **Summary**,
  carrying its Status across.
- **Failed verification** → move it to **Excluded** with the verdict and reason. If the job
  had a captured JD, delete `jobs/output/jds/<job-id>.md` along with it.
- Update the header block: `Last updated`, `Last search`, counts, `Sites unswept`,
  `Filters inactive`.

`/job search` captures a *summary* of each posting into `## Details` — enough to rank and
triage. It does **not** write `jobs/output/jds/<job-id>.md`; the full JD is `/job jd`'s
job, run once the user has shortlisted. Hand off with
`Next: /job triage`, then `/job jd`.

**Carry the backlog.** Everything ranked in Step 3 but never reached in Step 4 goes into the
**Unverified backlog** table — provisional score, role, company, location, source, URL, and
`Status: found`. This is what makes the limit cheap rather than lossy.

The detail block matters more than the table. "Why it fits" must reference actual lines from
`jobs/input/profile/master-resume.md`, and "Gaps" must be honest — that section is what `/job tailor`
and `/job apply` use to avoid overclaiming.

## Step 7 — Report

```
Limit 10 reached — stopped early.

  Collected  31 candidates from 5 of 10 sites (stopped once the pool was full)
  Verified   14 checked → 10 on the list, 4 excluded
             (2 closed reqs, 1 agency repost, 1 evergreen ghost)
  Backlog    17 candidates carried forward, 5 sites unswept

  On the list:  6 verified-live · 3 likely-live · 1 unverified
  Apply now (80+): 3  ·  Worth a look (60+): 5  ·  Stretch (40+): 2
  Filters inactive: sponsorship, comp floor  (no value in the answer bank)

  Top: Forward Deployed Engineer @ Anthropic (91) · AI Engineer @ Sierra (86)

  → jobs/output/jobs.md   (view: make web  →  http://localhost:8080)
  Next: /job triage to shortlist, or /job apply <job-id> to start one
```

Exhaustive runs report against coverage and the clock instead:

```
Exhaustive run complete — 26 min of a 30 min budget.

  Swept      all 10 enabled sites · 94 candidates collected
  Verified   94 checked → 71 on the list, 23 excluded
             (11 closed reqs, 7 agency reposts, 4 evergreen ghosts, 1 scam)
  Backlog    none — pool fully verified

  On the list:  38 verified-live · 29 likely-live · 4 unverified
  Apply now (80+): 9  ·  Worth a look (60+): 24  ·  Stretch (40+): 38
```

If the deadline hit first, say so and give the remainder plainly:
`Stopped at the 30 min budget — 52 of 94 candidates verified, 42 carried as backlog.
Re-run to continue from the backlog, or raise Exhaustive run budget.`

Say the limit was the reason the run stopped, so an early finish never reads as a thin
search. If the user wants more, the fix is one line — raise **Result limit** in
`jobs/input/config/search-profile.md`, or re-run with `--limit N`; with a backlog on file the next run
is much faster than the first.

The file is the deliverable — don't reprint the jobs in the terminal.
