// Command web serves a dashboard for jobs.md.
//
// It re-reads and re-parses jobs/output/jobs.md on every request, so status changes written
// by /job-apply or /job-triage show up on refresh — no restart, no build step.
//
//	go run ./jobs/dashboard            # then open http://localhost:8080
//	go run ./jobs/dashboard -f /other/jobs.md -addr :9000
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type table struct {
	Cols []string
	Rows [][]string
}

type page struct {
	Meta     []string
	Summary  table
	Backlog  table
	Excluded table
	Stats    stats
	File     string
	Loaded   string

	// Interview-experience index (interviews/output/interviews.md). Independent of the
	// jobs index: a missing or empty file yields an empty tab, never an error.
	// Snapshots written by jobs/bin/snapshot.sh before each run, newest first. Viewing
	// one re-renders the whole page from that file; Viewing is empty on the live index.
	History []snapshot
	Viewing string

	// Stages[i] is the computed stage for Summary.Rows[i]. SumCols describes the summary
	// table's columns — Stage prepended to Summary.Cols — so the header can carry a
	// per-column filter of the right kind and a sort affordance.
	Stages  []string
	SumCols []col

	// Cells[r] holds row r's values in SumCols order — Stage first, then the visible
	// source columns. Precomputed so the template renders one flat list instead of
	// index-juggling against Summary.Cols, and so hiding a column needs no template change.
	Cells [][]string

	// ResumeJob[r] is row r's job ID when its tailored resume exists on disk, "" otherwise.
	// The Status and Resume cells become links to /doc for those rows — filled by linkDocs,
	// which stats the file rather than trusting the ✓ in the column.
	ResumeJob []string

	// ColMeta is SumCols as JSON for the filter popover, which needs each column's kind
	// and value vocabulary client-side. Emitted into a <script type="application/json">
	// so it never goes through attribute escaping.
	ColMeta template.JS

	// Blockquote warnings from the jobs.md preamble ("> ⚠️ Fit scores are provisional…").
	// Separate from Meta because they are conditions the user should act on, not facts.
	Warnings []string

	IX         table
	IXBacklog  table
	IXExcluded table
	IXStats    ixstats
	IXFile     string
}

// stageOf collapses Status + the three artifact columns into the one label a person
// actually wants: how far has this job got, and is it waiting on me? Derived rather than
// stored, so it can never disagree with the columns it is computed from.
func stageOf(status, jd, resume, form string) string {
	done := func(v string) bool { return strings.HasPrefix(strings.TrimSpace(v), "✓") }
	st := strings.ToLower(strings.TrimSpace(status))
	switch st {
	case "submitted", "interviewing", "offer", "rejected", "ghosted":
		return strings.Title(st)
	case "skipped":
		return "Skipped"
	}
	switch {
	case done(form):
		return "Ready for submission"
	case done(resume):
		return "Tailored"
	case strings.Contains(form, "progress"):
		return "Filling form"
	case done(jd):
		return "JD captured"
	case strings.Contains(jd, "manual"):
		return "Needs your JD"
	}
	return "Searched"
}

// cell returns row[i], or "" when the row is short — jobs.md rows are hand-edited and a
// truncated one must not panic the dashboard.
func cell(row []string, i int) string {
	if i < len(row) {
		return row[i]
	}
	return ""
}

// shortURL is the link text a URL cell renders as. The filter tallies the same string,
// because the filter compares against the cell's rendered text — tallying the full URL
// would build a vocabulary that can never match anything on screen.
func shortURL(s string) string {
	u := strings.TrimPrefix(strings.TrimPrefix(s, "https://"), "http://")
	if len(u) > 42 {
		return u[:42] + "…"
	}
	return u
}

// opt is one selectable value in a column's filter, with how many rows carry it.
type opt struct {
	V string `json:"v"`
	N int    `json:"n"`
}

// col describes one summary column. Kind drives sorting only ("num" compares numerically);
// every column filters the same way — a searchable multi-select over the values present.
type col struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Options []opt  `json:"options"`
}

