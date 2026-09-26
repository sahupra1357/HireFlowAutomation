<!-- TEMPLATE for jobs/input/inbox.md — `make init` (or the first `/job add --inbox`) copies
     this there. The live copy is gitignored: it is your list, nobody else's. -->

# Job inbox

Jobs **you** found — a link a friend sent, a posting you saw on a site the search doesn't
cover, a company you want to work for. Paste one link per line under **Add**, whenever you
like. `/job add --inbox` (and every bare `/job` run) picks up each one that isn't in
`jobs/output/jobs.md` yet, verifies it, captures its JD, and puts it on the list as
`shortlisted` — you picked it, so it skips triage.

The agent **never edits this file.** A link already in the index is recognised and skipped,
so you can leave processed lines here or delete them — either is fine. A link it could not
read (login wall, dead page) is reported at the end of the run with what to do.

Optional after the link, separated by `|`: company, role, a note. Useful when the page
can't be read without a login (LinkedIn), so the row still gets a sensible name.

## Add

- https://job-boards.greenhouse.io/<company>/jobs/<id>
- https://jobs.lever.co/<company>/<posting-id> | Company | Role title | referred by a friend

<!-- Delete the two example lines above; they are skipped because they contain <...>. -->

## Pasted job descriptions

No link, only the text? Paste it into Claude as `/job add` followed by the text, and say the
company and role. It is saved verbatim as that job's JD.
