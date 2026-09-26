<!-- TEMPLATE for jobs/input/config/setup-families.md — written by `/job setup`, nothing else.
     It holds ONLY the role families the user confirmed during setup. When this file exists it
     is the authoritative family list for every task and for `make daily`; the family list in
     search-profile.md is ignored for searching and becomes the generic fallback.
     Everything else (country, keywords, hard filters, limits, fit scoring) stays in
     search-profile.md. Gitignored — it is personal. -->

# Setup families

- **Confirmed by the user:** {{YYYY-MM-DD}} via `/job setup`
- **Derived from:** `jobs/input/profile/master-resume.md`

To change what `make daily` searches, re-run `/job setup` or edit the families below. Delete a
`###` block to stop searching it.

## Families

### {{Family name — used verbatim in the Profile column}}
- Titles: {{exact title, synonyms, seniority variants}}
- Signals: {{JD-body phrases that confirm the role is really this work}}
- Evidence: {{which master-resume lines support it}}