// hiddenCols are columns kept in jobs.md but not shown in the table. "Fit" is load-bearing
// — /job auto shortlists on its bands and the Dashboard tab charts it — but it is noise in
// a row you are reading, so it stays in the data and out of the view.
var hiddenCols = map[string]bool{"Fit": true}

// maxOpts caps a column's filter list. A column of unique URLs would otherwise ship one
// checkbox per row; the search box makes a long list usable, but not an unbounded one.
const maxOpts = 300

// showStage puts the computed Stage column back in the table. Off, because Status plus the
// JD / Resume / Form columns already say the same thing in the same row — Stage was a
// fourth restatement of three columns you can read directly. The stage itself is still
// computed: the Dashboard tab and the "Needs you" panel are built from it.
const showStage = false

// buildCols returns the visible columns (Stage prepended when shown) and, for each source
// column kept, its index in the original row — so the caller can build matching cells.
func buildCols(names []string, rows [][]string, stages []string) ([]col, []int) {
	var out []col
	if showStage {
		out = append(out, col{Name: "Stage", Kind: "text", Options: tallyOpts(stages)})
	}
	var src []int
	for i, n := range names {
		if hiddenCols[n] {
			continue
		}
		kind := "text"
		if n == "#" || n == "Fit" {
			kind = "num"
		}
		var vals []string
		for _, r := range rows {
			if i < len(r) {
				vals = append(vals, r[i])
			}
		}
		out = append(out, col{Name: n, Kind: kind, Options: tallyOpts(vals)})
		src = append(src, i)
	}
	return out, src
}

// tallyOpts counts each distinct value, most-common first then alphabetical, so the values
// worth filtering on sit at the top of the list.
func tallyOpts(vals []string) []opt {
	n := map[string]int{}
	for _, v := range vals {
		v = strings.TrimSpace(tick.Replace(v))
		if v == "" || v == "—" || v == "-" {
			continue
		}
		if linkRe.MatchString(v) {
			v = shortURL(v) // match what the cell actually shows
		}
		n[v]++
	}
	out := make([]opt, 0, len(n))
	for v, c := range n {
		out = append(out, opt{v, c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].N != out[j].N {
			return out[i].N > out[j].N
		}
		return out[i].V < out[j].V
	})
	if len(out) > maxOpts {
		out = out[:maxOpts]
	}
	return out
}

// snapshot is one archived jobs.md under jobs/output/history/.
type snapshot struct {
	Name  string // file name, used as the ?v= value
	When  string // human-readable timestamp parsed out of the name
	Total int    // rows on the list at that moment, for a sense of movement
	Live  bool
}

type stat struct {
	Label string
	N     int
	Pct   float64
}

type stats struct {
	Total, Backlog, Excluded  int
	Status, Verified, Source  []stat
	Location, FitBand         []stat
	Profile                   []stat
	CompListed, CompMissing   int
	MedianFit, TopFit, LowFit int
	Actionable                int

	// Pipeline progress across the three stage columns, plus the rows the pipeline
	// deliberately skipped and handed back to the user.
	JDDone, ResumeDone, FormDone int
	NeedsJD                      int
	Blocked                      []blockedRow
}

// blockedRow is a job the auto pipeline could not advance on its own — currently only
// "JD needs a manual download", the one hand-off the pipeline makes by design.
type blockedRow struct {
	ID, Role, Company, Why, URL string
}

type ixstats struct {
	Total, Backlog, Excluded     int
	Track, CredBand              []stat
	Source, Outcome              []stat
	MedianCred, TopCred, LowCred int
	Linked, TrackA, TrackB       int
}

var (
	rowSplit = regexp.MustCompile(`\s*\|\s*`)
	sepRow   = regexp.MustCompile(`^\|[\s:|-]+\|$`)
	tick     = strings.NewReplacer("`", "")
	linkRe   = regexp.MustCompile(`^https?://\S+$`)
	boldRe   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	codeRe   = regexp.MustCompile("`([^`]+)`")
)

