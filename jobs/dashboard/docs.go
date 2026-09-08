package main

// Document viewer. The table says a resume exists; this serves the thing itself, so
// "tailored" in the Status column is a link you can click rather than a claim you have to
// go and check on disk.
//
// One route, /doc, resolves a job ID plus a file name under jobs/output/ and renders the
// Markdown. It reads from disk on every request for the same reason the index does — a
// re-run of /job tailor shows up on refresh.

import (
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	// A job ID is a slug — never a path. Anything else is refused rather than cleaned,
	// because a "cleaned" path is how a viewer becomes a file-read primitive.
	jobIDRe = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*$`)
	docRe   = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*\.(md|txt|pdf|json|js)$`)

	mdCodeRe = regexp.MustCompile("`([^`]+)`")
	mdLinkRe = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)\)`)
	mdBoldRe = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	mdItalRe = regexp.MustCompile(`\*([^*\s][^*]*)\*`)
)

// linkDocs marks the rows whose tailored resume is actually on disk. A "✓" in the Resume
// column is a claim written by /job tailor; the file is the proof, and only a row with the
// file gets a link.
func linkDocs(p *page, liveDir string) {
	ids := p.Summary.col("Job ID")
	p.ResumeJob = make([]string, len(p.Summary.Rows))
	for i := range p.Summary.Rows {
		id := strings.TrimSpace(cell(ids, i))
		if !jobIDRe.MatchString(id) {
			continue
		}
		if _, err := os.Stat(filepath.Join(liveDir, "applications", id, "resume.md")); err == nil {
			p.ResumeJob[i] = id
		}
	}
}

// doc is one file the viewer can show, plus how to reach it.
type doc struct {
	Name  string // file name on disk, or "jd" for the captured job description
	Label string // what the tab says
	Href  string
	PDF   bool
	Code  bool // .js / .json — shown verbatim, not as Markdown
}

// docsFor lists everything readable for a job: the per-job application folder, then the
// captured JD, which lives elsewhere but is what the resume was written against.
func docsFor(liveDir, id string) []doc {
	var out []doc
	dir := filepath.Join(liveDir, "applications", id)
	ents, _ := os.ReadDir(dir)
	var names []string
	for _, e := range ents {
		if !e.IsDir() && docRe.MatchString(strings.ToLower(e.Name())) {
			names = append(names, e.Name())
		}
	}
	sort.Slice(names, func(i, j int) bool {
		// resume first, then alphabetical — it is what almost every visit is here for.
		ri, rj := strings.HasPrefix(names[i], "resume"), strings.HasPrefix(names[j], "resume")
		if ri != rj {
			return ri
		}
		return names[i] < names[j]
	})
	for _, n := range names {
		out = append(out, doc{
			Name:  n,
			Label: strings.NewReplacer("-", " ", ".md", "", ".txt", " (txt)").Replace(n),
			Code:  strings.HasSuffix(n, ".js") || strings.HasSuffix(n, ".json"),
			Href:  "/doc?job=" + id + "&f=" + n,
			PDF:   strings.HasSuffix(strings.ToLower(n), ".pdf"),
		})
	}
	if _, err := os.Stat(filepath.Join(liveDir, "jds", id+".md")); err == nil {
		out = append(out, doc{Name: "jd", Label: "job description", Href: "/doc?job=" + id + "&f=jd"})
	}
	return out
}

// resolveDoc maps (job, f) to a file on disk. "jd" is the one name that resolves outside
// the job's own folder.
func resolveDoc(liveDir, id, f string) (string, bool) {
	if !jobIDRe.MatchString(id) {
		return "", false
	}
	if f == "jd" {
		p := filepath.Join(liveDir, "jds", id+".md")
		_, err := os.Stat(p)
		return p, err == nil
	}
	if f == "" {
		f = "resume.md"
	}
	if !docRe.MatchString(strings.ToLower(f)) || f != filepath.Base(f) {
		return "", false
	}
	p := filepath.Join(liveDir, "applications", id, f)
	if _, err := os.Stat(p); err != nil {
		return "", false
	}
	return p, true
}

// docHandler serves one document, either as a standalone page or — with frag=1 — as the
// HTML fragment the dashboard's drawer drops in without a reload.
func docHandler(liveDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id, f := strings.TrimSpace(q.Get("job")), strings.TrimSpace(q.Get("f"))
		path, ok := resolveDoc(liveDir, id, f)
		if !ok {
			http.Error(w, "no such document", http.StatusNotFound)
			return
		}
		if strings.HasSuffix(strings.ToLower(path), ".pdf") {
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", "inline")
			http.ServeFile(w, r, path)
			return
		}
		b, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, "could not read "+filepath.Base(path), 500)
			return
		}
		if f == "" {
			f = "resume.md"
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// print=1 is the page the PDF is made of — same layout, straight from resume.md,
		// so ⌘P in the browser and `make pdf` produce the same document.
		if q.Get("print") == "1" {
			fmt.Fprint(w, resumeHTML(parseResume(string(b))))
			return
		}
		body := docFragment(id, f, liveDir, filepath.Base(path), string(b))
		if q.Get("frag") == "1" {
			fmt.Fprint(w, body)
			return
		}
		fmt.Fprintf(w, docPageHTML, html.EscapeString(id), docCSS, body)
	}
}

// isCode says whether a document is shown verbatim rather than rendered as Markdown.
func isCode(name string) bool {
	return strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".json")
}

// docFragment is the shared body: which job, which sibling documents exist, and the
// rendered Markdown. Shared so the drawer and the standalone page can never drift.
func docFragment(id, active, liveDir, base, md string) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="doc-h"><div class="doc-id">%s</div><div class="doc-tabs">`, html.EscapeString(id))
	for _, d := range docsFor(liveDir, id) {
		cls, tgt := "doc-tab", ""
		if d.Name == active || (active == "jd" && d.Name == "jd") {
			cls += " on"
		}
		if d.PDF {
			tgt = ` target="_blank" rel="noopener"`
		}
		fmt.Fprintf(&b, `<a class="%s" href="%s"%s>%s</a>`, cls, html.EscapeString(d.Href), tgt, html.EscapeString(d.Label))
	}
	if strings.HasPrefix(active, "resume") && strings.HasSuffix(active, ".md") {
		fmt.Fprintf(&b, `<a class="doc-tab print" href="/doc?job=%s&f=%s&print=1" target="_blank" rel="noopener">print view ↗</a>`,
			html.EscapeString(id), html.EscapeString(active))
	}
	fmt.Fprintf(&b, `</div><div class="doc-path">%s</div></div>`, html.EscapeString(base))
	if isCode(active) {
		// refill.js exists to be pasted into a browser console, so the useful affordance is
		// a copy button, not prose formatting.
		fmt.Fprintf(&b, `<div class="doc-act"><button class="doc-copy" type="button">copy all</button>`+
			`<span class="doc-hint">paste into the console on the application page</span></div>`+
			`<pre class="code"><code>%s</code></pre>`, html.EscapeString(md))
		return b.String()
	}
	// The HTML-comment receipt /job tailor appends is an audit trail, not the resume.
	fmt.Fprintf(&b, `<article class="md">%s</article>`, mdHTML(commentRe.ReplaceAllString(md, "")))
	return b.String()
}

