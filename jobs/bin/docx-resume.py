#!/usr/bin/env python3
"""Tailored resumes in the user's own .docx — the file is the format, only the words change.

    docx-resume.py map                 # /job setup: find the content slots in the source .docx
    docx-resume.py render <job-id|all> # tailored resume.md → resume.docx (+ resume.pdf via Word)
    docx-resume.py pdf <in.docx> <out.pdf>

`map` reads jobs/input/profile/source-resumes/*.docx and writes
jobs/input/config/resume-docx-map.json: which paragraphs are the summary, the skill lines,
each job's bullets, the education lines. It works from structure, not from any one
resume's wording — a bold paragraph ending in ":" is a section heading, Word-numbered
paragraphs are bullets, "Label: items" lines are skill groups. The JSON is plain and
editable, and `/job setup` shows it to the user.

`render` copies the source .docx for each job and swaps text into those slots:

  * a paragraph keeps its own paragraph and run formatting — only the characters change
  * surplus slots are deleted, extra ones are cloned from a sibling, so they inherit its format
  * headers, name, contact, dates, titles, margins, fonts, borders are never touched
  * education / certification lines are kept or dropped, never rewritten; an "earlier career"
    paragraph is kept as written

It adds nothing that is not in resume.md, which /job tailor wrote and verify-tailored.py
checked against master-resume.md. Needs python-docx (.venv/bin/python).
"""
import copy, glob, json, os, re, shutil, subprocess, sys, tempfile

import docx
from docx.oxml.ns import qn

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
SRC_DIR = os.path.join(ROOT, "jobs/input/profile/source-resumes")
MAP = os.path.join(ROOT, "jobs/input/config/resume-docx-map.json")
APPS = os.path.join(ROOT, "jobs/output/applications")


def source_docx():
    hits = sorted(glob.glob(os.path.join(SRC_DIR, "*.docx")))
    hits = [h for h in hits if not os.path.basename(h).startswith("~$")]
    return hits[0] if hits else None


# ── structure detection ──────────────────────────────────────────────────────

def text(p):
    return "".join(t.text or "" for t in p._p.iter(qn("w:t")))


def is_bullet(p):
    ppr = p._p.pPr
    if ppr is not None and ppr.find(qn("w:numPr")) is not None:
        return True
    st = p.style
    while st is not None:
        if st.element.pPr is not None and st.element.pPr.find(qn("w:numPr")) is not None:
            return True
        st = st.base_style
    return False


def all_bold(p):
    runs = [r for r in p._p.iter(qn("w:r")) if "".join(t.text or "" for t in r.iter(qn("w:t"))).strip()]
    if not runs:
        return False
    def bold(r):
        rpr = r.find(qn("w:rPr"))
        b = rpr.find(qn("w:b")) if rpr is not None else None
        return b is not None and b.get(qn("w:val")) not in ("0", "false")
    return all(bold(r) for r in runs)


def is_heading(p):
    t = text(p).strip()
    return bool(t) and len(t) < 60 and (t.endswith(":") or p.style.name.lower().startswith("heading 1")) \
        and (all_bold(p) or p.style.name.lower().startswith("heading 1")) and not is_bullet(p)


KIND = [("summary", ("summary", "profile", "objective", "about")),
        ("skills", ("skill", "competenc", "technolog", "expertise")),
        ("experience", ("experience", "employment", "work history")),
        ("earlier", ("earlier", "previous", "additional experience")),
        ("education", ("education", "certif", "training", "qualification")),
        ("projects", ("project",))]


def kind_of(heading):
    h = heading.lower()
    for k, keys in KIND:
        if any(x in h for x in keys):
            return k
    return "other"


LABEL_RE = re.compile(r"^\s*([^:]{2,60}):\s*(.+)$")


