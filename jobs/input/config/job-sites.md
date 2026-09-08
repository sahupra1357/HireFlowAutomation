# Job Sites to Search

`/job search` reads this file. Add, remove, or toggle sites here — no task edits needed.

Sites are tried **in tier order**, and the run stops sweeping once the candidate pool is
full. So tier order is a quality ranking: what is listed first is what you mostly get.

---

## Tier 1 — ATS board APIs  ⭐ the primary source

- Enabled: yes
- Method: `ats-api`

**Every job these return is live.** They are the boards' own public JSON endpoints — the same
data the careers page renders — so a closed req simply is not in the response. That removes
the single biggest source of waste: a `site:` web search returns a search-engine *index*,
which in a 2026-09-07 run was **44% dead reqs** (12 of 27 candidates).

They also return the **full job description in the same request**, so a role can be matched on
the *signals* in its JD body, not just words in its title. That matters more than it sounds:
a sweep of these boards found 826 live US engineering roles and **zero** with "Java" in the
title — but 44 at one company alone mentioned Java/Spring Boot in the body. Title-only
matching would have found none of them.

| ATS | Endpoint | Notes |
|---|---|---|
| Greenhouse | `https://boards-api.greenhouse.io/v1/boards/<slug>/jobs?content=true` | `content=true` is required for body matching. `location.name` holds the location. |
| Lever | `https://api.lever.co/v0/postings/<slug>?mode=json` | Body is in `descriptionPlain` + `lists`; location in `categories.location`. |
| Ashby | `https://api.ashbyhq.com/posting-api/job-board/<slug>` | Add `?includeCompensation=true` for pay bands. |

No key, no scraping, no rate limit hit in practice. One request per company returns its whole
board — fetch them in parallel.

### Companies

**This list is the main dial for search quality.** Add the employers you'd actually work for;
delete the rest. Every slug below was probed on 2026-09-07 and returned a live board — the
count is what it held that day, as a sanity check, not a promise.

To find a company's slug, open their careers page and read the ATS URL:
`job-boards.greenhouse.io/`**`tebra`** · `jobs.lever.co/`**`veeva`** · `jobs.ashbyhq.com/`**`ramp`**.
Verify a new one before relying on it:

```bash
curl -s "https://boards-api.greenhouse.io/v1/boards/<slug>/jobs" | head -c 200
```

#### Greenhouse
| Company | Slug | Jobs (2026-09-07) | Enabled |
|---|---|---|---|
| Databricks | `databricks` | 873 | yes |
| Stripe | `stripe` | 618 | yes |
| Anthropic | `anthropic` | 589 | yes |
| Datadog | `datadog` | 442 | yes |
| Cloudflare | `cloudflare` | 336 | yes |
| Samsara | `samsara` | 257 | yes |
| Roblox | `roblox` | 230 | yes |
| Affirm | `affirm` | 205 | yes |
| Coinbase | `coinbase` | 192 | yes |
| Flexport | `flexport` | 172 | yes |
| Airbnb | `airbnb` | 167 | yes |
| Twilio | `twilio` | 149 | yes |
| Reddit | `reddit` | 148 | yes |
| Robinhood | `robinhood` | 128 | yes |
| Asana | `asana` | 115 | yes |
| Instacart | `instacart` | 110 | yes |
| Gusto | `gusto` | 95 | yes |
| Chime | `chime` | 65 | yes |
| Discord | `discord` | 48 | yes |
| Dropbox | `dropbox` | 43 | yes |
| Tebra | `tebra` | 24 | yes |
| Blend | `blend` | 9 | yes |

#### Lever
| Company | Slug | Jobs (2026-09-07) | Enabled |
|---|---|---|---|
| Veeva | `veeva` | 900 | yes |
| Palantir | `palantir` | 310 | yes |
| Spotify | `spotify` | 78 | yes |
| Everbridge | `everbridge` | 21 | yes |
| Leadership Connect | `leadershipconnect` | 8 | yes |
| Resilinc | `resilinc` | 7 | yes |
| Termgrid | `termgrid` | 5 | yes |