// mdHTML renders the Markdown subset the workspace actually writes: headings, bullets,
// pipe tables, fenced code, rules, and inline bold/italic/code/links. Everything is
// escaped before any tag is introduced, so a resume can never inject markup.
func mdHTML(src string) string {
	var b strings.Builder
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	var para []string
	inList, inCode := false, false

	flushPara := func() {
		if len(para) > 0 {
			b.WriteString("<p>" + strings.Join(para, "<br>") + "</p>\n")
			para = nil
		}
	}
	flushList := func() {
		if inList {
			b.WriteString("</ul>\n")
			inList = false
		}
	}
	flush := func() { flushPara(); flushList() }

	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		t := strings.TrimSpace(raw)

		if strings.HasPrefix(t, "```") {
			if inCode {
				b.WriteString("</code></pre>\n")
			} else {
				flush()
				b.WriteString("<pre><code>")
			}
			inCode = !inCode
			continue
		}
		if inCode {
			b.WriteString(html.EscapeString(raw) + "\n")
			continue
		}
		switch {
		case t == "":
			flush()
		case t == "---" || t == "***" || t == "___":
			flush()
			b.WriteString("<hr>\n")
		case strings.HasPrefix(t, "|"):
			flush()
			var rows []string
			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "|") {
				rows = append(rows, strings.TrimSpace(lines[i]))
				i++
			}
			i--
			b.WriteString(mdTable(rows))
		case strings.HasPrefix(t, "#"):
			n := 0
			for n < len(t) && t[n] == '#' && n < 6 {
				n++
			}
			if n < len(t) && t[n] == ' ' {
				flush()
				txt := mdInline(strings.TrimSpace(t[n:]))
				fmt.Fprintf(&b, "<h%d>%s</h%d>\n", n, txt, n)
			} else {
				para = append(para, mdInline(t))
			}
		case strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* "):
			flushPara()
			if !inList {
				b.WriteString("<ul>\n")
				inList = true
			}
			b.WriteString("<li>" + mdInline(t[2:]) + "</li>\n")
		case strings.HasPrefix(t, "> "):
			flush()
			b.WriteString("<blockquote>" + mdInline(t[2:]) + "</blockquote>\n")
		default:
			flushList()
			para = append(para, mdInline(t))
		}
	}
	if inCode {
		b.WriteString("</code></pre>\n")
	}
	flush()
	return b.String()
}

