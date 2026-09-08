#!/usr/bin/env python3
"""Normalized fetcher for /interview-search Phase 1 sources.

Emits one JSON object per line: {id, source, origin, title, url, posted, body}
so the skill never has to parse RSS or Algolia payloads by hand.

  ix-fetch.py reddit --sub ExperiencedDevs --q '"AI interview"' [--since 18m] [--limit 25]
  ix-fetch.py hn --q '"interview experience"' [--tags comment] [--since 18m] [--limit 30]

Exit 0 with zero lines means "searched, found nothing" — a real answer, not a failure.
Exit 2 means the source could not be reached; the caller must record that, never treat
it as an empty result.
"""
import argparse, hashlib, html, json, os, re, sys, tempfile, time, urllib.error
import urllib.parse, urllib.request

UA = "ix-research/0.1 (personal job-prep script)"
TAG = re.compile(r"<[^>]+>")

# Reddit rate-limits unauthenticated RSS aggressively — a second request fired
# immediately after the first returns 429. Each invocation is a separate process, so
# the pacing state lives in a stamp file rather than in memory.
PACE = {"reddit": 12.0, "hn": 0.5}


def pace(source):
    stamp = os.path.join(tempfile.gettempdir(), f".ix-fetch-{source}.stamp")
    gap = PACE.get(source, 1.0)
    try:
        wait = gap - (time.time() - os.path.getmtime(stamp))
        if wait > 0:
            time.sleep(wait)
    except OSError:
        pass
    open(stamp, "w").close()


def months(s):
    m = re.fullmatch(r"(\d+)\s*([my])", (s or "18m").strip().lower())
    if not m:
        return 18
    n = int(m.group(1))
    return n * 12 if m.group(2) == "y" else n


def get(url, source, tries=4):
    """Fetch with pacing and exponential backoff. Raises only after every retry fails,
    so a transient 429 never gets mistaken for an empty result."""
    req = urllib.request.Request(url, headers={"User-Agent": UA, "Accept": "*/*"})
    for attempt in range(tries):
        pace(source)
        try:
            with urllib.request.urlopen(req, timeout=30) as r:
                return r.read().decode("utf-8", "replace")
        except urllib.error.HTTPError as e:
            if e.code not in (429, 500, 502, 503) or attempt == tries - 1:
                raise
            back = 5 * (2 ** attempt)
            print(f"ix-fetch: {source} {e.code}, retry in {back}s", file=sys.stderr)
            time.sleep(back)
    raise RuntimeError("unreachable")


def clean(s, cap=6000):
    s = TAG.sub(" ", html.unescape(s or ""))
    s = re.sub(r"[ \t]+", " ", s)
    s = re.sub(r"\n\s*\n\s*\n+", "\n\n", s)
    return s.strip()[:cap]


def rec(source, origin, title, url, posted, body):
    h = hashlib.sha1(url.encode()).hexdigest()[:8]
    return {"id": h, "source": source, "origin": origin, "title": clean(title, 300),
            "url": url, "posted": posted, "body": clean(body)}


def reddit(a):
    cutoff = time.time() - months(a.since) * 30.4 * 86400
    q = urllib.parse.quote(a.q)
    url = (f"https://www.reddit.com/r/{a.sub}/search.rss?q={q}"
           f"&restrict_sr=1&sort=new&t=year&limit={a.limit}")
    xml = get(url, "reddit")
    out = []
    for e in re.findall(r"<entry>(.*?)</entry>", xml, re.S):
        link = re.search(r'<link href="([^"]+)"', e)
        title = re.search(r"<title>(.*?)</title>", e, re.S)
        upd = re.search(r"<updated>(.*?)</updated>", e, re.S)
        cont = re.search(r'<content type="html">(.*?)</content>', e, re.S)
        if not (link and title):
            continue
        posted = (upd.group(1)[:10] if upd else "")
        if posted:
            try:
                if time.mktime(time.strptime(posted, "%Y-%m-%d")) < cutoff:
                    continue
            except ValueError:
                pass
        out.append(rec("reddit", "r/" + a.sub, title.group(1),
                       link.group(1), posted, cont.group(1) if cont else ""))
    return out


def hn(a):
    cutoff = int(time.time() - months(a.since) * 30.4 * 86400)
    url = ("https://hn.algolia.com/api/v1/search_by_date?query="
           + urllib.parse.quote(a.q)
           + f"&tags={a.tags}&hitsPerPage={a.limit}"
           + f"&numericFilters=created_at_i>{cutoff}")
    data = json.loads(get(url, "hn"))
    out = []
    for h in data.get("hits", []):
        oid = h.get("objectID")
        body = h.get("comment_text") or h.get("story_text") or ""
        title = h.get("title") or h.get("story_title") or clean(body, 120)
        out.append(rec("hn", "news.ycombinator.com", title,
                       f"https://news.ycombinator.com/item?id={oid}",
                       (h.get("created_at") or "")[:10], body))
    return out


p = argparse.ArgumentParser()
sub = p.add_subparsers(dest="cmd", required=True)
r = sub.add_parser("reddit"); r.add_argument("--sub", required=True)
h_ = sub.add_parser("hn"); h_.add_argument("--tags", default="(story,comment)")
for s in (r, h_):
    s.add_argument("--q", required=True)
    s.add_argument("--since", default="18m")
    s.add_argument("--limit", type=int, default=25)
a = p.parse_args()

try:
    rows = reddit(a) if a.cmd == "reddit" else hn(a)
except Exception as e:                                  # noqa: BLE001
    print(f"ix-fetch: {a.cmd} unreachable: {e}", file=sys.stderr)
    sys.exit(2)
for row in rows:
    print(json.dumps(row, ensure_ascii=False))
