# Interview experience index — section shapes

`/job interviews` writes `interviews/output/interviews.md` using exactly these shapes.
Read this before writing. The Go dashboard parses `## Reports`, `## Unverified backlog`,
and `## Excluded` as pipe tables, so column names must not drift.

---

## Reports (main table)

| # | Cred | Track | Company | Role / Topic | Interviewed | Rounds | Format | Outcome | Source | Link | Job ID |
|---|------|-------|---------|--------------|-------------|--------|--------|---------|--------|------|--------|
| 1 | 82 | a-company | Anthropic | Forward Deployed Engineer | 2026-05 | 5 | onsite + take-home | offer | reddit | https://... | `anthropic--forward-deployed-engineer` |
| 2 | 64 | b-ai-screener | HireVue | async video AI screen | 2026-07 | 1 | AI video | rejected | hn | https://... | — |

- **Cred** — 0–100 from the weights in `ix-profile.md`.
- **Track** — `a-company` or `b-ai-screener`. Nothing else.
- **Interviewed** — `YYYY-MM` when the interview happened, not when the post was written.
  If only the post date is known, use it and mark `posted-date-only` in Details.
- **Rounds** — an integer, or `—` if not stated. Never guess.
- **Outcome** — `offer` · `rejected` · `ghosted` · `withdrew` · `in-progress` · `—`.
- **Job ID** — the matching row in `jobs/output/jobs.md`, or `—`. This is the join that makes the
  index useful. Never invent one; it must already exist in `jobs/output/jobs.md`.

## Unverified backlog

Found and ranked, but not yet read in full. The next run reads from here before searching.

| Provisional | Track | Company | Title | Posted | Source | Link |
|---|---|---|---|---|---|---|

## Excluded

| Reason | Track | Title | Source | Link |
|---|---|---|---|---|

Reason is one of the junk-filter labels: `self-promo` · `vendor-marketing` ·
`news-not-experience` · `question-dump` · `venting` · `advice-thread` · `ragebait` ·
`too-old` · `low-credibility` · `duplicate`.

## Details

One block per entry in Reports. This is where the value actually lives — the table is
just an index into it.

```
### `ix--anthropic--fde-loop--reddit--7f3a91c2`

- **Track:** a-company · **Credibility:** 82 (`corroborated` by ix--anthropic--fde-loop--hn--2b8e40da)
- **Company:** Anthropic · **Role:** Forward Deployed Engineer
- **Interviewed:** 2026-05 · **Posted:** 2026-06-02 · **Source:** reddit r/ExperiencedDevs
- **Link:** https://...
- **Outcome:** offer

**The loop, as reported**
1. Recruiter screen, 30 min — background, why Anthropic.
2. ...

**Questions reported** (verbatim where quoted, <=40 words each)
- "..." — [source](https://...)

**Signal quality:** firsthand, names rounds and questions, single-source on the
take-home detail.

**Caveats:** one report only; the take-home step is not corroborated anywhere else.
```

**Rules that bind every Details block**

- Verbatim quotes only, <=40 words, each with its URL. Never paraphrase a question into
  something crisper than what was written.
- No usernames, no PII about the poster. The URL is the entire attribution.
- A claim appearing in exactly one report is labelled `single-source` and stays labelled.
  Never promote it to "Company X asks Y".
- Never fabricate a round, a question, a difficulty, or an outcome. Absent means `—`.
