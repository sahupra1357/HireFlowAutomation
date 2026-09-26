#!/usr/bin/env python3
"""Setup consistency check — is jobs/input/ ready, and is the repo still generic?

    jobs/bin/check-setup.py            # report; exit 1 if any ERROR
    make doctor                        # this plus check-index.py

Checks what every task assumes without re-checking: a resume dropped in, a populated master
resume that is not older than it, a search profile with no placeholders left, a family list,
an answer bank with the blocking questions answered, a way to render the resume in the
user's own format, and the tools the stages shell out to.

It also enforces the rule that matters most for a public repo: **nothing personal is
committed.** The user's name, email and phone are read from master-resume.md and searched for
in every tracked file, and every per-user file must stay untracked.

Reads only. Prints nothing personal — a hit names the file, never the value found.

    --preflight   what `make daily` and `/job` run before searching. Never fails: setup
                  gaps are reported, not enforced, because the pipeline already runs
                  degraded without them (no resume → search only, no confirmed families →
                  the generic list). The ones it degrades around are printed as DEGRADED
                  rather than ERROR; everything else keeps its label.
"""
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(os.environ.get("HIREFLOW_ROOT") or Path(__file__).resolve().parents[2])
INP = ROOT / "jobs" / "input"
PROFILE = INP / "profile"
CONFIG = INP / "config"
SOURCES = PROFILE / "source-resumes"
MASTER = PROFILE / "master-resume.md"
ANSWERS = CONFIG / "application-answers.md"
SEARCH = CONFIG / "search-profile.md"
FAMILIES = CONFIG / "setup-families.md"
DOCX_MAP = CONFIG / "resume-docx-map.json"
FORMAT = CONFIG / "resume-format.md"

# Per-user files: must exist only on this machine, never in git.
PERSONAL = [
    "jobs/input/profile/source-resumes/",
    "jobs/input/profile/master-resume.md",
    "jobs/input/profile/links.md",
    "jobs/input/inbox.md",
    "jobs/input/config/application-answers.md",
    "jobs/input/config/search-profile.md",
    "jobs/input/config/setup-families.md",
    "jobs/input/config/resume-format.md",
    "jobs/input/config/resume-docx-map.json",
    "jobs/output/",
    "interviews/output/",
]
RESUME_EXT = {".pdf", ".docx", ".doc", ".md", ".txt", ".rtf", ".odt"}

errors, warns, notes = [], [], []


PREFLIGHT = "--preflight" in sys.argv[1:]
degraded = []


def err(m): errors.append(m)


def degrade(m):
    """A gap the pipeline runs around. An error for `make doctor`, a note in preflight."""
    (degraded if PREFLIGHT else errors).append(m)
def warn(m): warns.append(m)


def read(p):
    try:
        return p.read_text(errors="ignore")
    except OSError:
        return ""


def git(*args):
    try:
        r = subprocess.run(["git", "-C", str(ROOT), *args], capture_output=True, text=True, timeout=30)
        return r.stdout.splitlines() if r.returncode == 0 else None
    except (OSError, subprocess.TimeoutExpired):
        return None


def section(text, name):
    m = re.search(rf"^## {re.escape(name)}\b.*?$(.*?)(?=^## |\Z)", text, re.M | re.S)
    return m.group(1) if m else ""


def check_profile():
    srcs = [p for p in SOURCES.glob("*") if p.is_file() and p.suffix.lower() in RESUME_EXT] \
        if SOURCES.exists() else []
    if not srcs:
        degrade("no resume in jobs/input/profile/source-resumes/ — drop one in, then /job setup")
    else:
        notes.append(f"source resume: {len(srcs)} file(s) ({', '.join(sorted({p.suffix.lower() for p in srcs}))})")

    m = read(MASTER)
    if not m:
        degrade("jobs/input/profile/master-resume.md missing — run /job setup")
        return srcs, {}
    if "STATUS: not yet populated" in m:
        degrade("master-resume.md is still the blank template — run /job setup")
        return srcs, {}

    contact = section(m, "Contact")
    ident = {}
    for k in ("Name", "Email", "Phone"):
        v = re.search(rf"^- {k}:\s*(.+?)\s*$", contact, re.M)
        v = v.group(1) if v else ""
        if not v or "TODO" in v:
            (degrade if k == "Name" else warn)(f"master-resume.md Contact → {k} is not filled")
        else:
            ident[k] = v
    for sec in ("Experience", "Skills"):
        body = section(m, sec)
        if not body.strip() or re.search(r"^\s*TODO\s*$", body, re.M):
            degrade(f"master-resume.md ## {sec} is empty or TODO — tailoring would have nothing to use")
    if "### " not in section(m, "Experience"):
        warn("master-resume.md ## Experience has no ### role entries — make pdf cannot map it")

    if srcs:
        newest = max(srcs, key=lambda p: p.stat().st_mtime)
        if newest.stat().st_mtime > MASTER.stat().st_mtime + 60:
            warn(f"{newest.suffix} source resume is newer than master-resume.md — "
                 "re-run /job setup so tailoring uses the new version")

    if not (PROFILE / "links.md").exists():
        warn("jobs/input/profile/links.md missing — forms will leave LinkedIn/GitHub blank")
    return srcs, ident