// mdTable renders a pipe table, dropping the |---|---| separator row.
func mdTable(rows []string) string {
	var b strings.Builder
	b.WriteString(`<div class="tw"><table>`)
	head := true
	for _, ln := range rows {
		if sepRow.MatchString(ln) {
			continue
		}
		cells := rowSplit.Split(strings.Trim(ln, "|"), -1)
		tag := "td"
		if head {
			tag = "th"
			b.WriteString("<thead>")
		}
		b.WriteString("<tr>")
		for _, c := range cells {
			fmt.Fprintf(&b, "<%s>%s</%s>", tag, mdInline(strings.TrimSpace(c)), tag)
		}
		b.WriteString("</tr>")
		if head {
			b.WriteString("</thead><tbody>")
			head = false
		}
	}
	if !head {
		b.WriteString("</tbody>")
	}
	b.WriteString("</table></div>\n")
	return b.String()
}

// mdInline escapes first, then introduces tags — never the other way round.
func mdInline(s string) string {
	s = html.EscapeString(s)
	s = mdCodeRe.ReplaceAllString(s, "<code>$1</code>")
	s = mdLinkRe.ReplaceAllString(s, `<a href="$2" target="_blank" rel="noopener">$1</a>`)
	s = mdBoldRe.ReplaceAllString(s, "<strong>$1</strong>")
	s = mdItalRe.ReplaceAllString(s, "<em>$1</em>")
	return s
}

// docCSS styles the rendered document. Shared by the standalone page and the dashboard
// drawer, injected into the dashboard template through the docCSS template function.
const docCSS = `
.doc-h{border-bottom:1px solid var(--line);padding-bottom:10px;margin-bottom:18px}
.doc-id{font-size:12px;color:var(--muted);font-family:ui-monospace,SFMono-Regular,Menlo,monospace}
.doc-tabs{display:flex;flex-wrap:wrap;gap:6px;margin:8px 0 6px}
.doc-tab{font-size:12px;padding:2px 9px;border:1px solid var(--line);border-radius:20px;
  text-decoration:none;color:var(--muted);background:var(--panel)}
.doc-tab:hover{color:var(--ink);border-color:var(--muted)}
.doc-tab.on{background:var(--accent);border-color:var(--accent);color:#fff;font-weight:600}
.doc-path{font-size:11px;color:var(--muted)}
.md{font-size:14px;line-height:1.62;max-width:78ch}
.md h1{font-size:23px;margin:0 0 4px;letter-spacing:-.01em}
.md h2{font-size:15px;margin:26px 0 8px;padding-bottom:5px;border-bottom:1px solid var(--line);
  text-transform:uppercase;letter-spacing:.06em;color:var(--muted)}
.md h3{font-size:14.5px;margin:18px 0 4px}
.md h4{font-size:13.5px;margin:14px 0 3px;color:var(--muted)}
.md p{margin:7px 0}
.md ul{margin:7px 0;padding-left:20px}
.md li{margin:3px 0}
.md hr{border:0;border-top:1px solid var(--line);margin:20px 0}
.md code{background:var(--chip);border-radius:4px;padding:1px 4px;font-size:12.5px}
.md pre{background:var(--chip);border-radius:8px;padding:12px;overflow-x:auto}
.md pre code{background:none;padding:0}
.md blockquote{margin:9px 0;padding:2px 12px;border-left:3px solid var(--line);color:var(--muted)}
.md .tw{overflow-x:auto}
.md table{border-collapse:collapse;font-size:13px;margin:10px 0}
.md th,.md td{border:1px solid var(--line);padding:6px 10px;text-align:left}
.md th{background:var(--chip)}
.doc-act{display:flex;align-items:center;gap:10px;margin-bottom:8px}
.doc-copy{font:inherit;font-size:12.5px;padding:4px 11px;border:1px solid var(--line);
  border-radius:6px;background:var(--panel);color:var(--ink);cursor:pointer}
.doc-copy:hover{background:var(--chip)}
.doc-hint{font-size:12px;color:var(--muted)}
pre.code{background:var(--chip);border:1px solid var(--line);border-radius:8px;padding:12px;
  overflow:auto;max-height:70vh;font-size:12px;line-height:1.5;white-space:pre;margin:0}
`

// docPageHTML is the standalone view — same tokens as the dashboard, so opening a resume
// in its own tab does not look like a different application.
const docPageHTML = `<!doctype html>
<html lang="en"><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s</title>
<style>
:root{
  --bg:#fbfbfa; --panel:#fff; --ink:#1c1c1a; --muted:#6b6b66; --line:#e6e5e1;
  --accent:#3b5bdb; --chip:#f2f1ed;
}
@media (prefers-color-scheme:dark){:root{
  --bg:#16161a; --panel:#1e1e23; --ink:#eceae5; --muted:#9a978f; --line:#2e2e35;
  --accent:#8ea2ff; --chip:#26262c;}}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--ink);
  font:14px/1.55 ui-sans-serif,-apple-system,"Segoe UI",Inter,system-ui,sans-serif}
.wrap{max-width:900px;margin:0 auto;padding:26px 22px 70px}
a{color:var(--accent)}
.back{font-size:12.5px;color:var(--muted);text-decoration:none;display:inline-block;margin-bottom:14px}
.back:hover{color:var(--ink)}
%s
</style></head><body><div class="wrap">
<a class="back" href="/">← back to jobs</a>
%s
</div></body></html>`
