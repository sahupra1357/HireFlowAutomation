# jobs/input/profile/ — your real data lives here

Drop your existing resume (PDF, DOCX, or Markdown) into `source-resumes/`, then run
`/job setup`. The agent reads it and populates `master-resume.md`, `links.md`, and
`jobs/input/config/application-answers.md` — creating them from `jobs/input/templates/` if they don't exist yet.

Nothing personal is hardcoded in this repo. Every name, email, and fact the agent uses
traces back to the file you drop in here.

- `master-resume.md` — source of truth; tailored resumes are subsets of it
- `links.md` — LinkedIn/GitHub/portfolio, plus references
- `source-resumes/` — your original files, kept as-is (these get uploaded to applications)

Nothing in this directory is sent anywhere except an employer's own application form.
