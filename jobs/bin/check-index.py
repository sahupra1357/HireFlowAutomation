#!/usr/bin/env python3
"""Workspace health check — is jobs/output/ internally consistent?

    jobs/bin/check-index.py            # report; exit 1 if any ERROR
    make doctor                        # this plus check-setup.py

Every stage merges into one Markdown index, and three stage columns there make claims about
files on disk. This checks each claim against the disk and the index against itself. It
reads only; it never edits jobs.md or tracker.md — a finding is fixed by re-running the task
that owns the column, or by hand after `jobs/bin/snapshot.sh`.

ERROR  the dashboard or a later stage will misbehave (broken row, duplicate ID, a ✓ with no
       file behind it, a status no stage can reach).
WARN   drift worth a look (orphan files, tracker disagreeing with the index, a JD the user
       dropped in that nobody has picked up yet).
"""
import os
import re
import sys
from pathlib import Path

ROOT = Path(os.environ.get("HIREFLOW_ROOT") or Path(__file__).resolve().parents[2])
OUT = ROOT / "jobs" / "output"
JOBS = OUT / "jobs.md"
TRACKER = OUT / "tracker.md"
JDS = OUT / "jds"
APPS = OUT / "applications"
LOCK = OUT / ".daily-run.lock"

STATUSES = ["found", "shortlisted", "jd-captured", "tailored", "filled-awaiting-user",
            "submitted", "interviewing", "offer", "rejected", "ghosted", "skipped"]
RANK = {s: i for i, s in enumerate(STATUSES)}
JOB_ID = re.compile(r"^[a-z0-9][a-z0-9-]*--[a-z0-9][a-z0-9-]*$")
STAGE_OK = re.compile(r"^(✓ \d{4}-\d{2}-\d{2}|⏳ manual|⏳ in-progress|—)$")

errors, warns, notes = [], [], []


def err(m): errors.append(m)
def warn(m): warns.append(m)


def tables(text):
    """{section heading: (header, [rows], [line numbers])} for every pipe table."""
    out, section, header, rows, lines = {}, None, None, [], []

    def flush():
        if section and header:
            out.setdefault(section, (header, rows, lines))

    for n, line in enumerate(text.splitlines(), 1):
        if line.startswith("## "):
            flush()
            section, header, rows, lines = line[3:].strip(), None, [], []
            continue
        if not line.startswith("|"):
            continue
        cells = [c.strip() for c in line.strip().strip("|").split("|")]
        if all(re.fullmatch(r":?-+:?", c) for c in cells if c):
            continue
        if header is None:
            header = cells
        else:
            rows.append(cells)
            lines.append(n)
    flush()
    return out


def find(tbls, prefix):
    for k, v in tbls.items():
        if k.lower().startswith(prefix):
            return v
    return None


def is_stub(p):
    try:
        return "manual-required" in p.read_text(errors="ignore")
    except OSError:
        return False