// parseSection pulls the first markdown pipe-table under the given "## Heading".
func parseSection(md, heading string) table {
	i := strings.Index(md, heading+"\n")
	if i < 0 {
		return table{}
	}
	rest := md[i+len(heading):]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		rest = rest[:j]
	}
	var t table
	for _, ln := range strings.Split(rest, "\n") {
		ln = strings.TrimSpace(ln)
		if !strings.HasPrefix(ln, "|") || sepRow.MatchString(ln) {
			continue
		}
		cells := rowSplit.Split(strings.Trim(ln, "|"), -1)
		for k := range cells {
			cells[k] = tick.Replace(strings.TrimSpace(cells[k]))
		}
		if t.Cols == nil {
			t.Cols = cells
			continue
		}
		// pad or trim to header width so a malformed row can't break the layout
		for len(cells) < len(t.Cols) {
			cells = append(cells, "")
		}
		t.Rows = append(t.Rows, cells[:len(t.Cols)])
	}
	return t
}

func (t table) idx(name string) int {
	for i, c := range t.Cols {
		if strings.EqualFold(c, name) {
			return i
		}
	}
	return -1
}

func (t table) col(name string) []string {
	i := t.idx(name)
	if i < 0 {
		return nil
	}
	out := make([]string, 0, len(t.Rows))
	for _, r := range t.Rows {
		out = append(out, r[i])
	}
	return out
}

func tally(vals []string, norm func(string) string) []stat {
	m := map[string]int{}
	n := 0
	for _, v := range vals {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if norm != nil {
			v = norm(v)
		}
		m[v]++
		n++
	}
	out := make([]stat, 0, len(m))
	for k, c := range m {
		p := 0.0
		if n > 0 {
			p = float64(c) * 100 / float64(n)
		}
		out = append(out, stat{k, c, p})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].N != out[j].N {
			return out[i].N > out[j].N
		}
		return out[i].Label < out[j].Label
	})
	return out
}

func locBucket(s string) string {
	l := strings.ToLower(s)
	switch {
	case strings.Contains(l, "remote"):
		return "Remote"
	case strings.Contains(l, "hybrid"):
		return "Hybrid"
	case strings.Contains(l, "onsite"), strings.Contains(l, "on-site"):
		return "Onsite"
	}
	return "Unspecified"
}

