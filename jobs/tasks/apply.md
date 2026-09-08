# Task: `apply`

Run via `/job apply`. Formerly the standalone `job-apply` skill.

Open a job's application page in the browser and fill out every field of the application form — then stop before submitting and hand off to the user for review. Requires a tailored resume to already exist (run /job tailor first). Use when the user is ready to apply to a job, wants to start an application, or says "apply to <job>".

**Tool budget for this task:** Bash, Read, Write, Edit, Glob, Grep, AskUserQuestion

**Never spawns a subagent. Never runs for more than one job at a time. Never submits.** This is the only task allowed to fill a form.

Anything outside that budget means you are running the wrong task — return to
`SKILL.md` and route again rather than reaching for the tool.

---

Drive the browser through a real job application: fill every field, upload the files — and
stop at the review screen.

**This command does not tailor.** It is the last stage of the pipeline and expects
`jobs/output/applications/<job-id>/resume.pdf` to already exist:

```
/job jd <job-id>  →  /job tailor <job-id>  →  /job apply <job-id>
```

**Never run in parallel.** One browser, one form, one job at a time — and the user is in the
loop at the end of every run.

**Argument:** a job ID (`/job apply anthropic--forward-deployed-engineer`). With no argument,
show the shortlist from the tracker and ask which one.

**Two modes.** Attended (the default, Steps 0–7 below) is one job with the user watching.
**`--batch`** is unattended: it fills what it can across several jobs, never asks a question,
never waits, and leaves behind a replay script per job so the filled form survives the
browser session. See **Batch mode** at the end. Both modes obey THE RULE.

---

## THE RULE

**You do not submit. Ever.**

Fill everything. Upload everything. Clear every validation error. Then stop, screenshot the
completed form, and tell the user to review and click Submit themselves.

Never click: Submit, Submit Application, Send Application, Apply, Finish, Complete, Done, or
any button that advances past the final review. Never `press Enter` on a form field that
might submit — click a neutral element to blur instead. On a multi-step wizard, "Next" and
"Continue" between steps are fine; the final step's action button is not.

If you are unsure whether a button submits, **do not click it** — ask the user.

---

## Step 0 — Setup and preflight

```bash
_ROOT=$(git rev-parse --show-toplevel 2>/dev/null)
B=""; [ -n "$_ROOT" ] && [ -x "$_ROOT/.claude/skills/gstack/browse/dist/browse" ] && B="$_ROOT/.claude/skills/gstack/browse/dist/browse"
[ -z "$B" ] && B="$HOME/.claude/skills/gstack/browse/dist/browse"
[ -x "$B" ] && echo "READY: $B" || echo "NEEDS_SETUP"

grep -n '<job-id>' jobs/output/jobs.md jobs/output/tracker.md
cat jobs/output/jds/<job-id>.md                # the full JD — form notes live here too
cat jobs/output/applications/<job-id>/log.md 2>/dev/null
ls -l jobs/output/applications/<job-id>/resume.pdf
cat jobs/input/config/application-answers.md
```

Checks before touching the browser:

1. **Already applied?** Check the **Status** column for this job ID in `jobs/output/jobs.md`. If it reads
   `filled-awaiting-user` or `submitted`, stop and tell the user. Never apply twice.
   Set **Form** to `⏳ in-progress` in `jobs/output/jobs.md` when you begin, so an interrupted run
   stays visible on the dashboard.
2. **Resume ready?** If `jobs/output/applications/<job-id>/resume.pdf` doesn't exist, **stop**:
   `No tailored resume for <job-id>. Run /job tailor <job-id> first.` Do not tailor here,
   and never upload the master resume as a substitute — it is a working document, not a
   resume anyone should read.
3. **Answer bank ready?** Any `TODO` in the fields this form will need — ask now, in one
   batch, rather than interrupting mid-form.
