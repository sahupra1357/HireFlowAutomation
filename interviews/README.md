# interviews/ — interview-experience intel

Separate namespace from `jobs/`. `/job interviews` reads `jobs/output/jobs.md` for
company names and never writes to it.

| | |
|---|---|
| `input/config/` | `ix-profile.md` (limits, scoring, junk filter), `ix-sources.md` (where to search) |
| `input/templates/` | `ix-results.md` — section shapes |
| `output/interviews.md` | The compiled index |
| `output/reports/<company-slug>/` | Per-company prep briefs |
| `bin/ix-fetch.py` | Reddit / HN fetcher |