func computeStats(p *page) {
	s := stats{Total: len(p.Summary.Rows), Backlog: len(p.Backlog.Rows), Excluded: len(p.Excluded.Rows)}
	s.Status = tally(p.Summary.col("Status"), strings.ToLower)
	s.Verified = tally(p.Summary.col("Verified"), strings.ToLower)
	s.Source = tally(p.Summary.col("Source"), strings.ToLower)
	s.Location = tally(p.Summary.col("Location"), locBucket)
	s.Profile = tally(p.Summary.col("Profile"), func(v string) string {
		if v = strings.TrimSpace(v); v == "" || v == "—" || v == "-" {
			return "unassigned"
		}
		return v
	})

	for _, c := range p.Summary.col("Comp") {
		if c == "" || strings.EqualFold(c, "not listed") {
			s.CompMissing++
		} else {
			s.CompListed++
		}
	}

	var fits []int
	bands := map[string]int{}
	for _, f := range p.Summary.col("Fit") {
		n, err := strconv.Atoi(strings.TrimSpace(f))
		if err != nil {
			continue
		}
		fits = append(fits, n)
		switch {
		case n >= 80:
			bands["Apply now (80+)"]++
		case n >= 60:
			bands["Worth a look (60–79)"]++
		case n >= 40:
			bands["Stretch (40–59)"]++
		default:
			bands["Skip (<40)"]++
		}
	}
	for _, k := range []string{"Apply now (80+)", "Worth a look (60–79)", "Stretch (40–59)", "Skip (<40)"} {
		if bands[k] > 0 {
			pct := float64(bands[k]) * 100 / float64(len(fits))
			s.FitBand = append(s.FitBand, stat{k, bands[k], pct})
		}
	}
	if len(fits) > 0 {
		sort.Ints(fits)
		s.LowFit, s.TopFit = fits[0], fits[len(fits)-1]
		s.MedianFit = fits[len(fits)/2]
	}
	// Per-row stage, and the Profile vocabulary for the filter dropdown.
	jdc, rec, foc, stc := p.Summary.col("JD"), p.Summary.col("Resume"), p.Summary.col("Form"), p.Summary.col("Status")
	for i := range p.Summary.Rows {
		p.Stages = append(p.Stages, stageOf(cell(stc, i), cell(jdc, i), cell(rec, i), cell(foc, i)))
	}
	var src []int
	p.SumCols, src = buildCols(p.Summary.Cols, p.Summary.Rows, p.Stages)
	for r, row := range p.Summary.Rows {
		out := make([]string, 0, len(src)+1)
		if showStage {
			out = append(out, p.Stages[r])
		}
		for _, i := range src {
			out = append(out, cell(row, i))
		}
		p.Cells = append(p.Cells, out)
	}
	if b, err := json.Marshal(p.SumCols); err == nil {
		p.ColMeta = template.JS(b)
	}

	// Stage-column progress. The pipeline writes "✓ <date>" when an artifact exists,
	// "⏳ manual" when it gave up and handed the job back, "—" when untouched.
	done := func(v string) bool { return strings.HasPrefix(strings.TrimSpace(v), "✓") }
	for _, v := range p.Summary.col("JD") {
		if done(v) {
			s.JDDone++
		} else if strings.Contains(v, "manual") {
			s.NeedsJD++
		}
	}
	for _, v := range p.Summary.col("Resume") {
		if done(v) {
			s.ResumeDone++
		}
	}
	for _, v := range p.Summary.col("Form") {
		if done(v) {
			s.FormDone++
		}
	}

	// The "needs you" list: which specific jobs are waiting on a paste, with the URL to
	// go get it from. Counting them is not enough — the user needs the links.
	jd, ids := p.Summary.col("JD"), p.Summary.col("Job ID")
	roles, comps := p.Summary.col("Role"), p.Summary.col("Company")
	careers, applies := p.Summary.col("Careers"), p.Summary.col("Apply URL")
	at := func(c []string, i int) string {
		if i < len(c) {
			return strings.Trim(strings.TrimSpace(c[i]), "`")
		}
		return ""
	}
	for i, v := range jd {
		if !strings.Contains(v, "manual") {
			continue
		}
		url := at(careers, i)
		if url == "" || url == "—" {
			url = at(applies, i)
		}
		s.Blocked = append(s.Blocked, blockedRow{
			ID: at(ids, i), Role: at(roles, i), Company: at(comps, i),
			Why: "JD needs a manual download", URL: url,
		})
	}

	// jobs still waiting on a first action
	for _, st := range p.Summary.col("Status") {
		switch strings.ToLower(strings.TrimSpace(st)) {
		case "found", "shortlisted", "jd-captured", "tailored":
			s.Actionable++
		}
	}
	p.Stats = s
}

func computeIX(p *page) {
	s := ixstats{Total: len(p.IX.Rows), Backlog: len(p.IXBacklog.Rows), Excluded: len(p.IXExcluded.Rows)}
	s.Track = tally(p.IX.col("Track"), strings.ToLower)
	s.Source = tally(p.IX.col("Source"), strings.ToLower)
	s.Outcome = tally(p.IX.col("Outcome"), func(v string) string {
		if v == "" || v == "—" || v == "-" {
			return "not stated"
		}
		return strings.ToLower(v)
	})

	for _, t := range p.IX.col("Track") {
		switch strings.ToLower(strings.TrimSpace(t)) {
		case "a-company":
			s.TrackA++
		case "b-ai-screener":
			s.TrackB++
		}
	}
	for _, j := range p.IX.col("Job ID") {
		if j = strings.TrimSpace(tick.Replace(j)); j != "" && j != "—" && j != "-" {
			s.Linked++
		}
	}

	var creds []int
	bands := map[string]int{}
	for _, c := range p.IX.col("Cred") {
		n, err := strconv.Atoi(strings.TrimSpace(c))
		if err != nil {
			continue
		}
		creds = append(creds, n)
		switch {
		case n >= 80:
			bands["Trust it (80+)"]++
		case n >= 60:
			bands["Useful, caveats (60–79)"]++
		default:
			bands["Weak (40–59)"]++
		}
	}
	for _, k := range []string{"Trust it (80+)", "Useful, caveats (60–79)", "Weak (40–59)"} {
		if bands[k] > 0 {
			s.CredBand = append(s.CredBand, stat{k, bands[k], float64(bands[k]) * 100 / float64(len(creds))})
		}
	}
	if len(creds) > 0 {
		sort.Ints(creds)
		s.LowCred, s.TopCred = creds[0], creds[len(creds)-1]
		s.MedianCred = creds[len(creds)/2]
	}
	p.IXStats = s
}

