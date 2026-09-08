# Task: `interviews`

Run via `/job interviews`. Formerly the standalone `interview-search` skill.

Search Reddit, Hacker News, and the open web for real interview experiences — both interviewing AT AI companies (track A) and being interviewed BY an AI screener (track B) — score them for credibility, and merge the survivors into interviews/output/interviews.md. Use when the user wants interview intel, prep for an upcoming loop, or asks what a company's interview process is actually like.

**Tool budget for this task:** Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion, WebSearch, WebFetch

Reads `jobs/output/jobs.md` for company names and **never writes to it**. Writes only under `interviews/output/`.

Anything outside that budget means you are running the wrong task — return to
`SKILL.md` and route again rather than reaching for the tool.

---

Compile real, dated, firsthand interview experiences into
`interviews/output/interviews.md`. This is intel gathering, not job hunting: it reads
`jobs/output/jobs.md` to know which companies matter, and **never writes to it**.

**The point is signal, not volume.** Most of what social media calls an "interview
experience" is self-promotion, venting, or news-about. A run that reports 4 good
experiences and 26 hard drops did its job. A run that reports 30 experiences did not.

**Optional arguments:**
- A company slug or name — `/job interviews anthropic` — targets track A at one company.
- `--limit N` · `--limit all` (also `--no-limit`, `--limit 0`) — exhaustive, clock-bounded.
- `--track a|b|both` · `--source reddit,hn,web` · `--since 6m`

## Hard rules

1. **Never fabricate** a round, a question, a difficulty, a timeline, or an outcome.
   Absent means `—`. Tailoring the truth is what `/job tailor` is for; this skill only
   records what someone actually wrote.
2. **Never assert a single anonymous post as fact.** One report = `single-source`, labelled
   in the table and in the Details block, forever. Two independent reports that agree =
   `corroborated`.
3. **No PII.** Never record a username, handle, real name, or employer-identifying detail
   about the *poster*. The URL is the entire attribution.
4. **Verbatim quotes only, <=40 words, each with its URL.** Never smooth a badly-written
   question into a crisp one — the sloppiness is evidence.
5. **Read-only toward `jobs/output/jobs.md`** and `jobs/output/applications/`. This skill never sets a job
   Status or adds a job row.
6. **Human-scale volume.** One request per query. If a source blocks us, record it in that
   source's Notes and move on — never route around a block.
7. **18-month hard cutoff** (or `--since`). A stale loop writeup is worse than none.

## Step 0 — Load context

```bash
cat interviews/input/config/ix-profile.md
grep -B2 -A8 'Enabled: yes' interviews/input/config/ix-sources.md
grep -oE 'ix--[a-z0-9-]+--[a-z0-9-]+--[a-z]+--[0-9a-f]{8}' interviews/output/interviews.md | sort -u   # dedupe set
grep -oE '^\| [0-9]+ \| [0-9]+ .*' jobs/output/jobs.md | head -30                                            # company seed
```

Read the profile fully. Load every `ix--` ID already in `interviews/output/interviews.md` —
Reports, Backlog, and Excluded together are the full dedupe set. Also load every URL hash;
a post already excluded must not be re-collected.

**Read the limit** from `ix-profile.md`, exactly as `/job search` does:

```
LIMIT  = --limit argument, else Result limit from config      (default 10)
TARGET = LIMIT x Candidate pool multiplier                    (default 30 candidates)
```

State both up front: `Target: 10 compiled experiences (collecting ~30 candidates).`

**Exhaustive mode** (`--limit all` / `--no-limit` / `--limit 0` / `Result limit: all`):
`LIMIT = TARGET = infinity`, and `DEADLINE = now + Exhaustive run budget`. Record the
deadline as a real clock time (`date -v+30M +%H:%M`) and announce it. The budget is what
guarantees the run ends.

**Seed the company list** for track A from `jobs/output/jobs.md` — Summary rows first, then
backlog. A company argument overrides the seed entirely. Without a seed, track A is
unbounded and will drown in noise.

## Step 1 — Build the queries

Use the query shapes in `ix-profile.md`. Announce the queries before running them.

- **Track A:** one query per seeded company, across the track-A subreddits. Cap at the top
  8 companies by Fit unless targeted — 30 companies x 8 subreddits is not a search, it is
  a scrape.
- **Track B:** the vendor and generic queries, across the track-B subreddits. Company-agnostic.

## Step 2 — Collect (cheap pass)

**Verify the backlog first.** Rows in `## Unverified backlog` are already ranked and cost
nothing to re-find. Only search sources if the backlog cannot fill `LIMIT`.

The fetcher normalizes both APIs to one JSON object per line —
`{id, source, origin, title, url, posted, body}`:

