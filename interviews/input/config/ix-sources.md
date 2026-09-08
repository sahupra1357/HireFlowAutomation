# Interview Experience Sources

`/job interviews` reads this file. Add, remove, or toggle sources here — no skill edits
needed. Same contract as `jobs/input/config/job-sites.md`: `Enabled: yes|no` gates the source,
`Method` tells the agent how to reach it, `Notes` carries gotchas the agent learned and
may append to.

This file is **only** about interview-experience intel. It never affects `/job search`.

---

## Phase 1 — free, no-auth, no-cost (enabled)

### Reddit
- Enabled: yes
- Method: api (RSS search)
- URL: `https://www.reddit.com/r/<sub>/search.rss?q=<query>&restrict_sr=1&sort=new&t=year`
- Fetch: `curl -sL -H 'User-Agent: ix-research/0.1 (personal job-prep script)'`
- Notes: **2026-09-01:** the old `search.json` endpoint now returns an HTML shell (403 on a
  default UA, HTML on a custom one) and `old.reddit.com` redirects to a login wall. The
  **`.rss` endpoint still works** and is the only reliable free path. It returns
  `<entry>` blocks carrying title, permalink, `<updated>` ISO date, and the **full selftext
  body** in `<content type="html">` — so a search costs exactly one request per query, with
  no per-post fetch needed. Caps out around 25 entries per query; page by varying the
  query, not by an offset param. RSS carries **no score and no comment count**, which is
  why the credibility model scores firsthand-ness instead of engagement.
  **2026-09-01 (first run):** rate limiting is aggressive — a second request fired
  immediately after the first returns **429**. `ix-fetch.py` now paces reddit calls 12s
  apart via a stamp file in the temp dir and backs off 5/10/20s on a 429; with that, 12
  consecutive queries ran clean. Do not lower `PACE["reddit"]`. Budget ~15s per query
  when estimating a run.
  **Yield, first run:** 12 subreddit queries returned 149 candidates, of which 8 became
  reports. r/recruitinghell was by far the richest for track B; r/leetcode returned
  almost pure noise (general FAANG threads, no AI roles) and is a candidate for removal.

#### Subreddits — track A (interviewing at a company)
`ExperiencedDevs`, `cscareerquestions`, `leetcode`, `csMajors`, `MachineLearning`,
`datascience`, `dataengineering`, `interviews`

#### Subreddits — track B (interviewed by AI)
`recruitinghell`, `ExperiencedDevs`, `cscareerquestions`, `jobs`, `interviews`

### Hacker News
- Enabled: yes
- Method: api (Algolia)
- URL: `https://hn.algolia.com/api/v1/search_by_date?query=<q>&tags=<story|comment>&hitsPerPage=30`
- Notes: Free, no key, no rate pain. Returns `points`, `num_comments`, `created_at_i`,
  and `comment_text` / `story_text` inline. Lower volume than Reddit but higher signal for
  AI-lab loops. Restrict by time with `numericFilters=created_at_i>N`.

### Open web / blogs
- Enabled: yes
- Method: websearch (`WebSearch` → `WebFetch` on the hits)
- Notes: Where the single best long-form writeups live (personal blogs, Medium,
  Substack). Low volume, high value. Skip aggregator SEO spam — sites that list
  "top 50 interview questions" are content farms, not experiences.

---

## Phase 2 — walled or metered (disabled by design)

Kept here so the shape is right and enabling one is a one-word edit. Do not enable
without deciding on the ToS and cost questions first.

### Glassdoor (Interviews tab)
- Enabled: no
- Method: browse (`$B`) + handoff for the login wall
- Notes: The most *structured* source that exists — rounds, difficulty, outcome, questions,
  per company. Cloudflare + login walled. Phase 2, and only for shortlisted companies.

### Blind (teamblind.com)
- Enabled: no
- Method: websearch (indexed pages only — never drive the site)
- Notes: High signal, low civility, unreliable narrators. Read what Google indexed;
  do not attempt the login wall.

### LinkedIn posts
- Enabled: no
- Method: apify
- Notes: Metered. Content skews performative — "thrilled to share" writeups rarely name a
  real question. Low priority.

### X / Twitter
- Enabled: no
- Method: apify
- Notes: Metered, costs money. Most recent signal for AI-lab loops. Ask the user before
  enabling — this one has a bill attached.
