package main

// Resume layout, read from jobs/input/config/resume-format.md.
//
// A tailored resume only changes words. Where they sit on the page comes from the resume the
// user dropped into jobs/input/profile/source-resumes/, and `/job setup` records that layout in
// resume-format.md — page size, font, margins, section headings, date style. Nothing about any
// one person's resume is baked in here: without the file you get a plain single-column page,
// and `Layout: sidebar` selects the two-column design in resume.go.

import (
	"encoding/base64"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type resumeFormat struct {
	Layout        string // "single-column" (default) or "sidebar"
	Page          string // CSS @page size: "A4" or "Letter"
	Font          string
	Size          float64 // body font size, pt
	NameSize      float64
	LineHeight    float64
	MarginTop     string
	MarginSide    string
	MarginBottom  string
	Justify       bool
	ShowHeadline  bool
	ContactOrder  []string // "phone", "email", "other"
	ContactSep    string
	Months        []string
	LocationComma bool   // keep the comma in "Pittsburgh, PA"
	EarlierLine   string // how an "Earlier career — <companies>" entry names its employers; {companies}
	Headings      map[string]string
	SkillsTable   bool // bordered one-row-per-group table (else plain lines)
}

var resumeFmt = defaultResumeFormat()

func defaultResumeFormat() resumeFormat {
	return resumeFormat{
		Layout: "single-column", Page: "Letter", Font: "Calibri", Size: 10.5, NameSize: 12, LineHeight: 1.2,
		MarginTop: "0.75in", MarginSide: "0.75in", MarginBottom: "0.75in",
		Justify: true, ShowHeadline: false,
		ContactOrder: []string{"phone", "email", "other"}, ContactSep: " | ",
		Months:        strings.Fields("Jan Feb Mar Apr May Jun Jul Aug Sep Oct Nov Dec"),
		LocationComma: true,
		EarlierLine:   "Career progression included positions with {companies}.",
		SkillsTable:   true,
		Headings:      map[string]string{},
	}
}

// Matches both "- **Key:** value" and "- Key: value".
var fmtLineRe = regexp.MustCompile(`^\s*[-*]\s*\*{0,2}([^:*]+?)(?::\*{0,2}|\*{0,2}:)\s*(.*?)\s*$`)

// loadResumeFormat reads "- Key: value" lines. Unknown keys are ignored and a missing file
// leaves the defaults, so an old workspace keeps rendering.
func loadResumeFormat(path string) resumeFormat {
	f := defaultResumeFormat()
	b, err := os.ReadFile(path)
	if err != nil {
		return f
	}
	num := func(v string, into *float64) {
		if x, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(v), "pt"), 64); err == nil {
			*into = x
		}
	}
	yes := func(v string) bool { v = strings.ToLower(v); return strings.HasPrefix(v, "y") || v == "true" }
	unq := func(v string) string { return strings.Trim(strings.TrimSpace(strings.TrimLeft(v, "* ")), "`\"") }
	for _, ln := range strings.Split(commentRe.ReplaceAllString(string(b), ""), "\n") {
		m := fmtLineRe.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		k, v := strings.ToLower(strings.TrimSpace(m[1])), unq(m[2])
		switch {
		case k == "layout":
			f.Layout = strings.ToLower(v)
		case k == "page":
			f.Page = v
		case k == "font":
			f.Font = v
		case k == "size":
			num(v, &f.Size)
		case k == "name size":
			num(v, &f.NameSize)
		case k == "line height":
			num(v, &f.LineHeight)
		case k == "margin top":
			f.MarginTop = v
		case k == "margin sides":
			f.MarginSide = v
		case k == "margin bottom":
			f.MarginBottom = v
		case k == "justify":
			f.Justify = yes(v)
		case k == "show headline":
			f.ShowHeadline = yes(v)
		case k == "contact order":
			f.ContactOrder = nil
			for _, p := range strings.Split(v, ",") {
				f.ContactOrder = append(f.ContactOrder, strings.ToLower(strings.TrimSpace(p)))
			}
		case k == "contact separator":
			f.ContactSep = strings.Trim(m[2], "`\"")
		case k == "months":
			if ms := strings.Fields(v); len(ms) == 12 {
				f.Months = ms
			}
		case k == "location comma":
			f.LocationComma = yes(v)
		case k == "earlier career line":
			f.EarlierLine = v
		case k == "skills table":
			f.SkillsTable = yes(v)
		case strings.HasPrefix(k, "heading "):
			f.Headings[strings.TrimSpace(strings.TrimPrefix(k, "heading "))] = v
		}
	}
	return f
}

