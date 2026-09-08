<!-- TEMPLATE for jobs/output/applications/<job-id>/log.md — /job apply writes and updates this. -->

# {{Role}} — {{Company}}

- **Job ID:** `{{job-id}}`
- **Status:** {{found|shortlisted|jd-captured|tailored|filled-awaiting-user|submitted|interviewing|offer|rejected|ghosted|skipped}}
- **JD:** `jobs/output/jds/{{job-id}}.md`  ·  captured via {{api|careers|apply-url|manual}}
- **Apply URL:** {{}}
- **ATS:** {{}}
- **Fit:** {{score}}/100
- **Found:** {{YYYY-MM-DD}} via {{source}}  ·  **Index row:** `jobs/output/jobs.md` → Summary
- **Resume sent:** `jobs/output/applications/{{job-id}}/resume.pdf`
- **Cover letter:** `jobs/output/applications/{{job-id}}/cover-letter.md` | none

## Timeline
| Date | Event |
|---|---|
| {{YYYY-MM-DD}} | Found via {{source}} |
| {{YYYY-MM-DD}} | JD captured ({{method}}) |
| {{YYYY-MM-DD}} | Resume tailored |
| {{YYYY-MM-DD}} | Form filled — awaiting user review |
| {{YYYY-MM-DD}} | **User submitted** |

## Form fields filled
| Field | Value used | Source |
|---|---|---|
| First name | {{from answer bank}} | answer bank |
| Resume | resume.pdf | jobs/output/applications/{{job-id}}/ |
| Work authorization | {{}} | answer bank |

## Questions the user had to answer
| Question | Answer given | Added to answer bank? |
|---|---|---|
| {{}} | {{}} | yes/no |

## Gaps flagged
Requirements from the JD the user does not meet, stated plainly. Not hidden, not fudged.
- {{}}

## Screenshots
- `jobs/output/applications/{{job-id}}/screens/01-form-top.png`
- `jobs/output/applications/{{job-id}}/screens/02-review.png`

## Follow-up
- Next action: {{e.g. "follow up if no reply by 2026-09-13"}}
- Contact / recruiter: {{}}

## Notes
{{anything the next run should know}}
