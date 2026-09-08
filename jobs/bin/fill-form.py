#!/usr/bin/env python3
"""Drive a visible browser through one job's application form.

    jobs/bin/fill-form.py <job-id> [--tab N] [--no-upload]

Reads  jobs/output/applications/<job-id>/form-fill.json
Drives the headed browse daemon ($B --headed) to fill every field in it
Leaves the tab open, filled, on screen — the human reviews and submits

This is the mechanical half of `/job apply --batch`. Working out *what* goes in each field
is the agent's job, done once per posting and written into form-fill.json. Putting it back
on screen is this script's job, and it is the same every morning — so a job that has been
mapped once refills in seconds without the agent thinking about it again.

It fills. It uploads. It never clicks a submit button: the only clicks it makes are on form
controls and dropdown options named in form-fill.json.

Stdlib only. Exit 0 if every field landed, 1 if any did not.
"""

import json
import os
import subprocess
import sys
import time

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
APPS = os.path.join(ROOT, "jobs", "output", "applications")

# Anything whose text looks like this is a submit control. The script never clicks one, but
# the guard is here so a bad form-fill.json cannot turn this into a submitter.
FORBIDDEN = ("submit", "send application", "apply now", "finish", "complete application")


def browse_bin():
    for c in (
        os.path.join(ROOT, ".claude", "skills", "gstack", "browse", "dist", "browse"),
        os.path.expanduser("~/.claude/skills/gstack/browse/dist/browse"),
    ):
        if os.access(c, os.X_OK):
            return c
    sys.exit("browse not found — install the gstack browse skill")


class B:
    """The headed browse daemon. Every call carries --headed: the daemon refuses to mix
    headed and headless config, and headed is the whole point here."""

    def __init__(self, tab=None):
        self.bin = browse_bin()
        self.tab = tab

    def run(self, *args, timeout=60):
        cmd = [self.bin, "--headed", *[str(a) for a in args]]
        if self.tab is not None and args[0] in ("goto", "click", "fill", "upload", "js", "screenshot"):
            cmd += ["--tab-id", str(self.tab)]
        p = subprocess.run(cmd, capture_output=True, text=True, timeout=timeout)
        return p.returncode, (p.stdout or "").strip(), (p.stderr or "").strip()

    def js(self, expr):
        _, out, _ = self.run("js", expr)
        return out.splitlines()[-1].strip() if out else ""


def fill_field(b, f):
    """Return (ok, note). One field, by type."""
    label = f.get("label", "?")
    sel = f.get("selector")
    typ = (f.get("type") or "text").lower()
    val = f.get("value")

    if typ == "file":
        return None, "file — handled separately"

    if not sel:
        return False, "no selector recorded"

    if typ in ("text", "email", "tel", "textarea", "number", "url"):
        rc, out, err = b.run("fill", sel, str(val))
        return rc == 0, (err or out)[:120]

    if typ == "checkbox":
        state = b.js("var e=document.querySelector(%s); e?e.checked:null" % json.dumps(sel))
        want = val in (True, "true", "yes", "Yes", "on")
        if state == str(want).lower():
            return True, "already set"
        rc, out, err = b.run("click", sel)
        return rc == 0, (err or out)[:120]

    if typ == "radio":
        rc, out, err = b.run("click", sel)
        return rc == 0, (err or out)[:120]

    if typ == "combobox":
        # react-select and friends: open the menu, then click the option by its text. The
        # menu renders a frame later, hence the pauses.
        rc, _, err = b.run("click", sel)
        if rc != 0:
            return False, "could not open the dropdown: %s" % err[:90]
        time.sleep(0.8)
        want = json.dumps(str(val))
        oid = b.js(
            "var o=[].filter.call(document.querySelectorAll('[id^=\"react-select\"][id*=\"option\"],[role=\"option\"]'),"
            "function(e){return e.textContent.trim()===%s})[0]; o?o.id:''" % want
        )
        if not oid or oid in ("undefined", "null", ""):
            b.run("press", "Escape")
            return False, "no option %r in the dropdown" % val
        if any(w in oid.lower() for w in FORBIDDEN):
            b.run("press", "Escape")
            return False, "refused: option id looks like a submit control"
        rc, out, err = b.run("click", "#" + oid)
        time.sleep(0.3)
        return rc == 0, (err or out)[:120]

    return False, "unknown field type %r" % typ


def main(argv):
    if not argv or argv[0].startswith("-"):
        sys.exit(__doc__)
    job_id = argv[0]
    tab = None
    if "--tab" in argv:
        tab = argv[argv.index("--tab") + 1]
    do_upload = "--no-upload" not in argv

    d = os.path.join(APPS, job_id)
    src = os.path.join(d, "form-fill.json")
    if not os.path.exists(src):
        sys.exit("no field map: %s\nRun the agent's mapping pass first (/job apply --batch %s)." % (src, job_id))
    with open(src, encoding="utf-8") as fh:
        data = json.load(fh)

    b = B(tab)
    url = data.get("apply_url")
    if not url:
        sys.exit("form-fill.json has no apply_url")

    here = b.js("location.href")
    if url.split("?")[0] not in here:
        print("  → %s" % url)
        rc, out, err = b.run("goto", url, timeout=120)
        blob = (out + err)
        if rc != 0 or "No active page" in blob or "Error" in blob[:40]:
            print("  ✗ could not open the posting: %s" % (blob.splitlines()[0] if blob else "browse failed"))
            print("     the browser session looks dead — `browse disconnect`, then run this again")
            return 1
        b.run("wait", "--networkidle", timeout=60)
        time.sleep(1)

    filled, failed = [], []
    for f in data.get("fields", []):
        ok, note = fill_field(b, f)
        if ok is None:
            continue
        (filled if ok else failed).append((f.get("label", "?"), note))
        print("  %s %s%s" % ("✓" if ok else "✗", f.get("label", "?")[:70], "" if ok else "  — " + note))

    # The resume: a real upload, because this is Playwright driving a real browser — the
    # thing the paste-in-console fallback cannot do.
    up = None
    if do_upload:
        for f in data.get("fields", []):
            if (f.get("type") or "") == "file":
                path = os.path.join(d, f.get("value") or "resume.pdf")
                name = os.path.basename(path)
                if not os.path.exists(path):
                    print("  ✗ %s — %s not on disk" % (f.get("label"), path))
                    failed.append((f.get("label", "resume"), "file missing"))
                    continue
                # Already attached? Re-uploading is not just wasted time: once a file is
                # attached the input is swapped for a Replace/Remove control and the upload
                # blocks until it times out. A morning re-run must sail past this.
                shown = b.js(
                    "var f=document.querySelector('form')||document.body;"
                    "f.textContent.indexOf(%s)>=0" % json.dumps(name)
                )
                if shown == "true":
                    print("  ✓ %s → %s (already attached)" % (f.get("label", "Resume"), name))
                    up = True
                    continue
                rc, out, err = b.run("upload", f.get("selector") or "input[type=file]", path, timeout=45)
                up = rc == 0
                print("  %s %s → %s" % ("✓" if up else "✗", f.get("label", "Resume"), name))
                if not up:
                    failed.append((f.get("label", "resume"), (err or out)[:120] or "upload timed out"))

    gaps = data.get("gaps") or []
    print("\n  filled %d · failed %d · left blank on purpose %d" % (len(filled), len(failed), len(gaps)))
    for g in gaps:
        print("     ▢ %s — %s" % (g.get("label", "?"), g.get("why", "")))
    print("\n  Tab is open and filled. NOT submitted — that click is yours.")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