// heading maps a resume.md section name to the label the source resume prints. Keys are
// matched on a word the section name contains: summary, skills, experience, education,
// certifications, projects, earlier career.
func (f resumeFormat) heading(name string) string {
	low := strings.ToLower(name)
	for _, k := range []string{"earlier career", "summary", "skill", "experience", "education", "certification", "project", "publication"} {
		if strings.Contains(low, k) {
			for key, v := range f.Headings {
				if strings.HasPrefix(key, k) {
					return v
				}
			}
		}
	}
	return name + ":"
}

var (
	phoneRe = regexp.MustCompile(`\+?\d[\d\s().-]{7,}\d`)
	emailRe = regexp.MustCompile(`[\w.%+-]+@[\w.-]+\.[A-Za-z]{2,}`)
	ymRe    = regexp.MustCompile(`\b(\d{4})-(\d{2})\b`)
)

func (f resumeFormat) contact(raw string) string {
	var parts []string
	for _, p := range regexp.MustCompile(`\s*(?:·|\||&nbsp;·&nbsp;)\s*`).Split(raw, -1) {
		if p = strings.TrimSpace(p); p != "" {
			parts = append(parts, p)
		}
	}
	kind := func(p string) string {
		switch {
		case emailRe.MatchString(p):
			return "email"
		case phoneRe.MatchString(p):
			return "phone"
		}
		return "other"
	}
	var out []string
	for _, want := range f.ContactOrder {
		for _, p := range parts {
			if kind(p) == want {
				out = append(out, p)
			}
		}
	}
	return strings.Join(out, html.EscapeString(f.ContactSep))
}

func (f resumeFormat) dates(s string) string {
	return ymRe.ReplaceAllStringFunc(s, func(m string) string {
		p := ymRe.FindStringSubmatch(m)
		if mo, err := strconv.Atoi(p[2]); err == nil && mo >= 1 && mo <= 12 {
			return f.Months[mo-1] + " " + p[1]
		}
		return m
	})
}

// fontFaces finds the font's own files when the system doesn't have it registered — Office
// installs Calibri, Cambria etc. inside its app bundles, where Chrome can't see them — and
// points @font-face at them. Windows-style names: calibri.ttf, calibrib (bold), calibrii
// (italic), calibriz (bold italic). Found nothing → the CSS fallback stack applies.
func fontFaces(family string) string {
	base := strings.ToLower(strings.ReplaceAll(family, " ", ""))
	dirs := []string{"/Library/Fonts", filepath.Join(os.Getenv("HOME"), "Library/Fonts"),
		"/Applications/Microsoft Word.app/Contents/Resources/DFonts",
		"/Applications/Microsoft PowerPoint.app/Contents/Resources/DFonts",
		"/Applications/Microsoft Outlook.app/Contents/Resources/DFonts",
		"/usr/share/fonts/truetype/msttcorefonts", "C:/Windows/Fonts"}
	var b strings.Builder
	for _, v := range []struct{ suf, weight, style string }{{"", "400", "normal"}, {"b", "700", "normal"}, {"i", "400", "italic"}, {"z", "700", "italic"}} {
		for _, d := range dirs {
			hits, _ := filepath.Glob(filepath.Join(d, "*"))
			var got string
			for _, h := range hits {
				n := strings.ToLower(filepath.Base(h))
				if n == base+v.suf+".ttf" || n == base+v.suf+".otf" {
					got = h
					break
				}
			}
			if got != "" {
				raw, err := os.ReadFile(got)
				if err != nil {
					continue
				}
				mime := "font/ttf"
				if strings.HasSuffix(strings.ToLower(got), ".otf") {
					mime = "font/otf"
				}
				// Inlined, so it works from a file:// page, the dashboard, and paths with spaces.
				fmt.Fprintf(&b, "@font-face{font-family:%q;src:url(data:%s;base64,%s);font-weight:%s;font-style:%s}\n",
					family, mime, base64.StdEncoding.EncodeToString(raw), v.weight, v.style)
				break
			}
		}
	}
	return b.String()
}