4. **Posting still live?** `/job jd` already checked, but a req can close between
   stages. Take the **Apply URL** from `jobs/output/jds/<job-id>.md` (or `jobs/output/jobs.md`),
   `$B goto` it, and confirm. If it's gone, don't give up yet — open the **Careers** URL from
   the same row and look for the role under a new req ID; ATS links churn while the job stays
   open. Found it → update the Apply URL cell and continue. Genuinely gone → move the row to
   **Excluded** with verdict `dead`, mark the log `skipped — posting closed`, and stop.

## Step 1 — Read the posting one more time

```bash
$B goto "<apply-url>"
$B wait --networkidle
$B text
```

Diff what's on screen against `jobs/output/jds/<job-id>.md`. If the posting changed
materially since capture, stop and re-run `/job jd <job-id> --refresh` then
`/job tailor <job-id> --force` — don't fill a form with a resume aimed at an older
version of the job. Note anything the form asks that isn't in the answer bank.

## Step 2 — Map the form

```bash
$B snapshot -i                # interactive elements with @e refs
$B forms                      # every field as JSON — types, names, required flags
```

**Iframes:** Greenhouse and many embedded boards render the form in an iframe. If
`snapshot -i` shows no form fields, switch context:

```bash
$B frame --url "greenhouse"   # or: $B frame <selector>
$B snapshot -i
```

**Dynamic forms** (Ashby, Workday, custom React): `$B wait --networkidle` first, and
re-snapshot after every action — refs go stale when the DOM re-renders.

Build the field inventory before filling anything: for each field, the label, type, whether
it's required, and where its value comes from (profile / answer bank / ask the user / write
fresh). Show the user any field in the "ask" column and collect all of them at once.

## Step 3 — Handle authentication, if required

Many ATS platforms (Workday especially) require an account.

- **Existing session** → continue; browse state persists.
- **Login needed** → `$B handoff "<company> needs you to log in"`, tell the user what to do,
  wait for confirmation, then `$B resume`.
- **Account creation** → **ask the user first.** Never create an account, accept terms, or
  agree to a privacy policy on their behalf. If they approve, hand off so they set the
  password themselves.
- **CAPTCHA / bot wall / MFA** → `$B handoff` immediately. Do not attempt to solve it.

## Step 4 — Fill the form

Work top to bottom, section by section, verifying as you go.

```bash
$B fill @e3 "<first name>"                # identity fields come from the Identity & contact
$B fill @e4 "<last name>"                 # table in jobs/input/config/application-answers.md — never
$B fill @e5 "<email>"                     # typed from memory or hardcoded
$B select @e12 "Yes"                     # dropdowns: by value, label, or visible text
$B click @e18                            # radios and checkboxes
$B upload "#resume-upload" "jobs/output/applications/<job-id>/resume.pdf"
```

**File uploads.** Use the absolute path. Check the accepted types first — some boards reject
anything but PDF, a few want DOCX. Confirm the upload landed
(`$B is visible ".upload-success"`, or re-snapshot and look for the filename) before moving
on; a silently failed resume upload is the most common way an application goes out empty.

**Resume auto-parse.** Many forms parse the PDF and prefill fields. Always re-snapshot after
upload and **check every prefilled value** — parsers mangle dates, drop employers, and put
the phone number in the address field. Correct what's wrong.

**Custom / typeahead dropdowns** (not real `<select>`): click to open, type, re-snapshot,
click the matching option. Verify the selection stuck.

**Free-text questions** ("Why do you want to work here?", "Describe a project"): draft an
answer grounded in `jobs/output/applications/<job-id>/resume.md` and this company's actual product,
**show it to the user before filling it**, and respect the character limit. These answers
are the part a human actually reads. Same rule as the resume: nothing that isn't already
true in `jobs/input/profile/master-resume.md`.

**Screening questions:** answer only from `jobs/input/config/application-answers.md`. Anything else —
stop and ask, then append the new answer to the bank with a changelog line.

**EEO / self-identification:** optional. Use the user's recorded preference; where none is
recorded, choose the decline option. Never guess a demographic value.

**Multi-step wizards:** at each step, snapshot → fill → verify → screenshot → advance. Log
progress after each step so an interrupted run can resume.

Take screenshots as you go into `jobs/output/applications/<job-id>/screens/`.