// loadIX layers the interview index onto p. A missing file is not an error — the
// dashboard is useful with jobs alone, and the tab renders its own empty state.
func loadIX(p *page, path string) {
	if path == "" {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	md := string(b)
	p.IXFile = path
	p.IX = parseSection(md, "## Reports")
	p.IXBacklog = parseSection(md, "## Unverified backlog")
	p.IXExcluded = parseSection(md, "## Excluded")
	computeIX(p)
}

func load(path string) (*page, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	md := string(b)
	p := &page{
		Summary:  parseSection(md, "## Summary"),
		Backlog:  parseSection(md, "## Unverified backlog"),
		Excluded: parseSection(md, "## Excluded"),
		File:     path,
		Loaded:   time.Now().Format("15:04:05"),
	}
	for _, ln := range strings.Split(md, "\n") {
		if strings.HasPrefix(ln, "## ") {
			break
		}
		switch {
		case strings.HasPrefix(ln, "- **"):
			p.Meta = append(p.Meta, strings.TrimPrefix(ln, "- "))
		case strings.HasPrefix(ln, "> "):
			w := strings.TrimSpace(strings.TrimPrefix(ln, ">"))
			if w == "" {
				continue
			}
			// A wrapped blockquote continues the previous warning.
			if n := len(p.Warnings); n > 0 && !strings.HasPrefix(w, "⚠️") {
				p.Warnings[n-1] += " " + w
			} else {
				p.Warnings = append(p.Warnings, w)
			}
		case strings.HasPrefix(ln, "  ") && strings.TrimSpace(ln) != "" && len(p.Meta) > 0:
			// continuation of the previous wrapped bullet
			p.Meta[len(p.Meta)-1] += " " + strings.TrimSpace(ln)
		}
	}
	computeStats(p)
	return p, nil
}

var funcs = template.FuncMap{
	"isURL": func(s string) bool { return linkRe.MatchString(s) },
	"short": shortURL,
	"slug": func(s string) string {
		return strings.ToLower(strings.NewReplacer(" ", "-", "(", "", ")", "", "+", "plus", "–", "-", "<", "lt").Replace(s))
	},
	"pct": func(f float64) string { return fmt.Sprintf("%.0f", f) },
	// docCSS styles the document drawer; it lives in docs.go so the standalone /doc page
	// and the drawer share one stylesheet.
	"docCSS": func() template.CSS { return template.CSS(docCSS) },
	"credClass": func(s string) string {
		n, _ := strconv.Atoi(strings.TrimSpace(s))
		switch {
		case n >= 80:
			return "c-hi"
		case n >= 60:
			return "c-mid"
		}
		return "c-lo"
	},
	"markdn": func(s string) template.HTML {
		e := template.HTMLEscapeString(s)
		e = boldRe.ReplaceAllString(e, "<strong>$1</strong>")
		e = codeRe.ReplaceAllString(e, "<code>$1</code>")
		return template.HTML(e)
	},
}

// scanHistory lists the snapshots beside the live index, newest first. A history
// directory that is missing or unreadable is not an error — it just means no run has
// snapshotted yet.
func scanHistory(liveDir string) []snapshot {
	dir := filepath.Join(liveDir, "history")
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []snapshot
	for _, e := range ents {
		n := e.Name()
		if e.IsDir() || !strings.HasPrefix(n, "jobs-") || !strings.HasSuffix(n, ".md") {
			continue
		}
		ts := strings.TrimSuffix(strings.TrimPrefix(n, "jobs-"), ".md")
		when := ts
		if t, err := time.Parse("2006-01-02T15-04-05", ts); err == nil {
			when = t.Format("Mon 2 Jan, 15:04")
		}
		total := 0
		if b, err := os.ReadFile(filepath.Join(dir, n)); err == nil {
			total = len(parseSection(string(b), "## Summary").Rows)
		}
		out = append(out, snapshot{Name: n, When: when, Total: total})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name > out[j].Name })
	return out
}

