# Search Profile

What counts as a job worth surfacing. `/job search` and `/job triage` both read this.
Edit freely — this is the main dial for tuning result quality.

**Nothing here is domain-specific by design.** The role families below are whatever *you*
are looking for — the search machinery builds its queries from the titles you list, so a
nurse practitioner, a controller, and an LLM engineer all use the same file. `/job setup`
proposes a starting set by reading your resume; edit them afterwards.

---

## Target role families

List as many as you want — one section per family, numbered. Delete the placeholders and
write your own. A family needs a **title list** (what the posting is actually called, in all
its variants) and a **signals list** (phrases in the JD body that confirm it's the right
kind of role, not just the right words in the title).

### 1. {{Primary role family}}
- Titles: {{the exact title, plus every synonym and seniority variant you'd accept}}
- Signals: {{tools, responsibilities, or phrases that appear in the JD body}}

### 2. {{Second role family}}
- Titles: {{...}}
- Signals: {{...}}

### 3. {{Adjacent / open bucket — catch what the market invents}}
- Titles: {{emerging or non-standard titles for work you'd still take}}
- Rule: if a title is new-to-the-market but the JD is clearly the work you want, surface it
  under this bucket rather than dropping it. Note the new title here so the list can grow.

*(Add families 4, 5, … as needed. There is no fixed count — `/job search` iterates whatever
is listed.)*

---

## Keywords to boost (higher fit score)
{{Skills, tools, and phrases from your resume that a good posting would also mention.}}

## Keywords to exclude (drop or heavily penalize)
{{e.g.}} `unpaid`, `internship` *(unless you say otherwise)*, `commission only`,
`clearance required` *(unless you hold one)*, `{{degree}} required` *(if you lack it)*,
`{{N}}+ years` *(if beyond your experience)*, staffing-agency reposts with no named company

---

## Hard filters (a job failing any of these is `skipped`, with the reason logged)
- **Work authorization:** must not require citizenship/clearance you lack
  (see `jobs/input/config/application-answers.md`)
- **Sponsorship:** if you need it, the posting must not say "no sponsorship"
- **Location:** see below
- **Posted within:** last 30 days (stale postings are usually filled)
- **Must verify as live:** the posting must survive `/job search` Step 4 — reachable apply
  page, no "no longer accepting applications", and for 60+ scores, present on the employer's
  own careers page
- **Named employer:** no "Confidential" / "Our client" staffing reposts
- **Legitimate apply path:** a real ATS or the company's own domain — never a personal email
  address or a bare Google Form

## Scam hard-drops (never surfaced, regardless of fit)
Asks for SSN, DOB, passport, or bank details at application time · asks for any payment ·
interviews only over Telegram/WhatsApp/Signal · an offer with no interview.

## Location & work mode
- Preferred: {{Remote (country) | Hybrid | your metro}}
- Open to relocation: **ASK THE USER** — record the answer in `application-answers.md`
- Time zone constraint: {{edit}}

## Compensation
- Floor: {{edit — e.g. $150k base}}
- If the posting lists no band, do not skip; note `comp: not listed`

## Seniority
- Target: {{edit — e.g. Mid → Senior}}
- Accept: {{edit — e.g. Staff if the JD's actual requirements match}}

---

## Fit scoring (used by `/job triage`, 0–100)

| Weight | Dimension |
|---|---|
| 30 | Role-family match (exact title in a list above = full marks) |
| 25 | Skill overlap with `jobs/input/profile/master-resume.md` |
| 15 | Seniority match |
| 10 | Location / work-mode match |
| 10 | Employer quality (see below) |
| 10 | Recency of posting |

**Employer quality — define it for your field.** Edit this line to say what a good employer
looks like to you, and the agent scores against it instead of guessing:

> {{e.g. "well-funded startups building a real product" · "Magnet-recognized hospitals" ·
> "firms with a named partner track" · "companies with public engineering blogs"}}

Bands: **80+ apply now** · **60–79 worth a look** · **40–59 stretch/backup** · **<40 skip**

## Volume

### Result limit: 10

**This is the main dial.** `/job search` stops once this many jobs are *verified* and on the
list. Change the number here and the next run fetches that many — no skill edit needed.

Because verification is the slow part (~20s per job), this number sets the runtime almost
entirely: 10 jobs ≈ 4–6 minutes, 25 ≈ 10–15, 50 ≈ 20–30.

Set it to `all` for an exhaustive run — sweep every enabled site and verify everything
found. Bounded by the time budget below rather than a count.

Override for a single run without editing this file:

```
/job search <role family> --limit 25     # this run only
/job search <role family> --limit all    # exhaustive; also accepts --no-limit
```

Supporting knobs:

- **Candidate pool multiplier: 3** — how many candidates to gather before verifying.
  At limit 10 the agent collects ~30 candidates, ranks them, then verifies the best ones
  until 10 pass. Raise it if too many candidates are dying in verification; lower it for
  speed.
- **Max applications queued at once: 10**

### Exhaustive run budget: 30 minutes

Only applies when the limit is `all`. A hard wall-clock stop, so an unlimited run always
terminates. On hitting it the agent writes everything verified so far and reports what was
left unswept — nothing is lost, and the next run resumes from the backlog.

Raise it for an overnight sweep (`Exhaustive run budget: 120 minutes`). There is no way to
set it to unbounded, by design.