## Step 5 — Verify before handing off

```bash
$B snapshot -i
$B screenshot "jobs/output/applications/<job-id>/screens/99-review.png"
```

Then read the screenshot back with `Read` so the user can see it.

Check: every required field filled · resume attached and correctly named · cover letter
attached if required · no validation errors showing · dropdowns show real selections, not
placeholders · nothing auto-parsed and left wrong · no field contains a fabricated value.

If a required field is still empty because you couldn't determine the answer, say so
explicitly rather than filling it with something plausible.

## Step 6 — Hand off to the user

Update `jobs/output/applications/<job-id>/log.md` (fields filled, questions asked, gaps flagged,
screenshot paths), then set **Status** to `filled-awaiting-user` and **Form** to
`✓ <date>` in **both** `jobs/output/jobs.md` and `jobs/output/tracker.md`. `jobs/output/jobs.md` is what the dashboard reads — a stale Status there means
the user's board is lying to them.

Then:

```
✅ Application filled — NOT submitted.

  Forward Deployed Engineer @ Anthropic
  https://job-boards.greenhouse.io/anthropic/jobs/...

  Filled:    23 fields
  Attached:  <First>-<Last>-FDE-Resume.pdf, cover-letter.pdf
  Answered:  work auth (Yes), sponsorship (No), travel 50% (Yes)
  Flagged:   JD wants Kubernetes — not claimed on your resume
  Review:    jobs/output/applications/anthropic--forward-deployed-engineer/screens/99-review.png

  👉 The browser is open on the completed form. Review it and click Submit yourself.
     Tell me when you've submitted and I'll update the tracker.
```

Leave the browser on the completed form. If the user wants a visible window to review in,
`$B handoff "Application ready for your review — please check and submit"`.

**Do not click submit even if the user says "go ahead and submit."** Explain that this
workspace is built so the final send is always theirs, and hand off the visible browser
instead.

## Step 7 — After the user confirms

When they say they submitted: set Status `submitted` in `jobs/output/jobs.md`, the log, and the tracker;
add a timeline entry; set a follow-up date (default +14 days); and offer the next job in the
queue. Never set `submitted` from your own action — only from the user's confirmation.

---

# Batch mode — `/job apply --batch [job-id ...]`

```
/job apply --batch                       # every job that is tailored and not yet filled
/job apply --batch tebra--senior-software-engineer logicgate--software-engineer-back-end
```

Unattended. It fills whatever it can source, records what it filled, and hands back a list.
**It still never submits.** Everything in THE RULE applies unchanged.

**Serial, never parallel.** One browser, one form at a time, jobs in order. `$B` is a single
shared page — two forms at once collide, and a half-filled form is worse than none.

## What changes from the attended flow

| Attended | `--batch` |
|---|---|
| Asks you any answer not in the bank | **Never asks.** Leaves the field blank, records it in `gaps`, moves on |
| Shows you free-text answers before typing them | Drafts them from the tailored resume, fills them, and flags each one for your review |
| Hands off and waits on login / CAPTCHA / account creation | **Skips the job** and records why. No waiting, ever |
| One job | Every tailored job, serially, each in its own tab |

## It runs in a visible browser, and leaves it open

**This is the point of the mode, not a detail.** A human has to click Submit, so the run has
to end with a real window on the user's screen showing a filled form. Headless would produce
a filled form nobody can see.

```bash
$B disconnect                       # the daemon refuses to mix headed and headless config
$B --headed goto "<apply-url>"      # visible Chromium; every later command needs --headed too
```

Browse runs the browser as a **long-lived daemon**: the window survives this run ending and
stays up until someone runs `$B disconnect`. That is what makes "fill it overnight, submit it
over coffee" work. Two consequences:

- **Every `$B` command in this mode carries `--headed`.** Without it the CLI tries to talk to
  a headless daemon, sees a config mismatch, and refuses.
- **Never `$B disconnect` at the end of a batch.** Closing the daemon throws away every
  filled form. Disconnect only at the *start*, and only to switch a headless daemon to headed.