```bash
./interviews/bin/ix-fetch.py reddit --sub ExperiencedDevs --q '"AI interview"' --since 18m --limit 25
./interviews/bin/ix-fetch.py hn --q '"interview experience" Anthropic' --since 18m --limit 30
```

Reddit RSS returns the **full selftext body** inline, so a track-A/B post needs no second
fetch. Only follow a link when the body is a stub pointing at a blog.

**Exit codes matter.** `0` with no lines = searched, found nothing (a real answer).
`2` = source unreachable — record it in the run header and in that source's Notes; never
report it as an empty result.

For `web`: `WebSearch` the query, then `WebFetch` only the hits that look like a personal
writeup. Skip anything shaped like a content farm.

**Dedupe as you collect** — by URL hash, and for HN by story: Algolia returns one hit per
*comment*, so five comments on one thread are one candidate, not five.

**Stop at `TARGET`** and go to Step 3, even with sources left unswept. Record which ones
were skipped so the next run starts there and coverage rotates.

## Step 3 — Junk filter (this is the whole game)

Apply the hard drops from `ix-profile.md` to every candidate **before** spending any
attention on it. Each drop is recorded with its reason label and counted in the header.

The two that will dominate the drop count:

- **`self-promo`** — someone launching an interview-prep tool, AI interview coach,
  bootcamp, or resume service. Endemic in exactly the subreddits worth searching. Tell:
  the post ends in a link to a product, or the "experience" is a feature list.
- **`news-not-experience`** — a post *about* AI interviewing (an article, a hot take, a
  "this is awful" reaction to something read elsewhere). Track B needs someone who **sat
  through one**. A reaction to a news story is not an interview experience, no matter how
  on-topic it reads.

Ask of every survivor: **did this person sit in this interview, and can I tell what
happened in it?** If no to either, it is not a report.

## Step 4 — Read and extract

For each survivor, pull only what is actually stated:

- Company (track A) or vendor (track B)
- Role / topic
- **When the interview happened** — not when the post was written. If only the post date
  exists, use it and flag `posted-date-only` in Details.
- Rounds (integer or `—`), format, timeline, outcome
- Questions actually reported — verbatim, <=40 words, each with the URL

## Step 5 — Score and dedupe

Score 0–100 on the five weights in `ix-profile.md` (specificity 30 · recency 25 ·
corroboration 20 · firsthand 15 · source quality 10). Be honest: an inflated credibility
score is worse than a missing report, because it gets trusted before a real interview.

Corroboration is computed **across the whole index**, not just this run — a new report may
corroborate an existing single-source one, which raises both. When it does, update the
older entry's score and its `single-source` label.

Assign the ID: `ix--<company-slug>--<topic-slug>--<source>--<url-hash8>`. The `ix--`
prefix guarantees no collision with a job ID anywhere.

Anything scoring below **Min credibility** goes to Excluded as `low-credibility`.

## Step 6 — Merge

```bash
cat interviews/input/templates/ix-results.md    # the exact section shapes
```

**Merge into `interviews/output/interviews.md`. Never rewrite it from scratch.**

- New report → append to `## Reports` + a block under `## Details`.
- Already present → leave it; refresh only Cred and the corroboration label.
- Promoted from backlog → move the row, carrying its ID.
- Failed the filter → `## Excluded` with the reason label.
- Re-number `#` by Cred descending after merging.
- Update the header block: `Last updated`, `Last search`, counts, `Sources swept`.

**Job ID column** — fill it only when the company matches a row already in `jobs/output/jobs.md`.
Never invent one.

If the run was company-targeted, also write
`interviews/output/reports/<company-slug>/prep.md` — the compiled brief: the loop as reported,
questions by round, what is corroborated vs. single-source, and what is simply unknown.
The unknowns matter as much as the knowns.

## Step 7 — Report

Keep the terminal short. The file is the deliverable.

```
Limit 10 reached — stopped early.

  Collected  34 candidates · reddit 22 · hn 9 · web 3
  Filtered   24 dropped (11 self-promo, 7 news-not-experience, 4 venting, 2 too-old)
  Compiled   10 reports · 6 track A · 4 track B

  Credibility:  80+ 2 · 60-79 5 · 40-59 3
  Corroborated 3 · single-source 7
  Sources unswept: r/leetcode, r/datascience  (next run starts here)

  Best: Anthropic FDE loop (82, corroborated) · HireVue async screen (74)

  → interviews/output/interviews.md   (view: make web → http://localhost:8080 → Interviews)
```

**Say the drop count and its reasons out loud.** A high drop rate is the skill working,
and the user needs to see the ratio to judge whether a source is worth keeping enabled.

If a run compiles almost nothing, say that plainly and name the reason — a source blocked
us, the query was too narrow, or the material genuinely is not out there. Never pad the
Reports table to make a run look productive.