// resumeHTMLSingle lays a tailored resume out the way a conventional single-column resume
// is set: centred name and contact, bold "Heading:" sections, a bordered skills table,
// "Company, Location" with dates flush right and the title beneath.
func resumeHTMLSingle(d resumeDoc, f resumeFormat) string {
	var b strings.Builder
	align := "left"
	if f.Justify {
		align = "justify"
	}
	fmt.Fprintf(&b, `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>%s</title><style>
%s@page{size:%s;margin:%s %s %s %s}
html,body{margin:0;padding:0}
body{font:%.1fpt/%.2f %q,"Carlito","Segoe UI","Helvetica Neue",Arial,sans-serif;color:#000;
  -webkit-print-color-adjust:exact;print-color-adjust:exact}
.name{text-align:center;font-weight:700;font-size:%.1fpt}
.contact,.headline{text-align:center}
.headline{font-weight:700}
.sec{font-weight:700;margin:13pt 0 6pt;break-after:avoid;page-break-after:avoid}
p{margin:0 0 6pt;text-align:%s}
table.sk{border-collapse:collapse;width:100%%}
table.sk td{border:0.75pt solid #000;padding:0.5pt 4pt;vertical-align:top}
.sk-line{margin:0 0 2pt}
.job{margin-top:10pt;break-after:avoid;page-break-after:avoid}
.jl{display:flex;justify-content:space-between;font-weight:700}
.role{font-weight:700;margin-bottom:6pt}
ul{margin:4pt 0 6pt;padding-left:0.55in}
li{margin:0 0 2pt;text-align:%s;padding-left:4pt;break-inside:avoid;page-break-inside:avoid}
.line{margin:0}
b,strong{font-weight:700}
</style></head><body>
`, html.EscapeString(d.Name+" — Resume"), fontFaces(f.Font), f.Page, f.MarginTop, f.MarginSide, f.MarginBottom, f.MarginSide,
		f.Size, f.LineHeight, f.Font, f.NameSize, align, align)

	fmt.Fprintf(&b, `<div class="name">%s</div>`, html.EscapeString(d.Name))
	if c := f.contact(d.Contact); c != "" {
		fmt.Fprintf(&b, `<div class="contact">%s</div>`, c)
	}
	if f.ShowHeadline && d.Title != "" {
		fmt.Fprintf(&b, `<div class="headline">%s</div>`, mdInline(d.Title))
	}
	if len(d.Summary) > 0 {
		fmt.Fprintf(&b, `<div class="sec">%s</div>`, html.EscapeString(f.heading("summary")))
		for _, p := range d.Summary {
			fmt.Fprintf(&b, `<p>%s</p>`, p)
		}
	}
	if len(d.Skills) > 0 {
		fmt.Fprintf(&b, `<div class="sec">%s</div>`, html.EscapeString(f.heading("skills")))
		if f.SkillsTable {
			b.WriteString(`<table class="sk">`)
		}
		for _, g := range d.Skills {
			cell := g.Items
			if g.Label != "" {
				cell = "<b>" + g.Label + ":</b> " + g.Items
			}
			if f.SkillsTable {
				fmt.Fprintf(&b, `<tr><td>%s</td></tr>`, cell)
			} else {
				fmt.Fprintf(&b, `<div class="sk-line">%s</div>`, cell)
			}
		}
		if f.SkillsTable {
			b.WriteString(`</table>`)
		}
	}
	last := ""
	for _, s := range d.Sections {
		label := f.heading(s.Heading)
		if label != last { // two sections sharing a label ("Education & Certifications:") print once
			fmt.Fprintf(&b, `<div class="sec">%s</div>`, html.EscapeString(label))
			last = label
		}
		low := strings.ToLower(s.Heading)
		plain := strings.Contains(low, "educ") || strings.Contains(low, "cert")
		b.WriteString(singleBody(s.Raw, f, plain, &last))
	}
	b.WriteString(`</body></html>`)
	return b.String()
}