// resolveVersion maps a ?v= value to a file on disk. It rejects anything that isn't a
// plain snapshot file name, so the query parameter can never walk out of the history dir.
func resolveVersion(liveDir, v string) (string, bool) {
	if v == "" || v == "live" {
		return "", false
	}
	if v != filepath.Base(v) || !strings.HasPrefix(v, "jobs-") || !strings.HasSuffix(v, ".md") {
		return "", false
	}
	full := filepath.Join(liveDir, "history", v)
	if _, err := os.Stat(full); err != nil {
		return "", false
	}
	return full, true
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	file := flag.String("f", "", "path to jobs.md (default: ./jobs/output/jobs.md)")
	ixFile := flag.String("i", "", "path to interviews.md (default: ./interviews/output/interviews.md)")
	pdf := flag.String("pdf", "", "render resume.pdf for a job id (or \"all\") and exit, instead of serving")
	flag.Parse()

	path := *file
	if path == "" {
		for _, c := range []string{
			filepath.Join("jobs", "output", "jobs.md"),
			filepath.Join("..", "output", "jobs.md"),
			filepath.Join("..", "..", "jobs", "output", "jobs.md"),
		} {
			if _, err := os.Stat(c); err == nil {
				path = c
				break
			}
		}
	}
	if path == "" {
		log.Fatal("jobs/output/jobs.md not found — run from the project root or jobs/dashboard/, or pass -f <path>")
	}
	abs, _ := filepath.Abs(path)

	ixPath := *ixFile
	if ixPath == "" {
		for _, c := range []string{
			filepath.Join("interviews", "output", "interviews.md"),
			filepath.Join("..", "..", "interviews", "output", "interviews.md"),
		} {
			if _, err := os.Stat(c); err == nil {
				ixPath = c
				break
			}
		}
	}
	ixAbs := ""
	if ixPath != "" {
		ixAbs, _ = filepath.Abs(ixPath)
	}

	if *pdf != "" {
		fmt.Printf("rendering resumes in the source-resume format → %s/applications/<job-id>/resume.pdf\n", filepath.Dir(abs))
		if renderAllPDFs(filepath.Dir(abs), *pdf) > 0 {
			os.Exit(1)
		}
		return
	}

	tpl := template.Must(template.New("page").Funcs(funcs).Parse(pageHTML))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		liveDir := filepath.Dir(abs)
		src, viewing := abs, ""
		if f, ok := resolveVersion(liveDir, r.URL.Query().Get("v")); ok {
			src, viewing = f, filepath.Base(f)
		}
		p, err := load(src)
		if err != nil {
			http.Error(w, "could not read "+src+": "+err.Error(), 500)
			return
		}
		p.Viewing = viewing
		p.History = scanHistory(liveDir)
		linkDocs(p, liveDir)
		loadIX(p, ixAbs)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tpl.Execute(w, p); err != nil {
			log.Println("render:", err)
		}
	})

	http.HandleFunc("/doc", docHandler(filepath.Dir(abs)))

	if ixAbs == "" {
		log.Printf("interviews index not found — the Interviews tab will be empty (run /interview-search)")
	}
	log.Printf("jobs dashboard → http://localhost%s   (jobs: %s)", *addr, abs)
	if ixAbs != "" {
		log.Printf("  interviews: %s", ixAbs)
	}
	log.Fatal(http.ListenAndServe(*addr, nil))
}
