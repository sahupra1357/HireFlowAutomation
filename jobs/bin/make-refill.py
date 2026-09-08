#!/usr/bin/env python3
"""Generate a job's replay script from the field map /job apply --batch recorded.

    jobs/bin/make-refill.py <job-id> [<job-id> ...]
    jobs/bin/make-refill.py --all

Reads  jobs/output/applications/<job-id>/form-fill.json
Writes jobs/output/applications/<job-id>/refill.js          (paste into the page console)
       jobs/output/applications/<job-id>/refill-bookmarklet.txt

The batch run fills the form in a browser session that will not survive the day. These two
files are what does: they put the same answers back into the same form, in any browser, at
any time — so "the agent filled it" does not depend on a tab staying open.

Neither file submits anything, and neither can attach the resume: browsers forbid setting a
file input from script. Both say so on the page.

Stdlib only, like everything else in jobs/bin/.
"""

import json
import os
import sys
import urllib.parse

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
APPS = os.path.join(ROOT, "jobs", "output", "applications")
ENGINE = os.path.join(ROOT, "jobs", "bin", "refill-engine.js")
TOKEN = "/*__DATA__*/ {}"

HEADER = """// {role}{at}
// Replay script for {job_id} — generated {captured} by /job apply --batch.
//
// HOW TO USE
//   1. Open the application form:
//      {apply_url}
//   2. DevTools → Console (⌥⌘J on macOS, F12 on Windows). Chrome asks you to type
//      "allow pasting" the first time you paste into a console — type it, once.
//   3. Paste this whole file, press Enter.{frame_note}
//   4. Attach {resume_file} yourself — a script is not allowed to fill a file input.
//   5. Read every field, fix anything the form re-rendered, then submit it yourself.
//
// It fills {n_fields} fields and clicks no buttons. It cannot submit.
{gap_note}
"""


def load_engine():
    with open(ENGINE, encoding="utf-8") as fh:
        src = fh.read()
    if TOKEN not in src:
        sys.exit("refill-engine.js has no %s slot — did it get edited?" % TOKEN)
    return src


def build(job_id, engine):
    d = os.path.join(APPS, job_id)
    src = os.path.join(d, "form-fill.json")
    if not os.path.exists(src):
        return "no form-fill.json — run /job apply --batch %s first" % job_id
    with open(src, encoding="utf-8") as fh:
        try:
            data = json.load(fh)
        except json.JSONDecodeError as e:
            return "form-fill.json is not valid JSON: %s" % e

    fields = data.get("fields") or []
    if not fields:
        return "form-fill.json lists no fields"
    for i, f in enumerate(fields):
        if not isinstance(f, dict) or "label" not in f:
            return "field %d has no label — every field needs one, it is the fallback locator" % i
        if f.get("type") != "file" and "value" not in f:
            return "field %r has no value" % f.get("label")

    gaps = data.get("gaps") or []
    resume_file = data.get("resume_file") or "your tailored resume PDF"
    data.setdefault("resume_file", resume_file)

    frame_note = ""
    if data.get("frame"):
        frame_note = (
            "\n//      This form lives in an iframe (%s): in the console's frame dropdown\n"
            "//      (top-left, next to the filter box) pick that frame first, or the script\n"
            "//      will not see the fields." % data["frame"]
        )
    gap_note = ""
    if gaps:
        gap_note = "//\n// STILL NEEDS YOU — no answer on file, left blank on purpose:\n"
        for g in gaps:
            gap_note += "//   · %s — %s\n" % (g.get("label", "?"), g.get("why", "you decide"))

    head = HEADER.format(
        role=data.get("role", "Application"),
        at=(" @ " + data["company"]) if data.get("company") else "",
        job_id=job_id,
        captured=data.get("captured", "—"),
        apply_url=data.get("apply_url", "(no apply URL recorded)"),
        frame_note=frame_note,
        resume_file=resume_file,
        n_fields=len([f for f in fields if f.get("type") != "file"]),
        gap_note=gap_note,
    )

    body = engine.replace(TOKEN, json.dumps(data, indent=2, ensure_ascii=False))
    js = head + "\n" + body

    with open(os.path.join(d, "refill.js"), "w", encoding="utf-8") as fh:
        fh.write(js)

    # The bookmarklet is the convenience path: one click instead of a console paste. Some
    # boards' Content-Security-Policy blocks javascript: URLs outright, which is why the
    # console file above stays the primary route.
    mark = "javascript:" + urllib.parse.quote(
        "(function(){%s})();" % body.strip().rstrip(";"), safe="",
    )
    with open(os.path.join(d, "refill-bookmarklet.txt"), "w", encoding="utf-8") as fh:
        fh.write(
            "Drag-to-bookmark link for %s.\n"
            "Make a new bookmark, paste the line below as its URL, open the form, click it.\n"
            "If nothing happens the site's CSP blocked it — use refill.js in the console.\n\n%s\n"
            % (job_id, mark)
        )
    return None


def main(argv):
    if not argv or argv[0] in ("-h", "--help"):
        sys.exit(__doc__)
    engine = load_engine()
    ids = argv
    if argv[0] == "--all":
        ids = sorted(
            n for n in os.listdir(APPS)
            if os.path.exists(os.path.join(APPS, n, "form-fill.json"))
        )
        if not ids:
            sys.exit("no job has a form-fill.json yet")
    bad = 0
    for job_id in ids:
        err = build(job_id, engine)
        if err:
            print("  ✗ %s — %s" % (job_id, err))
            bad += 1
        else:
            print("  ✓ %s → jobs/output/applications/%s/refill.js" % (job_id, job_id))
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