var layoutBoldRe = regexp.MustCompile(`\*\*(.+?)\*\*`)

// singleBody renders one section. plain sections (education, certifications) print each
// entry as its own unbulleted line, as most single-column resumes do.
func singleBody(body []string, f resumeFormat, plain bool, last *string) string {
	var out, bullets, para []string
	var earlier *struct {
		companies string
		text      []string
	}
	flush := func() {
		if len(bullets) > 0 {
			out = append(out, "<ul>"+strings.Join(bullets, "")+"</ul>")
			bullets = nil
		}
		if len(para) > 0 {
			out = append(out, "<p>"+strings.Join(para, " ")+"</p>")
			para = nil
		}
	}
	endEarlier := func() {
		if earlier == nil {
			return
		}
		lbl := f.heading("earlier career")
		out = append(out, `<div class="sec">`+html.EscapeString(lbl)+`</div>`)
		*last = lbl
		txt := strings.Join(earlier.text, " ")
		if earlier.companies != "" && f.EarlierLine != "" {
			txt += " " + html.EscapeString(strings.ReplaceAll(f.EarlierLine, "{companies}", earlier.companies))
		}
		out = append(out, "<p>"+txt+"</p>")
		earlier = nil
	}
	for _, ln := range body {
		t := strings.TrimSpace(ln)
		switch {
		case t == "", t == "---", t == "***", t == "___":
			flush()
		case strings.HasPrefix(t, "### "):
			flush()
			endEarlier()
			h := strings.TrimSpace(t[4:])
			if strings.HasPrefix(strings.ToLower(h), "earlier career") {
				_, rest, _ := strings.Cut(h, "—")
				rest = strings.TrimSpace(regexp.MustCompile(`\s*·.*$`).ReplaceAllString(rest, ""))
				earlier = &struct {
					companies string
					text      []string
				}{companies: rest}
				continue
			}
			out = append(out, singleJobHead(h, f))
		case earlier != nil && (strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ")):
			earlier.text = append(earlier.text, mdInline(t[2:]))
		case plain:
			t = strings.TrimPrefix(strings.TrimPrefix(t, "- "), "* ")
			out = append(out, `<div class="line">`+mdInline(layoutBoldRe.ReplaceAllString(t, "$1"))+`</div>`)
		case strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* "):
			if len(para) > 0 {
				out = append(out, "<p>"+strings.Join(para, " ")+"</p>")
				para = nil
			}
			bullets = append(bullets, "<li>"+mdInline(t[2:])+"</li>")
		default:
			para = append(para, mdInline(t))
		}
	}
	flush()
	endEarlier()
	return strings.Join(out, "")
}

// singleJobHead: "### Role — Company · Location · dates" → "Company, Location" bold with the
// dates flush right, then the role in bold on its own line.
func singleJobHead(h string, f resumeFormat) string {
	role, rest, ok := strings.Cut(h, " — ")
	if !ok {
		role, rest, ok = strings.Cut(h, " – ")
	}
	if !ok {
		return `<div class="job"><div class="jl"><span>` + mdInline(h) + `</span></div></div>`
	}
	parts := strings.Split(rest, " · ")
	company, loc, dates := strings.TrimSpace(parts[0]), "", ""
	if n := len(parts); n > 1 {
		if lastp := strings.TrimSpace(parts[n-1]); dateRe.MatchString(lastp) {
			dates = f.dates(lastp)
			parts = parts[:n-1]
		}
		if len(parts) > 1 {
			loc = strings.TrimSpace(strings.Join(parts[1:], " "))
		}
	}
	if !f.LocationComma {
		loc = strings.ReplaceAll(loc, ",", "")
	}
	left := company
	if loc != "" {
		left += ", " + loc
	}
	return `<div class="job"><div class="jl"><span>` + mdInline(left) + `</span><span>` + mdInline(dates) +
		`</span></div><div class="role">` + mdInline(strings.TrimSpace(role)) + `</div></div>`
}