def build_map(path):
    d = docx.Document(path)
    ps = d.paragraphs
    secs, cur = [], None
    for i, p in enumerate(ps):
        if is_heading(p):
            cur = {"heading": text(p).strip(), "kind": kind_of(text(p)), "paras": []}
            secs.append(cur)
        elif cur is not None and text(p).strip():
            cur["paras"].append(i)
    m = {"source": os.path.basename(path), "summary": [], "skills": [], "jobs": [], "earlier": [],
         "education": [], "unmapped_sections": []}
    for s in secs:
        k, idx = s["kind"], s["paras"]
        if k == "summary":
            m["summary"] = idx
        elif k == "skills":
            m["skills"] = [i for i in idx if LABEL_RE.match(text(ps[i]))]
        elif k == "experience":
            job = None
            for i in idx:
                if is_bullet(ps[i]):
                    if job is None:
                        continue
                    job["bullets"].append(i)
                else:
                    if job is None or job["bullets"]:
                        job = {"header": [], "bullets": []}
                        m["jobs"].append(job)
                    job["header"].append(i)
            for j in m["jobs"]:
                h = text(ps[j["header"][0]])
                j["key"] = re.split(r"[,\t]| {2,}", h.strip())[0].strip()
        elif k == "earlier":
            m["earlier"] = idx
        elif k == "education":
            m["education"] += idx
        else:
            m["unmapped_sections"].append(s["heading"])
    return m


# ── resume.md (what /job tailor wrote) ───────────────────────────────────────

def clean(s):
    s = re.sub(r"\*\*(.+?)\*\*", r"\1", s)
    s = re.sub(r"(?<!\w)\*(.+?)\*(?!\w)", r"\1", s)
    return s.replace("`", "").strip()


def parse_md(md):
    md = re.sub(r"<!--.*?-->", "", md, flags=re.S)
    secs, cur = [], None
    for ln in md.splitlines():
        if ln.startswith("## "):
            cur = {"h": ln[3:].strip(), "lines": []}
            secs.append(cur)
        elif cur is not None:
            cur["lines"].append(ln.rstrip())
    out = {"summary": [], "skills": [], "jobs": [], "earlier": [], "edu": []}
    for n, s in enumerate(secs):
        h = s["h"].lower()
        paras, buf = [], []
        for ln in s["lines"] + [""]:
            if ln.strip() in ("", "---"):
                if buf:
                    paras.append(" ".join(buf)); buf = []
            else:
                buf.append(ln.strip())
        if n == 0 and "skill" not in h and not any(l.lstrip().startswith(("###", "- ")) for l in s["lines"]):
            out["summary"] = [clean(p) for p in paras]
        elif "skill" in h:
            for ln in s["lines"]:
                m = re.match(r"^\s*[-*]?\s*\*\*(.+?):\*\*\s*(.*)$", ln)
                if m:
                    out["skills"].append((clean(m.group(1)), clean(m.group(2))))
        elif "experience" in h or "employment" in h:
            job = None
            for ln in s["lines"]:
                t = ln.strip()
                if t.startswith("### "):
                    hd = t[4:]
                    if hd.lower().startswith("earlier career"):
                        job = {"earlier": True, "bullets": []}
                        out["jobs"].append(job)
                        continue
                    role, _, rest = hd.partition(" — ")
                    job = {"company": rest.split(" · ")[0].strip(), "role": role.strip(), "bullets": []}
                    out["jobs"].append(job)
                elif t.startswith(("- ", "* ")) and job is not None and not t[2:].lstrip().startswith("*Keywords"):
                    job["bullets"].append(clean(t[2:]))
            for j in [j for j in out["jobs"] if j.get("earlier")]:
                out["earlier"] = j["bullets"]
            out["jobs"] = [j for j in out["jobs"] if not j.get("earlier")]
        elif "earlier" in h:
            out["earlier"] = [clean(p.lstrip("-* ")) for p in paras]
        elif any(k in h for k in ("education", "certif")):
            for ln in s["lines"]:
                if ln.strip():
                    out["edu"].append(clean(ln.strip().lstrip("-* ")))
    return out


# ── editing the copy ─────────────────────────────────────────────────────────

def first_rpr(p, want_bold=None):
    for r in p._p.iter(qn("w:r")):
        if not "".join(t.text or "" for t in r.iter(qn("w:t"))).strip():
            continue
        rpr = r.find(qn("w:rPr"))
        b = rpr is not None and rpr.find(qn("w:b")) is not None
        if want_bold is None or b == want_bold:
            return copy.deepcopy(rpr) if rpr is not None else None
    return None