#### Ashby
| Company | Slug | Jobs (2026-09-07) | Enabled |
|---|---|---|---|
| OpenAI | `openai` | 781 | yes |
| Clera | `clera` | 250 | yes |
| Cohere | `cohere` | 143 | yes |
| Ramp | `ramp` | 142 | yes |
| Notion | `notion` | 132 | yes |
| Vanta | `vanta` | 111 | yes |
| Ashby | `ashby` | 71 | yes |
| Sift | `siftstack` | 48 | yes |
| CodeRabbit | `coderabbit` | 47 | yes |
| Allica Bank | `allica-bank` | 34 | yes |
| Modal | `modal` | 30 | yes |
| Linear | `linear` | 29 | yes |
| Peek | `peek` | 10 | yes |
| Tradeify | `tradeify` | 6 | yes |
| Runway | `runway` | 4 | yes |

*8,182 live postings across 44 boards as of 2026-09-07.*

---

## Tier 2 — Broad free APIs

Tier 1 covers employers you named. This tier covers the ones you didn't — banks, insurers,
retailers, consultancies, big non-startup tech. That is where a lot of enterprise work
(Java/.NET/Oracle especially) actually lives, and a curated ATS list structurally misses it.

### The Muse

- Enabled: yes
- Method: `themuse`
- Endpoint: `https://www.themuse.com/api/public/jobs`
- **No API key.** Rate limit **500 requests/hour** (`X-Ratelimit-Limit: 500`, resets in 3600s),
  20 results per page — ~10,000 jobs/hour, far more than any run needs.

```bash
curl -s "https://www.themuse.com/api/public/jobs?page=0&location=United%20States&category=Software%20Engineering"
```

| Param | Notes |
|---|---|
| `location` | `United States` works as a whole-country value (~6,470 jobs). `Flexible / Remote` for remote-only. Repeat the param for several. |
| `category` | **Exact match, and an unknown value returns `total: 0` with no error** — see the trap below. |
| `level` | `Entry Level` · `Mid Level` · `Senior Level` · `Internship` |
| `page` | 0-indexed; the response carries `page_count` and `total`. |

**⚠️ The category trap.** `category=Engineering` returns **0**. So does `Nursing`, and so does
any typo. A wrong category looks exactly like "no jobs today" rather than an error, which is
how a run silently returns nothing. Use only values observed in live data:

> Software Engineering · Data and Analytics · Science and Engineering · Computer and IT ·
> Design and UX · Product Management · Project Management · Business Operations · Sales ·
> Account Management · Legal Services · Education · Customer Service · Writing and Editing ·
> Administration and Office · Construction · Management

When unsure, **omit `category` entirely** and filter on title + signals yourself. Verify a new
value returns a non-zero `total` before trusting a run that used it.

**Two more things to know:**

- `contents` is the full JD **as HTML** — strip tags before matching signals against it.
- `refs.landing_page` is a **themuse.com redirect page, not the employer's ATS**. It proves
  the job exists but is not an apply URL. `/job jd` must follow it through to the real ATS
  form before the row is usable by `/job apply`.
- `publication_date` is present, so the recency filter works without fetching anything.
- **The apply URL is behind a JS button.** `refs.landing_page` renders an "Apply on company
  site" *button*, not a link, and it opens a new tab — so the employer URL is not in the
  static HTML for every posting. About half resolve by regexing the page for a known ATS
  host (Workday, Greenhouse, iCIMS, Lever); the rest must be marked `apply-path: unconfirmed`
  and resolved by `/job jd`. Budget for that: Muse rows cost more to make applyable than
  Tier 1 rows, which arrive with a real ATS URL already.
- **A job can list several locations** — 9 of 20 on a sample page did. `locations[0]` is not
  necessarily the US one, so check the whole array and record the location that actually
  matched the filter. Reading only the first entry mislabels a US-eligible role as foreign.


**2026-09-07:** verified. 4,286 US Software Engineering jobs; page 1 alone returned Bank of
America, Uber and Flexport — none of which are on any Tier 1 board.

### Adzuna *(optional — needs a free key)*

- Enabled: no
- Method: `api`
- Endpoint: `https://api.adzuna.com/v1/api/jobs/us/search/1?app_id=<ID>&app_key=<KEY>&what=<query>`
- **2026-09-07:** reachable — returns `401` with dummy credentials, so it is live and only
  wants a key. Register free at <https://developer.adzuna.com/>, put the ID and key here, and
  set Enabled to yes. Real US search with location and salary filters, broad aggregation.

