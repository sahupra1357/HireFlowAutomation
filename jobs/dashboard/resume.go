package main

// Resume PDF rendering.
//
// /job tailor writes resume.md — the words. This turns those words back into the layout of
// the resume the user actually sends out: the two-column format of the source resume in
// jobs/input/profile/source-resumes/, measured off that file — US Letter, 1in margins, a
// 2.4in shaded skills sidebar on the left, the main column at 2.47in, Aptos 12/14.6.
// Nothing here invents content; it only decides where on the page each line already in
// resume.md goes.
//
//	bin/jobs-dashboard -pdf <job-id>   # one job
//	bin/jobs-dashboard -pdf all        # every job with a resume.md
//
// The PDF engine is whatever Chrome is on the machine, in --headless --print-to-pdf mode:
// no LibreOffice, no pandoc, no Go dependency.

import (
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// commentRe strips the HTML-comment tailoring receipt /job tailor appends to resume.md.
// It is an audit trail for the user, not part of the resume, and must never reach the PDF
// (or the viewer).
var commentRe = regexp.MustCompile(`(?s)<!--.*?-->`)

type skillGroup struct {
	Label string
	Items string
}

type jobEntry struct {
	Company, Location, Role, Dates string
	Raw                            string   // the whole ### line, when it doesn't parse
	Blocks                         []string // rendered HTML for the entry body
}

type section struct {
	Heading string
	Body    []string // rendered HTML
}

type resumeDoc struct {
	Name     string
	Title    string
	Contact  string
	Summary  []string
	Skills   []skillGroup
	Sections []section
}

// parseResume reads the Markdown /job tailor writes. The shape it expects is the shape of
// jobs/input/templates/master-resume.md: an H1 name, contact lines, an H2 role title with a
// summary under it, an H2 "Skills" section of "**Label:** items" lines, then any number of
// further H2 sections with H3 job entries and bullets.
func parseResume(md string) resumeDoc {
	var d resumeDoc
	lines := strings.Split(commentRe.ReplaceAllString(md, ""), "\n")

	// Head: the H1 name and everything before the first H2 is contact detail.
	i := 0
	var contact []string
	for ; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if strings.HasPrefix(t, "## ") {
			break
		}
		switch {
		case strings.HasPrefix(t, "# "):
			d.Name = strings.TrimSpace(t[2:])
		case t != "":
			contact = append(contact, mdInline(t))
		}
	}
	d.Contact = strings.Join(contact, " &nbsp;·&nbsp; ")

	// The rest is H2 sections in file order.
	type rawSec struct {
		head string
		body []string
	}
	var secs []rawSec
	for ; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if strings.HasPrefix(t, "## ") {
			secs = append(secs, rawSec{head: strings.TrimSpace(t[3:])})
			continue
		}
		if len(secs) > 0 {
			secs[len(secs)-1].body = append(secs[len(secs)-1].body, lines[i])
		}
	}

	for n, s := range secs {
		switch {
		// The first section, if it carries no structure, is the target-role title and the
		// summary beneath it — that is how the template writes it.
		case n == 0 && !hasStructure(s.body) && !isSkills(s.head):
			d.Title = s.head
			for _, p := range paras(s.body) {
				d.Summary = append(d.Summary, p)
			}
		case isSkills(s.head):
			d.Skills = append(d.Skills, parseSkills(s.body)...)
		default:
			d.Sections = append(d.Sections, section{Heading: s.head, Body: renderBody(s.body)})
		}
	}
	return d
}

func isSkills(h string) bool { return strings.Contains(strings.ToLower(h), "skill") }

// hasStructure reports whether a section body is anything other than prose.
func hasStructure(body []string) bool {
	for _, ln := range body {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "###") || strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") ||
			(strings.HasPrefix(t, "**") && strings.Contains(t, ":**")) {
			return true
		}
	}
	return false
}

