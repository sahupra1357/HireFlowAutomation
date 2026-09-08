<!-- TEMPLATE for jobs/output/jds/<job-id>.md — /job jd writes this. -->
<!-- One file per job. The full JD lives here; jobs/output/jobs.md keeps only the short summary. -->

# {{Role}} — {{Company}}

- **Job ID:** `{{job-id}}`
- **Captured:** {{YYYY-MM-DD}}
- **Capture method:** `api` | `careers` | `apply-url` | `manual` | `manual-required`
- **Liveness:** `active` | `inconclusive` | `dead`
  <!-- `active` ONLY when the role was found on the employer's own careers page or the ATS
       API's live board index. A JD read off the apply URL alone is `inconclusive` — the
       posting rendered, but nothing proved it is still on the board. Never upgrade. -->
- **Proved by:** {{the exact URL that established liveness, or "not established"}}
- **Apply URL:** {{}}
- **Careers URL:** {{}}
- **ATS:** greenhouse | lever | ashby | workday | recruitee | workable | other
- **Location / mode:** {{}}
- **Compensation:** {{as stated on the posting, or "not listed"}}
- **Posted / deadline:** {{if stated}}

## Full job description

<!-- Verbatim. Do not summarize, do not paraphrase, do not drop sections. This is the text
     /job tailor reads. Keep the employer's own headings. -->

{{full JD text}}

## Requirements — extracted

- **Must-have:**
  - {{}}
- **Nice-to-have:**
  - {{}}
- **Responsibilities:**
  - {{}}

## ATS keywords

{{exact terms an automated screen will look for, comma-separated, in the JD's own spelling}}

## Application form — observed

- Account required: yes | no | unknown
- Cover letter: required | optional | not offered
- Resume format accepted: {{PDF / DOCX / …}}
- Custom questions seen: {{list, or "not inspected"}}
- Notes: {{iframes, multi-step wizard, login wall, anything /job apply should expect}}

## Capture notes

{{What was tried and what happened — which ladder rungs ran, what failed and why.
  On `manual-required`, this section states exactly what the user needs to do.}}
