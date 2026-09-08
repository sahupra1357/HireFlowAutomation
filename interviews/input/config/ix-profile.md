# Interview Intel Profile

What counts as an interview experience worth compiling. `/job interviews` reads this
file. Edit freely — this is the main dial for tuning signal vs. junk.

---

## Run limits

| Setting | Value | Meaning |
|---|---|---|
| **Result limit** | `10` | Compiled experiences per run. `all` = exhaustive. |
| **Candidate pool multiplier** | `3` | Collect ~`limit x 3` raw posts to keep `limit`. |
| **Exhaustive run budget** | `30 minutes` | Hard stop for `--limit all`. |
| **Per-source cap** | `5` | No single subreddit or site dominates a run. |
| **Recency window** | `18 months` | Older than this is a hard drop — stale loops mislead. |
| **Min credibility** | `40` | Below this goes to Excluded, not the Reports table. |
| **Track filter** | `both` | `a` \| `b` \| `both` |

Command-line overrides apply to one run only and are never written back here:

```
/job interviews                          # limit 10, both tracks, all enabled sources
/job interviews --limit 25
/job interviews --limit all              # exhaustive; also --no-limit, --limit 0
/job interviews anthropic                # company-targeted (track A)
/job interviews --track b                # only AI-screener experiences
/job interviews --source reddit,hn       # subset of enabled sources
/job interviews --since 6m               # tighter recency window
```

---

## Track A — interviewing **at** a company

Seed the company list from `jobs/output/jobs.md` (read-only) so this stays bounded to companies
actually in the pipeline. A company-targeted run overrides the seed.

**Query shapes:**
```
"<company>" interview experience
"<company>" interview loop OR onsite OR "final round"
"<company>" "Forward Deployed" OR "Applied AI" interview
```

**What makes an A-track post valuable:** it names the rounds, what was actually asked, the
format, and the outcome. A post that only says "I bombed it, they ghosted me" is venting,
not intel.

## Track B — interviewed **by** AI

Company-agnostic. This is the playbook for AI-run screens, which now front-load the exact
startups in `jobs/output/jobs.md`.

**Query shapes:**
```
"AI interview" experience
HireVue OR Ribbon OR Mercor OR "AI recruiter" interview experience
"AI screening" OR "async video interview" OR "AI interviewer" experience
```

**Vendors to recognize:** HireVue, Ribbon, Mercor, Karat, CodeSignal, HackerRank AI,
Sapia, Paradox/Olivia, Micro1, Alex, Apriora. New vendors appear constantly — record any
new name in the entry's Details block so this list can grow.

---

## Credibility score (0–100)

Reddit RSS carries no score or comment count, so engagement is not a scoreable dimension.
Firsthand-ness is scored instead — it is a better proxy for truth anyway.

| Dimension | Weight | What earns it |
|---|---|---|
| **Specificity** | 30 | Names actual rounds, questions asked, format, timeline. Zero if it is all vibes. |
| **Recency** | 25 | Full marks under 6 months, half at 12, zero at 18+. |
| **Corroboration** | 20 | An independent report agrees. One source = 0 here, by definition. |
| **Firsthand** | 15 | First person, past tense, "I interviewed". Secondhand, news-about, or speculation scores low. |
| **Source quality** | 10 | Long-form blog > HN > topical subreddit > general subreddit. |

**Bands:** `80+` trust it · `60–79` useful with caveats · `40–59` weak, listed only ·
`<40` Excluded.

---

## Junk filter — hard drops, never surfaced

These are what turn this from intel into a scrapbook. Drop on sight:

- **Self-promotion.** Someone shipping an interview-prep tool, bootcamp, course, resume
  service, or "AI interview coach". Very common in the exact subreddits worth searching.
- **Recruiter / vendor marketing** dressed as a candidate story.
- **News-about, not experience-of.** An article *about* AI interviewing is not an
  interview experience. Track B needs someone who sat through one.
- **Question dumps** with no experience attached — "top 50 LLM interview questions".
- **Pure venting** with no round, question, format, or timeline recoverable.
- **Advice threads** where nobody reports an actual interview.
- **Karma-farm rage bait** — no company, no specifics, maximum outrage.
- Anything outside the recency window.

Every hard drop is counted in the run header with its reason, so a thin run never reads
as a broken one.