### Hacker News "Who is Hiring?"

- Enabled: yes
- Method: `api`
- Newest thread (author-scoped and date-sorted — a plain relevance search returns threads
  from 2016):

```bash
curl -s "https://hn.algolia.com/api/v1/search_by_date?tags=story,author_whoishiring&query=hiring&hitsPerPage=3"
curl -s "https://hn.algolia.com/api/v1/items/<objectID>"     # then read the comments
```

- Notes: posted on the 1st of each month; each top-level comment is one employer. Unstructured
  free text — parse company, role, location, and remote-ness out of it. Very high signal,
  always current, and free. **2026-09-07:** verified working; the September 2026 thread
  (`id=49522897`) had 350 comments.

## Tier 3 — LinkedIn

- Enabled: yes
- Method: `apify` (`mcp__apify__harvestapi--linkedin-job-search`)
- **BLOCKED until approved once.** The Actor needs account-level permission before it will
  run. Approve it here, then it works for every later run:

  <https://console.apify.com/actors/zn01OAlzP853oqn4Z?approvePermissions=true>

- Notes: the largest single source by volume. Pass the search terms, location, and a
  `postedLimit` matching the recency filter; retrieve with `mcp__apify__get-dataset-items`.
  While unapproved, the run **skips it and says so once** — never stall waiting on it.

---

## Tier 4 — Web search by `site:`  ⚠️ last resort

- Enabled: yes
- Method: `websearch`
- Queries: `site:job-boards.greenhouse.io "<title>" <signal>` · `site:jobs.lever.co …` ·
  `site:jobs.ashbyhq.com …`
- **Use only when Tiers 1–3 cannot fill the limit**, and treat everything it returns as
  unverified. It searches a crawler's index, not the boards, so it cannot know whether a req
  is open, cannot filter by country or date, and over-represents staffing firms that SEO
  hard. In the 2026-09-07 run it produced 12 dead reqs and 7 agency reposts out of 27.
- Every candidate from here **must** go through the full Step 4 verification.

---

## Tier 5 — Company career pages (direct)

For employers whose ATS slug you don't have, or who host their own board. Add rows freely.

| Company | Careers URL | Enabled |
|---|---|---|
| *(add your targets — this tier is empty by default)* | | no |

---

## Disabled — tried and found wanting

Kept with the reason so they are not re-added by mistake.

| Site | Why |
|---|---|
| ai-jobs.net | **2026-08-30:** URL keyword params do not filter — `?key=` and `?q=` both return the full 48k listing (a delivery-driver post came back for "AI Engineer"). Search is client-side only. |
| Arbeitnow | **2026-09-07:** free and keyless, but EU/Germany-focused — **0 of 250** rows on page 1 were US. Useless while `Countries: United States`. |
| RemoteOK | **2026-09-07:** `?tag=java` is ignored — returned "Quality Checker" and "Project Systems Specialist" in Cayman Islands, UK and India. No usable filtering. |
| Indeed | **2026-09-07:** `403` + Cloudflare CAPTCHA on job search. The public Publisher API (`api.indeed.com/ads/apisearch`) is retired — the host no longer resolves. What remains is partner-only, requiring an employer/ATS agreement. Blocked to automated access, so **do not work around it** (rule 8). Most of its unique content is staffing reposts the Named-employer filter drops anyway. |
| Monster | **2026-09-07:** `403` + CAPTCHA on job search; `api.monster.com` no longer resolves. Same conclusion as Indeed, with far less volume. |
| Wellfound | Login-walled after a few listings; expect a handoff on first use each session. |
| Workday | Multi-step wizard, account required, heavy dynamic IDs. Only worth it for a specific target employer. |
| Y Combinator — Work at a Startup | Requires a free account to see contact details. Re-enable if you have one. |

---

## Adding a site

Copy a block, fill in the endpoint or URL, set `Enabled: yes`. When the agent hits something
surprising (a login wall, an iframe, a rate limit, a filter that silently does nothing), it
appends a dated line to that site's Notes so the next run is faster — and, if the site turns
out to be useless, moves it to **Disabled** with the evidence.