def check_config(srcs):
    if not (CONFIG / "job-sites.md").exists():
        err("jobs/input/config/job-sites.md missing — search has no sites")

    s = read(SEARCH)
    if not s:
        degrade("jobs/input/config/search-profile.md missing — run /job setup")
    else:
        ph = len(re.findall(r"\{\{[^}]*\}\}", s))
        if ph:
            warn(f"search-profile.md has {ph} {{{{placeholder}}}}(s) left — each is a dimension search skips")
        if not re.search(r"^[\s#*-]*Countries:", s, re.M | re.I):
            warn("search-profile.md has no Countries: line — the country hard filter is off")
        if not re.search(r"^### Auto-tailor threshold:\s*\d+", s, re.M):
            warn("search-profile.md has no '### Auto-tailor threshold: N' — /job auto uses 75")
        if "provisional" in s.lower():
            warn("search-profile.md still carries provisional (inferred) role families — confirm them in /job setup")

    f = read(FAMILIES)
    if not f:
        warn("setup-families.md missing — searches fall back to search-profile.md's generic families")
    else:
        fams = [l for l in section(f, "Families").splitlines() if l.startswith("### ") and "{{" not in l]
        if not fams:
            degrade("setup-families.md has no confirmed families — make daily would search nothing")
        else:
            notes.append(f"role families: {len(fams)} confirmed")

    a = read(ANSWERS)
    if not a:
        degrade("jobs/input/config/application-answers.md missing — run /job setup")
    else:
        todo = len(re.findall(r"\|\s*TODO\s*\|", a))
        auth = section(a, "Work authorization")
        blocking = len(re.findall(r"\|\s*TODO\s*\|", auth))
        if blocking:
            warn(f"answer bank: {blocking} work-authorization answer(s) still TODO — "
                 "every batch-filled form will leave them blank")
        notes.append(f"answer bank: {todo} TODO answer(s)")

    # How make pdf will render: the user's own .docx, or the measured layout.
    docx = [p for p in srcs if p.suffix.lower() == ".docx"]
    if docx:
        if not DOCX_MAP.exists():
            warn("a .docx resume is on file but resume-docx-map.json is missing — "
                 "make pdf falls back to the measured layout; re-run /job setup to map it")
        else:
            m = re.search(r'"source"\s*:\s*"([^"]+)"', read(DOCX_MAP))
            if m and not (SOURCES / m.group(1)).exists():
                warn("resume-docx-map.json points at a source .docx that is no longer there — re-run /job setup")
            py = ROOT / ".venv" / "bin" / "python"
            if py.exists():
                ok = subprocess.run([str(py), "-c", "import docx"], capture_output=True).returncode == 0
                if not ok:
                    warn(".venv exists but python-docx is not installed in it — make pdf will install it")
    elif srcs and not FORMAT.exists():
        warn("resume-format.md missing — make pdf renders a generic layout, not yours; re-run /job setup")


def check_tools(srcs):
    b = ROOT / ".claude/skills/gstack/browse/dist/browse"
    if not (os.access(b, os.X_OK) or os.access(Path.home() / ".claude/skills/gstack/browse/dist/browse", os.X_OK)):
        warn("gstack browse not found — jd capture in the browser, apply, and make morning need it")
    if not shutil.which("claude"):
        warn("claude CLI not on PATH — make daily cannot run its agent stages")
    if not shutil.which("go"):
        warn("go not installed — make web / make pdf cannot build the dashboard")
    if not (ROOT / "jobs/output/tracker.md").exists():
        warn("jobs/output/tracker.md missing — run: make init")


def check_repo(ident):
    tracked = git("ls-files")
    if tracked is None:
        notes.append("not a git checkout — skipped the committed-data checks")
        return
    leaked = [p for p in tracked
              if any(p == x or (x.endswith("/") and p.startswith(x)) for x in PERSONAL)
              and not p.endswith(".gitkeep") and not p.endswith("README.md")]
    for p in leaked:
        err(f"per-user file is committed: {p} — git rm --cached it")
    for p in git("ls-files", "-ci", "--exclude-standard") or []:
        if p not in leaked:
            warn(f"tracked file matches .gitignore: {p}")

    # The user's identity must not appear in any committed file.
    needles = {}
    if ident.get("Name") and len(ident["Name"]) >= 5:
        needles["name"] = re.compile(re.escape(ident["Name"]), re.I)
    if ident.get("Email"):
        needles["email"] = re.compile(re.escape(ident["Email"]), re.I)
        local = ident["Email"].split("@")[0]
        if len(local) >= 6:
            needles["email handle"] = re.compile(rf"\b{re.escape(local)}\b", re.I)
    digits = re.sub(r"\D", "", ident.get("Phone", ""))[-10:]
    if len(digits) == 10:
        needles["phone"] = re.compile(r"\D?".join(digits))
    if not needles:
        return
    hits = []
    for p in tracked:
        f = ROOT / p
        if not f.is_file() or f.stat().st_size > 2_000_000:
            continue
        try:
            text = f.read_bytes().decode("utf-8")
        except UnicodeDecodeError:
            continue  # binary
        found = [k for k, rx in needles.items() if rx.search(text)]
        if found:
            hits.append((p, found))
    for p, found in hits:
        err(f"committed file contains your {' + '.join(found)}: {p} — the repo must stay generic")


def main():
    srcs, ident = check_profile()
    check_config(srcs)
    check_tools(srcs)
    check_repo(ident)

    print("Setup consistency — jobs/input/ and the repo" + ("  (preflight)" if PREFLIGHT else ""))
    for m in notes:
        print(f"  · {m}")
    for m in errors:
        print(f"  ERROR  {m}")
    for m in degraded:
        print(f"  DEGRADED  {m}")
    for m in warns:
        print(f"  WARN   {m}")
    if not errors and not warns and not degraded:
        print("  ✓ profile, config and repo are consistent")
    print(f"  {len(errors)} error(s), {len(warns)} warning(s)"
          + (f", {len(degraded)} degraded" if degraded else ""))
    if PREFLIGHT:
        return 0
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
