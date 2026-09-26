<!-- TEMPLATE for jobs/output/applications/<job-id>/evaluation.md — /job evaluate writes this.
     One per job. Advisory: it never changes Status by itself. /job tailor reads section E. -->

# Evaluation — {{Role}} — {{Company}}

- **Job ID:** `{{job-id}}`
- **Evaluated:** {{YYYY-MM-DD}}
- **Verdict:** `apply` | `apply-with-caveats` | `skip`
- **Fit:** {{0–100, re-scored against the full JD}} (index had {{N}})
- **Role family:** {{from the active family list, or Adjacent}}
- **JD:** `jobs/output/jds/{{job-id}}.md` ({{capture method}}, liveness {{active | inconclusive}})
- **Apply URL:** {{}}

> {{One-paragraph bottom line: should the user spend an application on this, and the one
> reason that decides it.}}

---

## A) Role summary

| | |
|---|---|
| Role family | {{}} |
| Function | {{build · operate · advise · manage · sell · research — whatever the JD actually asks for}} |
| Seniority (JD) | {{as stated or implied, with the words that imply it}} |
| Work mode | {{remote · hybrid (N days) · onsite}} — {{location}} |
| Team / reports to | {{if stated, else —}} |
| Comp (posted) | {{as stated, else "not listed"}} |
| TL;DR | {{one sentence: what this person does all day}} |

## B) Requirement match

Every requirement from the JD's *Requirements — extracted* section, matched to the line in
`master-resume.md` that evidences it. Quote the master; never paraphrase it upward.

| # | JD requirement | Must/Nice | Evidence in master-resume.md | Match |
|---|---|---|---|---|
| 1 | {{}} | must | "{{exact master line}}" — {{section / role}} | ✓ strong · ~ partial · ✗ none |

**Coverage:** {{N}}/{{M}} must-haves evidenced · {{N}}/{{M}} nice-to-haves

### Gaps

One entry per `~` or `✗` above. A gap is reported, never papered over.

| Gap | Blocker? | Adjacent evidence that is really there | Honest mitigation |
|---|---|---|---|
| {{}} | hard blocker · likely screen-out · nice-to-have | "{{master line}}" or — | {{what the cover letter or an interview answer can truthfully say; or "none — leave it out"}} |

## C) Level & positioning

- **JD level vs the user's level:** {{e.g. JD wants Staff; master shows 4 yrs as Senior with
  cross-team scope in <role>}}
- **Lead with:** {{the 2–3 master-resume facts that make the strongest truthful case}}
- **Don't claim:** {{anything the JD rewards that the master does not support — so tailoring
  and form answers stay honest}}
- **If they level you lower:** {{is the role still worth it at the next level down — comp,
  scope, path back up}}

## D) Comp & demand

| Source | Figure | For | Date |
|---|---|---|---|
| Posting | {{as listed}} | this req | {{}} |
| {{Levels.fyi / Glassdoor / Blind / other public source}} | {{}} | {{title, level, location}} | {{}} |

- **vs the user's floor** (`search-profile.md` → Compensation): {{above · at · below · unknown}}
- **Read:** {{one or two lines. If no public data was found, say "no data found" — never
  estimate a number.}}

## E) Tailoring plan

What `/job tailor` should do for this job. Every row re-weights something already in the
master; none adds anything.

| # | Resume section | Change | Why (JD line) |
|---|---|---|---|
| 1 | Headline / Summary | {{}} | {{}} |
| 2 | {{role}} bullets | {{promote / reorder / rephrase in the JD's vocabulary}} | {{}} |

**Keywords to hit (only where the master truly supports them):** {{comma-separated}}
**Keywords to leave out (not supported):** {{comma-separated}}

## F) Interview prep

3–6 stories built **only** from master-resume.md, each mapped to what this JD will probe.
Situation, task, action, result and a reflection. A part the master doesn't record is `—`,
for the user to fill in — never invented.

| # | JD requirement it answers | Story (master-resume source) | S | T | A | R | Reflection |
|---|---|---|---|---|---|---|---|
| 1 | {{}} | {{role · bullet}} | {{}} | {{}} | {{}} | {{}} | {{— or the user's own words}} |

**Questions they'll likely ask about the gaps**, and the truthful answer:
- {{question}} → {{answer grounded in the master, or "acknowledge; say what you'd do"}}

## G) Verdict

- **Verdict:** `apply` | `apply-with-caveats` | `skip`
- **Why:** {{the deciding reasons, including any hard blocker from B}}
- **Before applying:** {{e.g. answer the sponsorship question in the answer bank; ask a
  referral; confirm the location}}
- **Next:** `/job tailor {{job-id}}` · or leave it — `skip` is advice, not a status change