def main():
    if not JOBS.exists():
        print("jobs/output/jobs.md not found — nothing to check yet (run /job first).")
        return 0
    tbls = tables(JOBS.read_text(errors="ignore"))
    summary = find(tbls, "summary")
    if not summary:
        err("jobs.md has no ## Summary table — the dashboard cannot render it")
        return report()

    header, rows, lines = summary
    col = {h: i for i, h in enumerate(header)}
    need = ["Job ID", "JD", "Resume", "Form", "Status"]
    missing = [c for c in need if c not in col]
    if missing:
        err(f"Summary header is missing {', '.join(missing)}")
        return report()

    seen = {}
    ids_in_index = set()
    for cells, ln in zip(rows, lines):
        where = f"jobs.md:{ln}"
        if len(cells) != len(header):
            err(f"{where} has {len(cells)} cells, header has {len(header)} — the table renders crooked")
            continue
        jid = cells[col["Job ID"]].strip("` ")
        status = cells[col["Status"]]
        jd, resume, form = cells[col["JD"]], cells[col["Resume"]], cells[col["Form"]]

        if not JOB_ID.match(jid):
            err(f"{where} job ID `{jid}` is not <company-slug>--<role-slug>")
        if jid in seen:
            err(f"{where} duplicate job ID `{jid}` (also line {seen[jid]})")
        seen[jid] = ln
        ids_in_index.add(jid)

        if status not in RANK:
            err(f"{where} `{jid}` has non-canonical Status `{status}`")
        for name, v in (("JD", jd), ("Resume", resume), ("Form", form)):
            if not STAGE_OK.match(v):
                err(f"{where} `{jid}` {name} column is `{v}` — expected ✓ <date>, ⏳ manual/in-progress, or —")

        jd_file = JDS / f"{jid}.md"
        app = APPS / jid
        # JD column vs disk
        if jd.startswith("✓"):
            if not jd_file.exists():
                err(f"`{jid}` JD ✓ but jobs/output/jds/{jid}.md is missing")
            elif is_stub(jd_file):
                err(f"`{jid}` JD ✓ but its JD file is still a manual-required stub")
        elif jd.startswith("⏳"):
            srcs = list(JDS.glob(f"{jid}-source.*"))
            if jd_file.exists() and not is_stub(jd_file):
                warn(f"`{jid}` JD says ⏳ manual but a real JD is on disk — /job jd {jid} will flip it")
            elif srcs:
                warn(f"`{jid}` has a dropped-in {srcs[0].name} not picked up yet — run /job jd {jid}")
        elif jd_file.exists() and not is_stub(jd_file):
            warn(f"`{jid}` JD column is — but jobs/output/jds/{jid}.md exists")

        # Resume column vs disk
        if resume.startswith("✓"):
            if not (app / "resume.md").exists():
                err(f"`{jid}` Resume ✓ but applications/{jid}/resume.md is missing")
            elif not (app / "resume.pdf").exists():
                warn(f"`{jid}` has resume.md but no resume.pdf — run: make pdf JOB={jid}")
            kw = app / "keywords.json"
            newest = max(p.stat().st_mtime for p in (app / "resume.md", jd_file) if p.exists())
            if not kw.exists():
                warn(f"`{jid}` has no keyword coverage — run: make keywords JOB={jid}")
            elif kw.stat().st_mtime < newest:
                warn(f"`{jid}` keyword coverage is older than its resume/JD — run: make keywords JOB={jid}")
        elif (app / "resume.md").exists():
            warn(f"`{jid}` Resume column is — but applications/{jid}/resume.md exists")

        # Form column vs disk
        if form.startswith("✓") and not (app / "form-fill.json").exists() \
                and not (app / "log.md").exists():
            warn(f"`{jid}` Form ✓ but applications/{jid}/ has neither form-fill.json nor log.md")

        # Status vs stage columns — Status can never be ahead of the artifacts it implies
        r = RANK.get(status, -1)
        if RANK["jd-captured"] <= r <= RANK["filled-awaiting-user"] and not jd.startswith("✓"):
            err(f"`{jid}` Status {status} but JD is `{jd}`")
        if RANK["tailored"] <= r <= RANK["filled-awaiting-user"] and not resume.startswith("✓"):
            err(f"`{jid}` Status {status} but Resume is `{resume}`")
        if status == "filled-awaiting-user" and not form.startswith("✓"):
            err(f"`{jid}` Status filled-awaiting-user but Form is `{form}`")
        if status == "found" and resume.startswith("✓"):
            warn(f"`{jid}` Status found but a resume is tailored — Status looks stale")

    # IDs anywhere else in the index
    other_ids = set()
    exc = find(tbls, "excluded")
    if exc:
        h, rs, ls = exc
        if "Job ID" in h:
            i = h.index("Job ID")
            for cells, ln in zip(rs, ls):
                if i < len(cells):
                    x = cells[i].strip("` ")
                    if x in ids_in_index:
                        err(f"jobs.md:{ln} `{x}` is in both Summary and Excluded")
                    other_ids.add(x)
    backlog = find(tbls, "unverified")
    if backlog:
        h, rs, ls = backlog
        for cells, ln in zip(rs, ls):
            if len(cells) != len(h):
                err(f"jobs.md:{ln} backlog row has {len(cells)} cells, header has {len(h)}")
            if "Status" in h and len(cells) == len(h) and cells[h.index("Status")] not in RANK:
                err(f"jobs.md:{ln} backlog row has non-canonical Status `{cells[h.index('Status')]}`")

    # Orphans on disk
    known = ids_in_index | other_ids
    if JDS.exists():
        for f in sorted(JDS.glob("*.md")):
            if f.stem not in known:
                warn(f"orphan JD jobs/output/jds/{f.name} — no row in jobs.md")
    if APPS.exists():
        for d in sorted(p for p in APPS.iterdir() if p.is_dir()):
            if d.name not in known:
                warn(f"orphan folder jobs/output/applications/{d.name}/ — no row in jobs.md")

    # Tracker vs index
    if TRACKER.exists():
        t = tables(TRACKER.read_text(errors="ignore"))
        act = find(t, "active")
        if act:
            h, rs, ls = act
            if "Job ID" in h:
                idx = h.index("Job ID")
                st = h.index("Status") if "Status" in h else None
                sum_status = {r[col["Job ID"]].strip("` "): r[col["Status"]]
                              for r in rows if len(r) == len(header)}
                tseen = set()
                for cells, ln in zip(rs, ls):
                    x = cells[idx].strip("` ") if idx < len(cells) else ""
                    if not x or x.startswith("_("):
                        continue
                    if x in tseen:
                        err(f"tracker.md:{ln} duplicate job ID `{x}`")
                    tseen.add(x)
                    if x not in known:
                        warn(f"tracker.md:{ln} `{x}` is not in jobs.md")
                    elif st is not None and st < len(cells) and x in sum_status \
                            and cells[st] != sum_status[x]:
                        warn(f"`{x}` Status differs: jobs.md `{sum_status[x]}` vs tracker `{cells[st]}`")
    else:
        warn("jobs/output/tracker.md missing — run: make init")

    # Stale daily-run lock
    # daily-run.sh calls this while holding the lock itself — its own lock is not news.
    pidf = LOCK / "pid"
    if pidf.exists() and not os.environ.get("HIREFLOW_DAILY_RUN"):
        try:
            os.kill(int(pidf.read_text().strip()), 0)
            notes.append("a daily run is in progress right now")
        except (ValueError, ProcessLookupError, PermissionError):
            warn("stale jobs/output/.daily-run.lock from a run that died — the next make daily clears it")

    notes.append(f"{len(rows)} jobs in Summary, "
                 f"{len(backlog[1]) if backlog else 0} in backlog, {len(other_ids)} excluded")
    return report()


def report():
    print("Workspace health — jobs/output/")
    for m in notes:
        print(f"  · {m}")
    for m in errors:
        print(f"  ERROR  {m}")
    for m in warns:
        print(f"  WARN   {m}")
    if not errors and not warns:
        print("  ✓ index, stage columns and files agree")
    print(f"  {len(errors)} error(s), {len(warns)} warning(s)")
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