// paras joins consecutive non-blank lines into one paragraph each.
func paras(body []string) []string {
	var out []string
	var cur []string
	flush := func() {
		if len(cur) > 0 {
			out = append(out, mdInline(strings.Join(cur, " ")))
			cur = nil
		}
	}
	for _, ln := range body {
		if t := strings.TrimSpace(ln); t == "" {
			flush()
		} else {
			cur = append(cur, t)
		}
	}
	flush()
	return out
}

var skillRe = regexp.MustCompile(`^\*\*(.+?):\*\*\s*(.*)$`)

// parseSkills turns "**Label:** a, b, c" lines into the sidebar's bold-label blocks. A
// bullet or bare line without a label joins the block above it.
func parseSkills(body []string) []skillGroup {
	var out []skillGroup
	for _, ln := range body {
		t := strings.TrimSpace(ln)
		if t == "" {
			continue
		}
		t = strings.TrimPrefix(strings.TrimPrefix(t, "- "), "* ")
		if m := skillRe.FindStringSubmatch(t); m != nil {
			out = append(out, skillGroup{Label: mdInline(m[1]), Items: mdInline(m[2])})
			continue
		}
		if n := len(out); n > 0 {
			out[n-1].Items = strings.TrimSpace(out[n-1].Items + " " + mdInline(t))
		} else {
			out = append(out, skillGroup{Items: mdInline(t)})
		}
	}
	return out
}

var dateRe = regexp.MustCompile(`\d{4}`)

// renderBody renders a main-column section: H3 job entries, bold sub-headings, bullets and
// prose, in the order they appear.
func renderBody(body []string) []string {
	var out []string
	var bullets []string
	flushBullets := func() {
		if len(bullets) > 0 {
			out = append(out, "<ul>"+strings.Join(bullets, "")+"</ul>")
			bullets = nil
		}
	}
	var para []string
	// Lines inside a section body are their own lines — two degrees under Education are two
	// lines, not one run-on sentence. Prose that must reflow (the summary) goes through
	// paras() instead, which joins on spaces.
	flushPara := func() {
		if len(para) > 0 {
			out = append(out, "<p>"+strings.Join(para, "<br>")+"</p>")
			para = nil
		}
	}
	flush := func() { flushBullets(); flushPara() }

	for _, ln := range body {
		t := strings.TrimSpace(ln)
		switch {
		case t == "", t == "---", t == "***", t == "___":
			flush()
		case strings.HasPrefix(t, "### "):
			flush()
			out = append(out, jobHead(strings.TrimSpace(t[4:])))
		case strings.HasPrefix(t, "#### "):
			flush()
			out = append(out, `<div class="proj">`+mdInline(strings.TrimSpace(t[5:]))+"</div>")
		case strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* "):
			flushPara()
			bullets = append(bullets, "<li>"+mdInline(t[2:])+"</li>")
		case strings.HasPrefix(t, "**") && strings.HasSuffix(t, "**") && !strings.Contains(t, ":**"):
			// a project / programme sub-heading on its own line
			flush()
			out = append(out, `<div class="proj">`+mdInline(strings.Trim(t, "*"))+"</div>")
		default:
			flushBullets()
			para = append(para, mdInline(t))
		}
	}
	flush()
	return out
}

