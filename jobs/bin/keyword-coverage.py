#!/usr/bin/env python3
"""Keyword coverage — how many of a JD's ATS keywords the tailored resume actually carries.

    jobs/bin/keyword-coverage.py <job-id> [<job-id> ...]
    jobs/bin/keyword-coverage.py --all          # every job with a tailored resume.md
    make keywords [JOB=<job-id>]

Reads the keyword list from the JD's `## ATS keywords` section
(jobs/output/jds/<job-id>.md), the tailored resume (applications/<job-id>/resume.md, its
tailoring receipt stripped), and the master resume. Every keyword lands in one bucket:

  hit           in the tailored resume
  missed        not in the resume, but the master resume has it — a tailoring miss worth
                a look (the one bucket tailoring can still act on)
  not in master not in the resume and not in the master either — correctly left out.
                Adding it would be fabrication; this is a gap, not a miss.

Separately, `rephrased` lists hits whose wording is in the resume but nowhere in the master —
usually the JD's vocabulary for something the master says differently, which tailoring is
allowed to do. Each one is worth a glance to confirm it is a fair rephrase, not a claim.

Coverage is hit / all keywords, the figure ATS screens care about. `supportable` is
hit / (hit + missed): how much of what the user could truthfully claim made it in. A low
coverage with a high supportable figure means the job asks for things the user doesn't
have, not that the tailoring was lazy.

Writes applications/<job-id>/keywords.json (read by the dashboard and check-index.py) and
prints one line per job. Deterministic, no model call. Never edits resume.md or jobs.md.
"""
import datetime
import json
import os
import re
import sys
from pathlib import Path

ROOT = Path(os.environ.get("HIREFLOW_ROOT") or Path(__file__).resolve().parents[2])
OUT = ROOT / "jobs" / "output"
MASTER = ROOT / "jobs" / "input" / "profile" / "master-resume.md"
COMMENT = re.compile(r"<!--.*?-->", re.S)


def keywords(jd_text):
    m = re.search(r"^## ATS keywords\s*$(.*?)(?=^## |\Z)", jd_text, re.M | re.S)
    if not m:
        return []
    body = COMMENT.sub("", m.group(1))
    raw = re.split(r"[,;\n]|\s·\s", body)
    seen, out = set(), []
    for k in raw:
        k = re.sub(r"^[\s*•-]+|[\s.]+$", "", k).strip("` ")
        key = k.lower()
        if k and "{{" not in k and len(k) <= 60 and key not in seen:
            seen.add(key)
            out.append(k)
    return out


def pattern(kw):
    """Whole-word, tolerant of the variants an ATS also accepts: hyphen/space/slash
    interchangeable, a trailing plural 's' either way. Case-insensitive — except a keyword
    of three letters or fewer, which must appear as written, capitalised or upper-case,
    so `Go`, `ML` or `Ray` never match ordinary prose ("to go live", "ray of light")."""
    k = kw.strip()
    if len(k) <= 3 and k.isalpha():
        # the JD's own spelling only if it isn't all lower-case ("go" is a verb; "Go" isn't)
        forms = {k.upper(), k.capitalize()} | ({k} if not k.islower() else set())
        alts = "|".join(map(re.escape, sorted(forms)))
        return re.compile(rf"(?<![A-Za-z0-9])(?:{alts})s?(?![A-Za-z0-9])")
    words = re.split(r"[\s\-/]+", k)
    words[-1] = re.sub(r"(?i)s$", "", words[-1]) if len(words[-1]) > 3 else words[-1]
    body = r"[\s\-/]*".join(re.escape(w) for w in words if w)
    return re.compile(rf"(?<![A-Za-z0-9]){body}(?:e?s)?(?![A-Za-z0-9])", re.I)


def master_text():
    t = COMMENT.sub("", MASTER.read_text(errors="ignore")) if MASTER.exists() else ""
    # "Things I have NOT done" lists gaps — mentioning a skill there is not having it.
    return re.sub(r"^## Things I have NOT done.*?(?=^## |\Z)", "", t, flags=re.M | re.S)


def cover(job_id, master):
    jd = OUT / "jds" / f"{job_id}.md"
    resume = OUT / "applications" / job_id / "resume.md"
    if not resume.exists():
        return None, "no tailored resume.md"
    if not jd.exists():
        return None, "no JD file"
    kws = keywords(jd.read_text(errors="ignore"))
    if not kws:
        return None, "JD has no ## ATS keywords section"
    text = COMMENT.sub("", resume.read_text(errors="ignore"))
    hit, missed, absent, rephrased = [], [], [], []
    for k in kws:
        rx = pattern(k)
        if rx.search(text):
            hit.append(k)
            if master and not rx.search(master):
                rephrased.append(k)
        elif master and rx.search(master):
            missed.append(k)
        else:
            absent.append(k)
    n = len(kws)
    res = {
        "job_id": job_id,
        "computed": datetime.date.today().isoformat(),
        "total": n,
        "coverage": round(len(hit) / n, 3),
        "supportable": round(len(hit) / (len(hit) + len(missed)), 3) if hit or missed else None,
        "hit": hit,
        "missed": missed,
        "not_in_master": absent,
        "rephrased": rephrased,
    }
    (resume.parent / "keywords.json").write_text(json.dumps(res, indent=2, ensure_ascii=False) + "\n")
    return res, None


def line(r):
    s = f"{round(r['coverage'] * 100)}% ({len(r['hit'])}/{r['total']})"
    if r["missed"]:
        s += f" · missed, in master: {', '.join(r['missed'])}"
    if r["not_in_master"]:
        s += f" · not in master: {', '.join(r['not_in_master'])}"
    if r["rephrased"]:
        s += f" · check rephrase: {', '.join(r['rephrased'])}"
    return s


def main(argv):
    if not argv or argv[0] in ("-h", "--help"):
        print(__doc__.strip())
        return 0 if argv else 2
    if argv[0] == "--all":
        apps = OUT / "applications"
        ids = sorted(p.parent.name for p in apps.glob("*/resume.md")) if apps.exists() else []
    else:
        ids = argv
    master = master_text()
    ok = 0
    for i in ids:
        r, why = cover(i, master)
        if r:
            ok += 1
            print(f"  {i:<60} {line(r)}")
        else:
            print(f"  {i:<60} skipped — {why}")
    if len(ids) > 1:
        print(f"  {ok} of {len(ids)} scored → applications/<job-id>/keywords.json")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
