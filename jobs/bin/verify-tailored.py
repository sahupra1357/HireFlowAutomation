import re,sys,pathlib
M=pathlib.Path("jobs/input/profile/master-resume.md").read_text()
_n=re.search(r'^- Name:\s*(.+?)\s*$',M,re.M) or re.search(r'^#\s*Master Resume\s*—\s*(.+?)\s*$',M,re.M)
NAME=_n.group(1).strip() if _n else None
MYEARS=set(re.findall(r'\b(19\d\d|20\d\d)\b',M))
CERTS=set(re.findall(r'AWS Certified [A-Za-z]+',M))
def check(p):
    f=pathlib.Path(p)/"resume.md"
    if not f.exists(): return [f"MISSING {f}"]
    r=f.read_text(); bad=[]
    em=[e for e in re.findall(r'[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}',r) if 'TODO' not in e]
    if em: bad.append(f"INVENTED EMAIL: {em}")
    fig=re.findall(r'(?<![\w.])[0-9]{1,3}%|\$[0-9][0-9,.]*[KMB]?\b',r)
    if fig: bad.append(f"INVENTED FIGURES: {fig[:6]}")
    yrs=set(re.findall(r'\b(19\d\d|20\d\d)\b',r))-MYEARS
    if yrs: bad.append(f"YEARS NOT IN MASTER: {sorted(yrs)}")
    cer=set(re.findall(r'AWS Certified [A-Za-z]+',r))-CERTS
    if cer: bad.append(f"CERTS NOT IN MASTER: {cer}")
    if NAME and NAME not in r: bad.append("NAME MISSING/ALTERED")
    # Split into blank-line-separated blocks. A block that anywhere contains a negation
    # marker is a gaps list — naming what you lack is correct, not a violation. Blocks
    # survive line wrapping, which line-by-line heuristics do not.
    NEG=('not claimed','gap','lack','missing','deliberately not','do not have',
         "n't have",'not evidenced','left out','not on master','unconfirmed','not asserted')
    for blk in re.split(r'\n\s*\n', r):
        low=blk.lower()
        if any(k in low for k in NEG): continue
        for tok in ("Scala","Golang","Azure","GCP","OpenShift","Node.js","Kotlin","Ruby","Rails"):
            if re.search(r'\b'+re.escape(tok)+r'\b', blk):
                line=[l for l in blk.split("\n") if tok in l][0]
                bad.append(f"CLAIMED SKILL NOT IN MASTER: {tok} -> {line.strip()[:70]}")
    return bad
for p in sys.argv[1:]:
    b=check(p); n=pathlib.Path(p).name
    print(f"{'✓ CLEAN ' if not b else '✗ REJECT'} {n[:56]}")
    for x in b: print(f"      {x}")