// jobHead lays out one "### Role — Company · Location · dates" line the way the source
// resume does: company and location, then the role, then the dates. A line that does not
// split that way is printed whole rather than mangled.
func jobHead(h string) string {
	e := jobEntry{Raw: h}
	role, rest, ok := strings.Cut(h, " — ")
	if !ok {
		role, rest, ok = strings.Cut(h, " – ")
	}
	if ok {
		e.Role = role
		parts := strings.Split(rest, " · ")
		e.Company = strings.TrimSpace(parts[0])
		if n := len(parts); n > 1 {
			if last := strings.TrimSpace(parts[n-1]); dateRe.MatchString(last) {
				e.Dates = last
				parts = parts[:n-1]
			}
			if len(parts) > 1 {
				e.Location = strings.TrimSpace(strings.Join(parts[1:], " · "))
			}
		}
	}
	var b strings.Builder
	b.WriteString(`<div class="job">`)
	if e.Company == "" {
		fmt.Fprintf(&b, `<div class="co"><b>%s</b></div>`, mdInline(e.Raw))
	} else {
		fmt.Fprintf(&b, `<div class="co"><b>%s</b>`, mdInline(e.Company))
		if e.Location != "" {
			fmt.Fprintf(&b, ` — %s`, mdInline(e.Location))
		}
		b.WriteString(`</div>`)
		fmt.Fprintf(&b, `<div class="role">%s</div>`, mdInline(e.Role))
		if e.Dates != "" {
			fmt.Fprintf(&b, `<div class="dates">%s</div>`, mdInline(e.Dates))
		}
	}
	b.WriteString(`</div>`)
	return b.String()
}