def new_run(txt, rpr):
    r = docx.oxml.OxmlElement("w:r")
    if rpr is not None:
        r.append(copy.deepcopy(rpr))
    t = docx.oxml.OxmlElement("w:t")
    t.set("{http://www.w3.org/XML/1998/namespace}space", "preserve")
    t.text = txt
    r.append(t)
    return r


def clear_runs(p):
    for el in list(p._p):
        if el.tag in (qn("w:r"), qn("w:hyperlink"), qn("w:ins"), qn("w:smartTag"), qn("w:proofErr")):
            p._p.remove(el)


def set_text(p, txt):
    rpr = first_rpr(p)
    clear_runs(p)
    p._p.append(new_run(txt, rpr))


def set_labeled(p, label, items):
    lab = first_rpr(p, True)
    if lab is None:
        lab = first_rpr(p)
    itm = first_rpr(p, False)
    if itm is None and lab is not None:
        itm = copy.deepcopy(lab)
        for b in itm.findall(qn("w:b")) + itm.findall(qn("w:bCs")):
            itm.remove(b)
    clear_runs(p)
    p._p.append(new_run(label + ":", lab))
    p._p.append(new_run(" " + items, itm))


def fill(paras, values, setter):
    """Reuse paras in order; clone the last for extras; delete the surplus."""
    if not paras:
        return len(values)
    # A repeated list (skill lines, bullets) is meant to look uniform. When one slot carries a
    # stray property of its own — a leftover hanging indent, say — content moved into it wraps
    # oddly, so every slot takes the paragraph properties most of its siblings share.
    if len(paras) > 2:
        from collections import Counter
        ppr = lambda p: str(p._p.pPr.xml) if p._p.pPr is not None else ""
        common, n = Counter(ppr(p) for p in paras).most_common(1)[0]
        if common and n > len(paras) / 2:
            tmpl = next(p._p.pPr for p in paras if ppr(p) == common)
            for p in paras:
                if ppr(p) != common:
                    if p._p.pPr is not None:
                        p._p.remove(p._p.pPr)
                    p._p.insert(0, copy.deepcopy(tmpl))
    for i, v in enumerate(values):
        if i < len(paras):
            setter(paras[i], v)
        else:
            el = copy.deepcopy(paras[-1]._p)
            paras[-1]._p.addnext(el)
            np = docx.text.paragraph.Paragraph(el, paras[-1]._parent)
            setter(np, v)
            paras.append(np)
    for p in paras[len(values):]:
        p._p.getparent().remove(p._p)
    return 0


def first_sentence(s):
    """Up to the first full stop that ends a sentence (". " then a capital, or end of text)."""
    s = s.strip()
    m = re.search(r"\.(?=\s+[A-Z(]|\s*$)", s)
    return s[: m.end()] if m else s


def words(s):
    return set(re.findall(r"[a-z0-9]+", s.lower()))


def render(job_id, m, src):
    md_path = os.path.join(APPS, job_id, "resume.md")
    md = parse_md(open(md_path).read())
    d = docx.Document(src)
    ps = d.paragraphs
    P = lambda idx: [ps[i] for i in idx]
    notes = []
    if md["summary"] and m["summary"]:
        summ = " ".join(md["summary"])
        if m.get("preferences", {}).get("keep_summary_first_sentence"):
            # The user's opening sentence stays exactly as they wrote it; tailoring may only
            # add or change what follows it.
            first = first_sentence(text(ps[m["summary"][0]]))
            rest = summ[len(first_sentence(summ)):].strip()
            if first:
                summ = (first + " " + rest).strip()
        fill(P(m["summary"]), [summ], set_text)
    if md["skills"] and m["skills"]:
        fill(P(m["skills"]), md["skills"], lambda p, v: set_labeled(p, v[0], v[1]))
    used = set()
    for tj in md["jobs"]:
        hit = next((j for n, j in enumerate(m["jobs"]) if n not in used and
                    (j["key"].lower() in tj["company"].lower() or tj["company"].lower() in j["key"].lower())), None)
        if hit is None:
            notes.append(f"job not in the .docx, left out: {tj['company']}")
            continue
        used.add(m["jobs"].index(hit))
        fill(P(hit["bullets"]), tj["bullets"], set_text)
    for n, j in enumerate(m["jobs"]):
        if n not in used:
            notes.append(f"job kept as in the .docx (not in resume.md): {j['key']}")
    # "Earlier career" is fixed history (and often names employers the tailored text folds
    # into a heading), so the user's own paragraph is kept as written.
    if md["edu"] and m["education"]:
        keep = [words(e) for e in md["edu"]]
        for p in P(m["education"]):
            w = words(text(p))
            if w and not any(len(w & k) / max(1, min(len(w), len(k))) >= 0.6 for k in keep):
                p._p.getparent().remove(p._p)
    out = os.path.join(APPS, job_id, "resume.docx")
    d.save(out)
    return out, notes


