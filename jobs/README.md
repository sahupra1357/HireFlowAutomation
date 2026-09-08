# jobs/ — profile setup → search → tailor → apply

| | |
|---|---|
| `input/` | What **you** provide. Drop your resume in `input/profile/source-resumes/`, tune `input/config/`. |
| `output/` | What **the agent** produces: `jobs.md` (index), `tracker.md`, `jds/`, `applications/<job-id>/`. |
| `dashboard/` | Go web dashboard over `output/jobs.md` — `make web`. |

Skills read from `input/` and write to `output/`. The only file you place under `output/` is a
hand-downloaded posting at `output/jds/<job-id>-source.*`, next to the JD it becomes.