// resumeHTML renders the print page. The measurements are the source resume's, taken off
// the PDF in jobs/input/profile/source-resumes/: sidebar 172.5pt wide starting at the 1in
// margin, main column at 249.9pt, body 12pt/14.64pt, headings bold at the same 12pt.
func resumeHTML(d resumeDoc) string {
	var b strings.Builder
	title := d.Name
	if title == "" {
		title = "Resume"
	}
	fmt.Fprintf(&b, resumePageTop, html.EscapeString(title+" — Resume"))

	b.WriteString(`<table class="lay"><tr><td class="side">`)
	if len(d.Skills) > 0 {
		b.WriteString(`<div class="sidebg"><div class="sh">CORE SKILLS</div>`)
		for _, g := range d.Skills {
			if g.Label != "" {
				fmt.Fprintf(&b, `<div class="sg">%s</div>`, g.Label)
			}
			if g.Items != "" {
				fmt.Fprintf(&b, `<div class="si">%s</div>`, g.Items)
			}
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</td><td class="main">`)

	fmt.Fprintf(&b, `<div class="name">%s</div>`, html.EscapeString(strings.ToUpper(d.Name)))
	if d.Title != "" {
		fmt.Fprintf(&b, `<div class="jobtitle">%s</div>`, mdInline(d.Title))
	}
	if d.Contact != "" {
		fmt.Fprintf(&b, `<div class="contact">%s</div>`, d.Contact)
	}
	if len(d.Summary) > 0 {
		b.WriteString(`<div class="sec">SUMMARY</div>`)
		for _, p := range d.Summary {
			fmt.Fprintf(&b, `<p>%s</p>`, p)
		}
	}
	for _, s := range d.Sections {
		fmt.Fprintf(&b, `<div class="sec">%s</div>`, html.EscapeString(strings.ToUpper(s.Heading)))
		for _, blk := range s.Body {
			b.WriteString(blk)
		}
	}
	b.WriteString(`</td></tr></table></body></html>`)
	return b.String()
}

// resumePageTop is everything up to the layout table. Aptos first, because that is what the
// source resume is set in and what the user's own machine will have; the fallbacks are the
// nearest humanist sans on a machine without it.
const resumePageTop = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>%s</title>
<style>
@page{size:Letter;margin:1in}
html,body{margin:0;padding:0}
body{font:12pt/1.22 "Aptos","Segoe UI","Helvetica Neue",Arial,sans-serif;color:#000;
  -webkit-print-color-adjust:exact;print-color-adjust:exact}
table.lay{border-collapse:collapse;width:100%%}
td{vertical-align:top;padding:0}
td.side{width:172.5pt}
td.main{width:290pt;padding-left:5.4pt}
.sidebg{background:#e8eef5;padding:5.4pt}
.sh{font-weight:700;margin-bottom:3pt}
.sg{font-weight:700;margin-top:8pt}
.si{margin-top:1pt}
.name{font-weight:700}
.jobtitle{font-weight:700}
.contact{margin-top:0}
.sec{font-weight:700;margin:12pt 0 2pt}
.job{margin-top:10pt}
.co b{font-weight:700}
.role{font-weight:700}
.dates{font-weight:700}
.proj{font-weight:700;margin-top:8pt}
p{margin:0 0 6pt}
ul{margin:2pt 0 6pt;padding-left:18pt}
li{margin:0 0 2pt}
b,strong{font-weight:700}
h1,h2,h3{font-size:12pt;margin:0}
/* keep an entry's heading with at least the start of its bullets */
.job,.proj,.sec{break-after:avoid;page-break-after:avoid}
li{break-inside:avoid;page-break-inside:avoid}
</style></head><body>
`

// findChrome locates a Chrome-family binary that can print to PDF. $CHROME wins, then the
// usual installs, then whatever Playwright has already downloaded for the browse skill —
// so a machine set up for /browse can render a PDF without installing anything else.
func findChrome() (string, error) {
	if c := os.Getenv("CHROME"); c != "" {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	cands := []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
		"/usr/bin/google-chrome", "/usr/bin/chromium", "/usr/bin/chromium-browser",
	}
	for _, c := range cands {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	for _, pat := range []string{
		filepath.Join(os.Getenv("HOME"), "Library/Caches/ms-playwright/chromium-*/chrome-mac/Chromium.app/Contents/MacOS/Chromium"),
		filepath.Join(os.Getenv("HOME"), ".cache/ms-playwright/chromium-*/chrome-linux/chrome"),
	} {
		if hits, _ := filepath.Glob(pat); len(hits) > 0 {
			sort.Strings(hits)
			return hits[len(hits)-1], nil
		}
	}
	return "", fmt.Errorf("no Chrome found — set CHROME=/path/to/chrome, or install Google Chrome")
}

// renderResumePDF turns one resume.md into resume.pdf beside it.
func renderResumePDF(mdPath, pdfPath string) error {
	src, err := os.ReadFile(mdPath)
	if err != nil {
		return err
	}
	chrome, err := findChrome()
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp("", "resume-*.html")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(resumeHTML(parseResume(string(src)))); err != nil {
		return err
	}
	tmp.Close()

	out, err := exec.Command(chrome,
		"--headless=new", "--disable-gpu", "--no-sandbox",
		"--no-pdf-header-footer", "--run-all-compositor-stages-before-draw",
		"--virtual-time-budget=4000",
		"--print-to-pdf="+pdfPath, "file://"+tmp.Name(),
	).CombinedOutput()
	if st, serr := os.Stat(pdfPath); serr != nil || st.Size() == 0 {
		return fmt.Errorf("chrome produced no PDF: %v\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// renderAllPDFs renders every job that has a resume.md, newest work included, and reports
// one line per job. Serial: each render is a browser launch.
func renderAllPDFs(liveDir, only string) int {
	dirs, _ := filepath.Glob(filepath.Join(liveDir, "applications", "*"))
	sort.Strings(dirs)
	bad := 0
	found := 0
	for _, d := range dirs {
		id := filepath.Base(d)
		if only != "" && only != "all" && only != id {
			continue
		}
		md := filepath.Join(d, "resume.md")
		if _, err := os.Stat(md); err != nil {
			continue
		}
		found++
		pdf := filepath.Join(d, "resume.pdf")
		if err := renderResumePDF(md, pdf); err != nil {
			fmt.Printf("  ✗ %s — %v\n", id, err)
			bad++
			continue
		}
		st, _ := os.Stat(pdf)
		fmt.Printf("  ✓ %s → %s (%.0f KB)\n", id, pdf, float64(st.Size())/1024)
	}
	if found == 0 {
		fmt.Printf("  no resume.md found for %q under %s/applications\n", only, liveDir)
		return 1
	}
	return bad
}