# ── PDF through Word ─────────────────────────────────────────────────────────

WORD_BOX = os.path.expanduser("~/Library/Containers/com.microsoft.Word/Data/tmp")


def pdf(docx_path, pdf_path):
    """Word itself exports the PDF, so it matches what the user would get from File → Save As.
    Word is sandboxed: working inside its own container avoids a file-access prompt per folder."""
    if not os.path.isdir("/Applications/Microsoft Word.app"):
        return "Microsoft Word not installed — resume.docx only"
    os.makedirs(WORD_BOX, exist_ok=True)
    tmp_in = os.path.join(WORD_BOX, "hireflow-resume.docx")
    tmp_out = os.path.join(WORD_BOX, "hireflow-resume.pdf")
    shutil.copy(docx_path, tmp_in)
    if os.path.exists(tmp_out):
        os.remove(tmp_out)
    script = f'''
tell application "Microsoft Word"
    set d to open file name (POSIX file "{tmp_in}" as string)
    save as active document file name (POSIX file "{tmp_out}" as string) file format format PDF
    close active document saving no
end tell'''
    r = subprocess.run(["osascript", "-e", script], capture_output=True, text=True, timeout=120)
    if not os.path.exists(tmp_out):
        return "Word export failed: " + (r.stderr.strip() or r.stdout.strip())
    shutil.move(tmp_out, pdf_path)
    os.remove(tmp_in)
    return None


# ── CLI ──────────────────────────────────────────────────────────────────────

def main(argv):
    if len(argv) < 2 or argv[1] not in ("map", "render", "pdf"):
        print(__doc__); return 2
    if argv[1] == "pdf":
        err = pdf(argv[2], argv[3]); print(err or f"✓ {argv[3]}"); return 1 if err else 0
    src = source_docx()
    if not src:
        print(f"no .docx in {SRC_DIR}"); return 1
    if argv[1] == "map":
        m = build_map(src)
        if os.path.exists(MAP):   # user preferences outlive a re-scan
            m["preferences"] = json.load(open(MAP)).get("preferences", {})
        m.setdefault("preferences", {"keep_summary_first_sentence": True})
        json.dump(m, open(MAP, "w"), indent=2)
        d = docx.Document(src)
        print(f"mapped {os.path.basename(src)} → {os.path.relpath(MAP, ROOT)}")
        print(f"  summary {len(m['summary'])} para · skills {len(m['skills'])} lines · "
              f"jobs {len(m['jobs'])} ({', '.join(j['key'] + ' ' + str(len(j['bullets'])) + ' bullets' for j in m['jobs'])}) · "
              f"earlier {len(m['earlier'])} · education {len(m['education'])}")
        if m["unmapped_sections"]:
            print("  left as-is (not tailored): " + ", ".join(m["unmapped_sections"]))
        return 0
    if not os.path.exists(MAP):
        print("no resume-docx-map.json — run: docx-resume.py map"); return 1
    m = json.load(open(MAP))
    only = argv[2] if len(argv) > 2 else "all"
    jobs = sorted(os.path.basename(os.path.dirname(p)) for p in glob.glob(os.path.join(APPS, "*", "resume.md")))
    if only != "all":
        jobs = [j for j in jobs if j == only]
    bad = 0
    for j in jobs:
        out, notes = render(j, m, src)
        err = pdf(out, os.path.join(APPS, j, "resume.pdf"))
        print(f"  {'✓' if not err else '⚠'} {j} → resume.docx" + ("" if err else " + resume.pdf") + (f"  ({err})" if err else ""))
        for n in notes:
            print(f"      {n}")
        bad += bool(err)
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