One tab per job: `$B --headed newtab "<apply-url>"` returns the tab id; pass it to later
commands with `--tab-id <n>`. Report the tab number per job so the user knows which is which.

## Per job

1. **Preflight** exactly as Step 0. Skip the job — do not stop the batch — if: it is already
   `filled-awaiting-user`/`submitted`, `resume.pdf` is missing, or the posting is dead.
2. **Already mapped?** If `jobs/output/applications/<job-id>/form-fill.json` exists and the
   form still matches it, skip straight to step 5. Mapping is the expensive part and it only
   has to happen once per posting.
3. **Map the form** (Step 2): label, selector, `name`, type, required, and where the value
   comes from. Greenhouse, Ashby and Lever render their dropdowns as **react-select
   comboboxes, not `<select>`** — record those as `"type": "combobox"`, whose `value` is the
   option's visible text.
4. **Write the field map** to `jobs/output/applications/<job-id>/form-fill.json` —
   `jobs/input/templates/form-fill.json` is the shape. Every value comes from
   `jobs/input/profile/master-resume.md` or `jobs/input/config/application-answers.md`; a
   question with no answer on file goes in `gaps` and is **left blank, never guessed**. Every
   field needs its `label` verbatim: it is the fallback locator when a selector goes stale.
5. **Fill the form from the map:**

   ```bash
   jobs/bin/fill-form.py <job-id> [--tab N]
   ```

   It drives the headed browser: text fields, comboboxes (open → pick the option by text),
   checkboxes, radios, and a **real upload of `resume.pdf`** — the thing a pasted script
   cannot do. It clicks no submit button, and refuses an option that looks like one. It
   prints what it filled, what failed, and every deliberate blank.

6. **Verify and screenshot** — no validation errors showing, resume visibly attached — to
   `screens/99-review.png`.
7. **Generate the offline fallback:**

   ```bash
   jobs/bin/make-refill.py <job-id>
   ```

   `refill.js` refills the form from the console if the window is closed or the machine
   reboots before the user gets to it. It is the backup, not the handoff.
8. **Update state** as in Step 6: log, Status `filled-awaiting-user`, Form `✓ <date>` in both
   `jobs/output/jobs.md` and `jobs/output/tracker.md`. A skipped job keeps its current status
   and gets a log line saying why.

## Stop conditions

Stop the whole batch — do not keep going — if the same failure hits **three jobs in a row**
(a dead browser session, an expired login everywhere, a systematic selector failure). One
broken job is data; three in a row is a broken run, and finishing it just produces garbage
in eight folders.

## The report

```
Filled 5 of 7 — none submitted. The browser is open with 5 tabs.

  tab 1  ✓ tebra--senior-software-engineer        13 fields   11 blank
           blank: email · work auth · sponsorship · address · desired comp
  tab 2  ✓ logicgate--software-engineer-back-end  14 fields    0 blank
  ⊘ databricks--senior-software-engineer   skipped — Workday account required
  ⊘ palantir--backend-software-engineer    skipped — CAPTCHA on the apply page

  👉 Switch to the browser, check each tab, submit the ones you're happy with.
     Closed the window before you got to it? Re-open the job and run
     jobs/bin/fill-form.py <job-id> — it refills in seconds.
```

Keep it to the table plus the blank-field lines. The detail is in each job's `log.md`.

## The two ways a filled form survives

A filled form lives in a browser tab, not in a URL — open the Apply URL again and you get an
empty form. So the run leaves two things behind:

1. **The open tab** (primary). Real window, resume attached, ready to submit.
2. **`refill.js` / `fill-form.py`** (fallback), for when the window is gone. `fill-form.py`
   is the one to prefer — it re-opens the posting and refills it in the visible browser,
   attachment included. `refill.js` is for refilling a page in a browser this workspace is
   not driving: paste it into the console, then attach the resume by hand.

Neither can submit. `refill.js` clicks no buttons at all; `fill-form.py` clicks only form
controls and dropdown options named in `form-fill.json`, and refuses any that look like a
submit control.
