package main

const pageHTML = `<!doctype html>
<html lang="en"><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Jobs</title>
<style>
:root{
  --bg:#fbfbfa; --panel:#fff; --ink:#1c1c1a; --muted:#6b6b66; --line:#e6e5e1;
  --accent:#3b5bdb; --ok:#2f7a4d; --warn:#a86a12; --bad:#b3402f; --chip:#f2f1ed;
}
@media (prefers-color-scheme:dark){:root:not([data-theme=light]){
  --bg:#16161a; --panel:#1e1e23; --ink:#eceae5; --muted:#9a978f; --line:#2e2e35;
  --accent:#8ea2ff; --ok:#6cc08b; --warn:#d9a441; --bad:#e0705d; --chip:#26262c;}}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--ink);
  font:14px/1.55 ui-sans-serif,-apple-system,"Segoe UI",Inter,system-ui,sans-serif}
.wrap{max-width:1400px;margin:0 auto;padding:28px 22px 60px}
header{display:flex;align-items:baseline;gap:14px;flex-wrap:wrap;margin-bottom:4px}
h1{font-size:21px;margin:0;letter-spacing:-.01em}
.sub{color:var(--muted);font-size:12.5px}
.meta{color:var(--muted);font-size:12.5px;margin:10px 0 20px;max-width:900px}
.meta div{margin:2px 0}
nav{display:flex;gap:4px;border-bottom:1px solid var(--line);margin-bottom:22px}
nav button{appearance:none;background:none;border:0;border-bottom:2px solid transparent;
  color:var(--muted);font:inherit;font-weight:500;padding:9px 15px;cursor:pointer;margin-bottom:-1px}
nav button:hover{color:var(--ink)}
nav button[aria-selected=true]{color:var(--accent);border-bottom-color:var(--accent)}
nav .count{display:inline-block;background:var(--chip);border-radius:9px;padding:0 6px;
  margin-left:6px;font-size:11px;color:var(--muted)}
.scroll{overflow-x:auto;border:1px solid var(--line);border-radius:9px;background:var(--panel)}
table{border-collapse:collapse;width:100%;font-size:13px}
th,td{text-align:left;padding:9px 12px;border-bottom:1px solid var(--line);vertical-align:top}
th{font-size:11px;text-transform:uppercase;letter-spacing:.05em;color:var(--muted);
  font-weight:600;white-space:nowrap;background:var(--chip);position:sticky;top:0}
tbody tr:last-child td{border-bottom:0}
tbody tr:hover{background:var(--chip)}
td.num{text-align:right;font-variant-numeric:tabular-nums;width:1%}
a{color:var(--accent)}
.pill{display:inline-block;padding:1px 8px;border-radius:20px;font-size:11.5px;
  white-space:nowrap;background:var(--chip);color:var(--muted);font-weight:500}
.s-found{background:color-mix(in srgb,var(--accent) 15%,transparent);color:var(--accent)}
.s-shortlisted,.s-jd-captured,.s-tailored{background:color-mix(in srgb,var(--warn) 18%,transparent);color:var(--warn)}
.s-filled-awaiting-user{background:color-mix(in srgb,var(--warn) 26%,transparent);color:var(--warn);font-weight:600}
.s-submitted,.s-interviewing,.s-offer{background:color-mix(in srgb,var(--ok) 18%,transparent);color:var(--ok)}
.s-rejected,.s-ghosted,.s-skipped{background:color-mix(in srgb,var(--bad) 15%,transparent);color:var(--bad)}
.v-verified-live{color:var(--ok)} .v-likely-live{color:var(--warn)} .v-unverified{color:var(--muted)}
.fit{font-weight:600;font-variant-numeric:tabular-nums}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(215px,1fr));gap:14px;margin-bottom:22px}
.card{background:var(--panel);border:1px solid var(--line);border-radius:9px;padding:15px 16px}
.card h3{margin:0 0 10px;font-size:11px;text-transform:uppercase;letter-spacing:.05em;
  color:var(--muted);font-weight:600}
.big{font-size:30px;font-weight:600;letter-spacing:-.02em;line-height:1.1}
.big small{font-size:12.5px;color:var(--muted);font-weight:400;letter-spacing:0}
.bar{display:flex;justify-content:space-between;gap:10px;align-items:center;margin:7px 0;font-size:12.5px}
.bar .lbl{flex:0 0 auto;max-width:58%;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.bar .track{flex:1;height:6px;background:var(--chip);border-radius:4px;overflow:hidden}
.bar .fill{height:100%;background:var(--accent);border-radius:4px}
.bar .n{flex:0 0 auto;color:var(--muted);font-variant-numeric:tabular-nums;min-width:44px;text-align:right}
.note{color:var(--muted);font-size:12px;margin-top:8px}
.empty{padding:34px;text-align:center;color:var(--muted)}
.t-a-company{background:color-mix(in srgb,var(--accent) 15%,transparent);color:var(--accent)}
.t-b-ai-screener{background:color-mix(in srgb,var(--warn) 20%,transparent);color:var(--warn)}
.c-hi{color:var(--ok)} .c-mid{color:var(--warn)} .c-lo{color:var(--muted)}
.s-offer{background:color-mix(in srgb,var(--ok) 18%,transparent);color:var(--ok)}
.s-not-stated{color:var(--muted)}
[hidden]{display:none!important}

.card.needs{border-color:#b45309;background:#fffbeb;margin-bottom:14px}
.card.needs h3{color:#b45309}
.card.needs .note{margin:2px 0 10px}
table.mini{width:100%;border-collapse:collapse;font-size:13px}
table.mini td{padding:4px 8px 4px 0;border-bottom:1px solid #f1e7d0;vertical-align:top}
table.mini tr:last-child td{border-bottom:0}
table.mini .why{color:#92400e}
@media (prefers-color-scheme:dark){
  .card.needs{border-color:#a16207;background:#231a06}
  .card.needs h3{color:#fbbf24}
  table.mini td{border-bottom-color:#3a2f14}
  table.mini .why{color:#fbbf24}
}

.warns{margin:10px 0 0;padding:9px 12px;border-left:3px solid #b45309;background:#fffbeb;
  border-radius:0 4px 4px 0;font-size:13px;line-height:1.55}
.warns div+div{margin-top:5px}
@media (prefers-color-scheme:dark){.warns{background:#231a06;border-left-color:#a16207}}

.viewing{margin:10px 0 0;padding:8px 12px;border-left:3px solid #6366f1;background:#eef2ff;
  border-radius:0 4px 4px 0;font-size:13px}
@media (prefers-color-scheme:dark){
  .viewing{background:#1e1b3a;border-left-color:#818cf8}
}

tr.hide{display:none}
.stage{display:inline-block;white-space:nowrap;padding:2px 8px;border-radius:11px;
  font-size:11px;font-weight:600;border:1px solid transparent}
.g-searched{background:#f4f4f5;color:#52525b;border-color:#e4e4e7}
.g-needs-your-jd{background:#fffbeb;color:#b45309;border-color:#fcd34d}
.g-jd-captured{background:#eff6ff;color:#1d4ed8;border-color:#bfdbfe}
.g-tailored{background:#f0fdf4;color:#15803d;border-color:#bbf7d0}
.g-filling-form{background:#eef2ff;color:#4338ca;border-color:#c7d2fe}
.g-ready-for-submission{background:#15803d;color:#fff;border-color:#15803d}
.g-submitted{background:#334155;color:#fff;border-color:#334155}
@media (prefers-color-scheme:dark){
  .filters select,.filters input,.filters button{background:#18181b;border-color:#3f3f46;color:#e4e4e7}
  .filters button:hover{background:#27272a}
  .g-searched{background:#27272a;color:#a1a1aa;border-color:#3f3f46}
  .g-needs-your-jd{background:#231a06;color:#fbbf24;border-color:#a16207}
  .g-jd-captured{background:#11203a;color:#93c5fd;border-color:#1e40af}
  .g-tailored{background:#0d2117;color:#86efac;border-color:#166534}
  .g-filling-form{background:#1e1b3a;color:#a5b4fc;border-color:#4338ca}
}

/* one sticky header row: the label sorts, the funnel opens a filter popover */
#jobs thead th{position:sticky;top:0;background:#fafafa;z-index:2;white-space:nowrap}
#jobs .hlab{cursor:pointer;user-select:none}
#jobs .hlab:hover{text-decoration:underline}
#jobs th.sorted{color:#1d4ed8}
.sar{display:inline-block;width:10px;font-size:10px;color:#1d4ed8}
.fbtn{border:0;background:none;padding:0 0 0 3px;cursor:pointer;color:#c4c4c8;
  line-height:1;vertical-align:middle;border-radius:3px}
.fbtn:hover{color:#52525b}
#jobs th.filtered .fbtn,#jobs th.open .fbtn{color:#1d4ed8}
#jobs th.filtered{color:#1d4ed8}
#jobs th.filtered .hlab::after{content:"•";margin-left:3px;color:#1d4ed8}

#fpop{position:fixed;z-index:50;width:268px;background:#fff;border:1px solid #d4d4d8;
  border-radius:10px;box-shadow:0 10px 30px rgba(0,0,0,.15);padding:12px;font-size:13px}
.fpop-h{display:flex;justify-content:space-between;align-items:baseline;margin:0 0 9px}
#fpop-t{font-weight:700;font-size:14px;color:#18181b}
#fpop-x{border:0;background:none;font:inherit;color:#2563eb;cursor:pointer;padding:0}
#fpop-x:hover{text-decoration:underline}
#fpop-q{font:inherit;width:100%;padding:7px 9px;border:1px solid #d4d4d8;border-radius:7px;
  background:#fff;color:inherit;box-sizing:border-box}
#fpop-q:focus{outline:2px solid #bfdbfe;border-color:#93c5fd}
.fpop-a{display:flex;justify-content:space-between;margin:9px 1px 6px}
.fpop-a button{border:0;background:none;font:inherit;font-size:12.5px;color:#2563eb;
  cursor:pointer;padding:0}
.fpop-a button:hover{text-decoration:underline}
.fpop-l{max-height:238px;overflow-y:auto;margin:0 -4px;padding:0 4px}
.fopt{display:flex;align-items:center;gap:9px;padding:5px 5px;border-radius:6px;cursor:pointer}
.fopt:hover{background:#f4f4f5}
.fopt input{margin:0;flex:0 0 auto;accent-color:#2563eb;width:15px;height:15px}
.fopt-t{flex:1 1 auto;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.fopt-n{flex:0 0 auto;color:#9ca3af;font-size:12px;font-variant-numeric:tabular-nums}
.fopt-e{padding:14px 5px;color:#9ca3af;text-align:center}
.fpop-f{display:flex;gap:8px;margin-top:11px;padding-top:11px;border-top:1px solid #e4e4e7}
.fpop-f button{flex:1;font:inherit;padding:7px;border:1px solid #d4d4d8;border-radius:7px;
  background:#fff;cursor:pointer;color:#3f3f46}
.fpop-f button:hover{background:#f4f4f5}
.fpop-f .prim{background:#2563eb;border-color:#2563eb;color:#fff;font-weight:600}
.fpop-f .prim:hover{background:#1d4ed8}

.tbar{display:flex;justify-content:flex-end;align-items:center;gap:10px;margin:0 0 6px;
  font-size:12px;color:#9ca3af}
.tbar button{font:inherit;padding:3px 9px;border:1px solid #d4d4d8;border-radius:5px;
  background:#fff;cursor:pointer;color:#6b7280}
.tbar button:hover{background:#f4f4f5}
tr.hide{display:none}
@media (prefers-color-scheme:dark){
  #jobs thead th{background:#141416}
  #jobs th.sorted,#jobs th.filtered,.sar,#jobs th.filtered .hlab::after,
  #jobs th.filtered .fbtn,#jobs th.open .fbtn{color:#93c5fd}
  .fbtn{color:#52525b} .fbtn:hover{color:#a1a1aa}
  #fpop{background:#18181b;border-color:#3f3f46;box-shadow:0 10px 30px rgba(0,0,0,.55)}
  #fpop-t{color:#fafafa}
  #fpop-x,.fpop-a button{color:#60a5fa}
  #fpop-q{background:#111113;border-color:#3f3f46;color:#e4e4e7}
  .fopt:hover{background:#27272a}
  .fopt-n,.fopt-e{color:#71717a}
  .fpop-f{border-top-color:#3f3f46}
  .fpop-f button{background:#18181b;border-color:#3f3f46;color:#d4d4d8}
  .fpop-f button:hover{background:#27272a}
  .fpop-f .prim{background:#2563eb;border-color:#2563eb;color:#fff}
  .tbar button{background:#18181b;border-color:#3f3f46;color:#a1a1aa}
  .tbar button:hover{background:#27272a}
}

/* a linked cell keeps its pill colour — the link is the affordance, not a blue word */
td a.doc{text-decoration:none;cursor:pointer}
td a.doc:hover{text-decoration:underline}

/* document drawer: the tailored resume, read without leaving the table */
#drawer{position:fixed;inset:0;z-index:60}
#dr-back{position:absolute;inset:0;background:rgba(20,20,22,.42)}
#dr-panel{position:absolute;top:0;right:0;bottom:0;width:min(760px,94vw);background:var(--bg);
  border-left:1px solid var(--line);box-shadow:-14px 0 40px rgba(0,0,0,.18);
  display:flex;flex-direction:column}
#dr-bar{display:flex;align-items:center;gap:10px;padding:10px 16px;border-bottom:1px solid var(--line);
  background:var(--panel)}
#dr-bar .sp{flex:1}
#dr-bar a,#dr-bar button{font:inherit;font-size:12.5px;color:var(--muted);background:none;
  border:0;cursor:pointer;text-decoration:none;padding:3px 6px;border-radius:5px}
#dr-bar a:hover,#dr-bar button:hover{color:var(--ink);background:var(--chip)}
#dr-x{font-size:15px!important;line-height:1}
#dr-body{overflow-y:auto;padding:18px 22px 60px}
#dr-body .doc-h{margin-bottom:16px}
/*DOCCSS*/{{docCSS}}
</style></head><body><div class="wrap">

<header>
  <h1>Jobs</h1>
</header>
{{if .Viewing}}<div class="viewing">Viewing a snapshot from <strong>{{.Viewing}}</strong> — this is history, not the current
  index. <a href="/">Back to live</a></div>{{end}}

<nav role="tablist">
  <button role="tab" aria-selected="true"  data-t="summary">Summary<span class="count">{{len .Summary.Rows}}</span></button>
  <button role="tab" aria-selected="false" data-t="backlog">Unverified backlog<span class="count">{{len .Backlog.Rows}}</span></button>
  <button role="tab" aria-selected="false" data-t="ix">Interviews<span class="count">{{len .IX.Rows}}</span></button>
  <button role="tab" aria-selected="false" data-t="dash">Dashboard</button>
</nav>

<section id="summary">
{{if .Summary.Rows}}
<div class="tbar"><span id="rowcount"></span><button id="clearf" type="button">Clear filters</button></div>
<div class="scroll"><table id="jobs">
<thead>
<tr class="hrow">{{range $i,$c := .SumCols}}<th data-i="{{$i}}" data-kind="{{$c.Kind}}"><span class="hlab" title="Sort by {{$c.Name}}">{{$c.Name}}</span><span class="sar"></span><button class="fbtn" type="button" title="Filter {{$c.Name}}" aria-label="Filter {{$c.Name}}"><svg viewBox="0 0 12 12" width="9" height="9" aria-hidden="true"><path d="M1 2h10L7 6.5V11L5 9.6V6.5z" fill="currentColor"/></svg></button></th>{{end}}</tr>
</thead>
<tbody>{{range $ri,$row := .Cells}}{{$job := index $.ResumeJob $ri}}<tr>
{{range $i,$c := $row}}{{$n := (index $.SumCols $i).Name}}{{if eq $n "Stage"}}<td><span class="stage g-{{slug $c}}">{{$c}}</span></td>
{{else if eq $n "Verified"}}<td><span class="v-{{slug $c}}">{{$c}}</span></td>
{{else if eq $n "Status"}}<td>{{if $job}}<a class="pill s-{{slug $c}} doc" href="/doc?job={{$job}}" title="Open the tailored resume">{{$c}}</a>{{else}}<span class="pill s-{{slug $c}}">{{$c}}</span>{{end}}</td>
{{else if eq $n "Resume"}}<td>{{if $job}}<a class="doc" href="/doc?job={{$job}}" title="Open the tailored resume">{{$c}}</a>{{else}}{{$c}}{{end}}</td>
{{else if isURL $c}}<td><a href="{{$c}}" target="_blank" rel="noopener">{{short $c}}</a></td>
{{else}}<td>{{$c}}</td>{{end}}{{end}}
</tr>{{end}}</tbody></table></div>
{{else}}<div class="empty">No jobs on the list yet — run <code>/job</code>.</div>{{end}}
</section>

<section id="backlog" hidden>
{{if .Backlog.Rows}}<div class="scroll"><table>
<thead><tr>{{range .Backlog.Cols}}<th>{{.}}</th>{{end}}</tr></thead>
<tbody>{{range .Backlog.Rows}}<tr>
{{range $i,$c := .}}{{if eq (index $.Backlog.Cols $i) "Provisional"}}<td class="num fit">{{$c}}</td>
{{else if eq (index $.Backlog.Cols $i) "Status"}}<td><span class="pill s-{{slug $c}}">{{$c}}</span></td>
{{else if isURL $c}}<td><a href="{{$c}}" target="_blank" rel="noopener">{{short $c}}</a></td>
{{else}}<td>{{$c}}</td>{{end}}{{end}}
</tr>{{end}}</tbody></table></div>
<p class="note">Ranked but not yet verified. The next <code>/job-search</code> verifies from here
before sweeping sites again.</p>
{{else}}<div class="empty">Backlog is empty.</div>{{end}}
</section>

<section id="ix" hidden>
{{if .IX.Rows}}
  <div class="grid">
    <div class="card"><h3>Reports</h3><div class="big">{{.IXStats.Total}}</div>
      <div class="note">{{.IXStats.TrackA}} at a company · {{.IXStats.TrackB}} by an AI</div></div>
    <div class="card"><h3>Credibility</h3>
      <div class="big">{{.IXStats.MedianCred}} <small>median</small></div>
      <div class="note">low {{.IXStats.LowCred}} · high {{.IXStats.TopCred}}</div></div>
    <div class="card"><h3>Linked to a job</h3>
      <div class="big">{{.IXStats.Linked}}<small>/{{.IXStats.Total}}</small></div>
      <div class="note">matched to a row in jobs.md</div></div>
    <div class="card"><h3>Filtered out</h3><div class="big">{{.IXStats.Excluded}}</div>
      <div class="note">self-promo · news · venting</div></div>
    {{if .IXStats.CredBand}}<div class="card"><h3>Bands</h3>{{range .IXStats.CredBand}}
      <div class="bar"><span class="lbl">{{.Label}}</span>
        <span class="track"><span class="fill" style="width:{{pct .Pct}}%"></span></span>
        <span class="n">{{.N}}</span></div>{{end}}</div>{{end}}
  </div>
  <div class="scroll"><table>
  <thead><tr>{{range .IX.Cols}}<th>{{.}}</th>{{end}}</tr></thead>
  <tbody>{{range .IX.Rows}}<tr>
  {{range $i,$c := .}}{{if eq $i 0}}<td class="num">{{$c}}</td>
  {{else if eq (index $.IX.Cols $i) "Cred"}}<td class="num fit {{credClass $c}}">{{$c}}</td>
  {{else if eq (index $.IX.Cols $i) "Track"}}<td><span class="pill t-{{slug $c}}">{{$c}}</span></td>
  {{else if eq (index $.IX.Cols $i) "Outcome"}}{{if eq $c "—"}}<td class="s-not-stated">—</td>{{else}}<td><span class="pill s-{{slug $c}}">{{$c}}</span></td>{{end}}
  {{else if isURL $c}}<td><a href="{{$c}}" target="_blank" rel="noopener">{{short $c}}</a></td>
  {{else}}<td>{{$c}}</td>{{end}}{{end}}
  </tr>{{end}}</tbody></table></div>
  <p class="note">Credibility is a judgement, not a measurement — <code>single-source</code>
  claims stay flagged in the Details section of the file. Full quotes, per-round questions
  and caveats live in <code>{{.IXFile}}</code>.</p>
{{else}}<div class="empty">No interview experiences compiled yet — run
  <code>/interview-search</code>.</div>{{end}}
</section>

<section id="dash" hidden>
{{if .Stats.Blocked}}<div class="card needs">
  <h3>Needs you — {{len .Stats.Blocked}} blocked</h3>
  <p class="note">The pipeline skipped these and carried on. Paste each posting into
    <code>jobs/output/jds/&lt;job-id&gt;.md</code> (or drop the file in as
    <code>&lt;job-id&gt;-source.pdf</code>), then re-run <code>/job</code> — it picks them up
    and tailors the resume.</p>
  <table class="mini"><tbody>{{range .Stats.Blocked}}<tr>
    <td><code>{{.ID}}</code></td><td>{{.Role}}</td><td>{{.Company}}</td>
    <td class="why">{{.Why}}</td>
    <td>{{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener">posting ↗</a>{{end}}</td>
  </tr>{{end}}</tbody></table>
</div>{{end}}
  <div class="grid">
    <div class="card"><h3>On the list</h3><div class="big">{{.Stats.Total}}</div>
      <div class="note">{{.Stats.Actionable}} awaiting action</div></div>
    <div class="card"><h3>Pipeline</h3>
      <div class="big">{{.Stats.ResumeDone}}<small>/{{.Stats.Total}} tailored</small></div>
      <div class="note">JD {{.Stats.JDDone}} · resume {{.Stats.ResumeDone}} · form {{.Stats.FormDone}}{{if .Stats.NeedsJD}} · {{.Stats.NeedsJD}} need a JD{{end}}</div></div>
    <div class="card"><h3>Backlog</h3><div class="big">{{.Stats.Backlog}}</div>
      <div class="note">ranked, not yet verified</div></div>
    <div class="card"><h3>Excluded</h3><div class="big">{{.Stats.Excluded}}</div>
      <div class="note">dead · stale · agency reposts</div></div>
    <div class="card"><h3>Fit range</h3>
      <div class="big">{{.Stats.MedianFit}} <small>median</small></div>
      <div class="note">low {{.Stats.LowFit}} · high {{.Stats.TopFit}}</div></div>
    <div class="card"><h3>Comp disclosed</h3>
      <div class="big">{{.Stats.CompListed}}<small>/{{.Stats.Total}}</small></div>
      <div class="note">{{.Stats.CompMissing}} list no band</div></div>
  </div>
  <div class="grid">
    {{if .Stats.Profile}}<div class="card"><h3>Profile</h3>{{range .Stats.Profile}}
      <div class="bar"><span class="lbl">{{.Label}}</span>
        <span class="track"><span class="fill" style="width:{{pct .Pct}}%"></span></span>
        <span class="n">{{.N}} · {{pct .Pct}}%</span></div>{{end}}</div>{{end}}
    {{if .Stats.Status}}<div class="card"><h3>Status</h3>{{range .Stats.Status}}
      <div class="bar"><span class="lbl">{{.Label}}</span>
        <span class="track"><span class="fill" style="width:{{pct .Pct}}%"></span></span>
        <span class="n">{{.N}} · {{pct .Pct}}%</span></div>{{end}}</div>{{end}}
    {{if .Stats.Verified}}<div class="card"><h3>Verification</h3>{{range .Stats.Verified}}
      <div class="bar"><span class="lbl">{{.Label}}</span>
        <span class="track"><span class="fill" style="width:{{pct .Pct}}%"></span></span>
        <span class="n">{{.N}} · {{pct .Pct}}%</span></div>{{end}}</div>{{end}}
    {{if .Stats.FitBand}}<div class="card"><h3>Fit bands</h3>{{range .Stats.FitBand}}
      <div class="bar"><span class="lbl">{{.Label}}</span>
        <span class="track"><span class="fill" style="width:{{pct .Pct}}%"></span></span>
        <span class="n">{{.N}}</span></div>{{end}}</div>{{end}}
    {{if .Stats.Location}}<div class="card"><h3>Work mode</h3>{{range .Stats.Location}}
      <div class="bar"><span class="lbl">{{.Label}}</span>
        <span class="track"><span class="fill" style="width:{{pct .Pct}}%"></span></span>
        <span class="n">{{.N}}</span></div>{{end}}</div>{{end}}
    {{if .Stats.Source}}<div class="card"><h3>Source / ATS</h3>{{range .Stats.Source}}
      <div class="bar"><span class="lbl">{{.Label}}</span>
        <span class="track"><span class="fill" style="width:{{pct .Pct}}%"></span></span>
        <span class="n">{{.N}}</span></div>{{end}}</div>{{end}}
  </div>
</section>

<script>
const tabs=[...document.querySelectorAll('nav button')];
const show=id=>{
  tabs.forEach(b=>b.setAttribute('aria-selected',b.dataset.t===id));
  ['summary','backlog','ix','dash'].forEach(s=>document.getElementById(s).hidden=(s!==id));
  try{location.hash=id}catch(e){}
};
tabs.forEach(b=>b.onclick=()=>show(b.dataset.t));
const h=(location.hash||'').replace('#','');
if(['summary','backlog','ix','dash'].includes(h))show(h);
</script>
</div>
<div id="fpop" hidden role="dialog" aria-label="Column filter">
  <div class="fpop-h"><span id="fpop-t"></span><button id="fpop-x" type="button">close</button></div>
  <input id="fpop-q" type="search" placeholder="contains…" autocomplete="off">
  <div class="fpop-a">
    <button id="fpop-all" type="button">select all</button>
    <button id="fpop-none" type="button">clear</button>
  </div>
  <div id="fpop-l" class="fpop-l"></div>
  <div class="fpop-f">
    <button id="fpop-apply" type="button" class="prim">Apply</button>
    <button id="fpop-reset" type="button">Reset</button>
  </div>
</div>
<div id="drawer" hidden>
  <div id="dr-back"></div>
  <aside id="dr-panel" role="dialog" aria-modal="true" aria-label="Document">
    <div id="dr-bar"><span class="sp"></span>
      <a id="dr-open" href="#" target="_blank" rel="noopener">open in a tab ↗</a>
      <button id="dr-x" type="button" title="Close (Esc)" aria-label="Close">✕</button></div>
    <div id="dr-body"></div>
  </aside>
</div>

<script id="colmeta" type="application/json">{{.ColMeta}}</script>
<script>
(function(){
  var tbl = document.getElementById('jobs'); if(!tbl) return;
  var tb = tbl.tBodies[0], rows = Array.prototype.slice.call(tb.rows);
  var heads = Array.prototype.slice.call(tbl.querySelectorAll('.hrow th'));
  var meta = []; try{ meta = JSON.parse(document.getElementById('colmeta').textContent); }catch(e){}

  var pop  = document.getElementById('fpop'),
      popT = document.getElementById('fpop-t'), popQ = document.getElementById('fpop-q'),
      popL = document.getElementById('fpop-l');

  // filters[i] = array of accepted values. Absent or empty = column not filtered.
  // KEY is bumped whenever the saved shape changes, so stale state is dropped rather than
  // misread. jobtable4 stores filters by COLUMN NAME: the old format keyed them by column
  // index, so removing a column silently moved every saved filter one column to the left.
  var filters = {}, draft = null, sortI = -1, sortDir = 1, openI = -1, KEY = 'jobtable4';

  function cell(r,i){ var c = r.cells[i]; return c ? c.textContent.trim() : ''; }
  function num(v){ var n = parseFloat(String(v).replace(/[^0-9.\-]/g,'')); return isNaN(n) ? null : n; }
  function sel(i){ return filters[i] && filters[i].length ? filters[i] : null; }

  function apply(){
    var n = 0;
    rows.forEach(function(r){
      var ok = Object.keys(filters).every(function(k){
        var s = sel(k); return !s || s.indexOf(cell(r, +k)) !== -1;
      });
      r.classList.toggle('hide', !ok); if(ok) n++;
    });
    heads.forEach(function(h,i){ h.classList.toggle('filtered', !!sel(i)); });
    var c = document.getElementById('rowcount');
    if(c) c.textContent = n === rows.length ? rows.length + ' jobs'
                                            : n + ' of ' + rows.length + ' jobs';
    save();
  }

  function sortBy(i){
    var kind = heads[i].dataset.kind;
    if(sortI === i){ sortDir = -sortDir; } else { sortI = i; sortDir = 1; }
    rows.slice().sort(function(a,b){
      var x = cell(a,i), y = cell(b,i);
      if(kind === 'num'){
        var nx = num(x), ny = num(y);
        if(nx === null && ny === null) return 0;
        if(nx === null) return 1;                  // blanks last in both directions
        if(ny === null) return -1;
        return (nx - ny) * sortDir;
      }
      if(x === '' && y === '') return 0;
      if(x === '') return 1;
      if(y === '') return -1;
      return x.localeCompare(y, undefined, {numeric:true, sensitivity:'base'}) * sortDir;
    }).forEach(function(r){ tb.appendChild(r); });
    heads.forEach(function(h,hi){
      h.querySelector('.sar').textContent = hi === i ? (sortDir === 1 ? '↑' : '↓') : '';
      h.classList.toggle('sorted', hi === i);
    });
    save();
  }

  // ---- popover -------------------------------------------------------------
  function opts(i){ return (meta[i] && meta[i].options) || []; }

  function renderList(){
    var q = popQ.value.trim().toLowerCase();
    popL.innerHTML = '';
    var shown = 0;
    opts(openI).forEach(function(o){
      if(q && o.v.toLowerCase().indexOf(q) === -1) return;
      shown++;
      var row = document.createElement('label'); row.className = 'fopt';
      var cb = document.createElement('input'); cb.type = 'checkbox'; cb.value = o.v;
      cb.checked = draft.indexOf(o.v) !== -1;
      cb.addEventListener('change', function(){
        var k = draft.indexOf(o.v);
        if(cb.checked){ if(k === -1) draft.push(o.v); } else if(k !== -1){ draft.splice(k,1); }
      });
      var t = document.createElement('span'); t.className = 'fopt-t'; t.textContent = o.v;
      var n = document.createElement('span'); n.className = 'fopt-n'; n.textContent = o.n;
      row.appendChild(cb); row.appendChild(t); row.appendChild(n);
      popL.appendChild(row);
    });
    if(!shown){
      var e = document.createElement('div'); e.className = 'fopt-e';
      e.textContent = opts(openI).length ? 'No value matches “' + popQ.value + '”'
                                         : 'Nothing to filter in this column';
      popL.appendChild(e);
    }
  }

  function visibleBoxes(){
    return Array.prototype.slice.call(popL.querySelectorAll('input[type=checkbox]'));
  }

  function closePop(){
    pop.hidden = true; openI = -1; draft = null;
    heads.forEach(function(h){ h.classList.remove('open'); });
  }

  function openPop(i, btn){
    if(openI === i){ closePop(); return; }
    closePop(); openI = i; heads[i].classList.add('open');
    draft = (filters[i] || []).slice();
    popT.textContent = (meta[i] && meta[i].name) || heads[i].querySelector('.hlab').textContent;
    popQ.value = '';
    renderList();

    // position:fixed so the table's overflow-x:auto cannot clip it
    pop.hidden = false;
    var r = btn.getBoundingClientRect();
    var left = Math.min(r.left, window.innerWidth - pop.offsetWidth - 10);
    var top  = r.bottom + 6;
    var over = top + pop.offsetHeight - window.innerHeight + 10;
    pop.style.left = Math.max(8, left) + 'px';
    pop.style.top  = Math.max(8, over > 0 ? top - over : top) + 'px';
    popQ.focus();
  }

  popQ.addEventListener('input', renderList);
  document.getElementById('fpop-all').addEventListener('click', function(){
    visibleBoxes().forEach(function(cb){
      if(!cb.checked){ cb.checked = true; if(draft.indexOf(cb.value) === -1) draft.push(cb.value); }
    });
  });
  document.getElementById('fpop-none').addEventListener('click', function(){
    visibleBoxes().forEach(function(cb){
      if(cb.checked){ cb.checked = false; var k = draft.indexOf(cb.value); if(k !== -1) draft.splice(k,1); }
    });
  });
  document.getElementById('fpop-apply').addEventListener('click', function(){
    filters[openI] = draft.slice(); apply(); closePop();
  });
  document.getElementById('fpop-reset').addEventListener('click', function(){
    draft = []; filters[openI] = []; popQ.value = ''; renderList(); apply();
  });
  document.getElementById('fpop-x').addEventListener('click', closePop);

  heads.forEach(function(h,i){
    h.querySelector('.hlab').addEventListener('click', function(){ sortBy(i); });
    h.querySelector('.fbtn').addEventListener('click', function(e){ e.stopPropagation(); openPop(i, this); });
  });
  pop.addEventListener('click', function(e){ e.stopPropagation(); });
  document.addEventListener('click', closePop);
  document.addEventListener('keydown', function(e){
    if(e.key === 'Escape') closePop();
    if(e.key === 'Enter' && openI !== -1 && !pop.hidden){
      filters[openI] = draft.slice(); apply(); closePop();
    }
  });
  window.addEventListener('resize', closePop);
  document.querySelector('.scroll').addEventListener('scroll', closePop);

  var clear = document.getElementById('clearf');
  if(clear) clear.addEventListener('click', function(){ filters = {}; closePop(); apply(); });

  function colName(i){
    return (meta[i] && meta[i].name) || heads[i].querySelector('.hlab').textContent.trim();
  }
  function colIndex(n){
    for(var i = 0; i < heads.length; i++){ if(colName(i) === n) return i; }
    return -1;
  }

  function save(){
    var f = {};
    Object.keys(filters).forEach(function(k){
      if(filters[k] && filters[k].length) f[colName(+k)] = filters[k];
    });
    try{
      localStorage.setItem(KEY, JSON.stringify({f:f, s: sortI >= 0 ? colName(sortI) : '', d:sortDir}));
      localStorage.removeItem('jobtable3');        // the index-keyed format, now unreadable
    }catch(e){}
  }

  (function restore(){
    try{ localStorage.removeItem('jobtable3'); }catch(e){}
    try{
      var st = JSON.parse(localStorage.getItem(KEY) || '{}');
      Object.keys(st.f || {}).forEach(function(n){
        var i = colIndex(n);
        if(i < 0) return;                          // that column is gone — drop its filter
        // Keep only values the column still has. A filter whose every value has vanished
        // would hide every row and look like a broken table, so it is dropped instead.
        var have = (meta[i] && meta[i].options || []).map(function(o){ return o.v; });
        var keep = st.f[n].filter(function(v){ return have.indexOf(v) !== -1; });
        if(keep.length) filters[i] = keep;
      });
      var si = st.s ? colIndex(st.s) : -1;
      if(si >= 0){
        sortI = -1; sortDir = 1;                   // sortBy toggles from a clean slate
        sortBy(si);                                // ascending
        if(st.d === -1) sortBy(si);
      }
    }catch(e){}
  })();
  apply();
})();
</script>

<script>
// The drawer. Every /doc link is a real href, so middle-click and no-JS still work; the
// click handler just prefers to open it in place, next to the row it came from.
(function(){
  var dr = document.getElementById('drawer'), body = document.getElementById('dr-body'),
      open = document.getElementById('dr-open');
  function show(href){
    open.href = href;
    body.innerHTML = '<div class="empty">Loading…</div>';
    dr.hidden = false;
    document.body.style.overflow = 'hidden';
    fetch(href + '&frag=1').then(function(r){
      if(!r.ok) throw new Error(r.status === 404 ? 'That document is not on disk.' : 'HTTP ' + r.status);
      return r.text();
    }).then(function(h){
      body.innerHTML = h;
      body.scrollTop = 0;
    }).catch(function(e){
      body.innerHTML = '<div class="empty">Could not open it — ' + e.message + '</div>';
    });
  }
  function close(){ dr.hidden = true; body.innerHTML = ''; document.body.style.overflow = ''; }

  document.addEventListener('click', function(e){
    var a = e.target.closest ? e.target.closest('a.doc, .doc-tab') : null;
    if(!a || e.metaKey || e.ctrlKey || e.shiftKey || a.target === '_blank') return;
    e.preventDefault();
    show(a.getAttribute('href'));
  });
  // "copy all" on a code document (refill.js) — the file exists to be pasted somewhere else.
  document.addEventListener('click', function(e){
    var btn = e.target.closest ? e.target.closest('.doc-copy') : null;
    if(!btn) return;
    var code = btn.closest('#dr-body, body').querySelector('pre.code code');
    if(!code) return;
    var done = function(ok){ btn.textContent = ok ? 'copied ✓' : 'selected — press ⌘C';
      setTimeout(function(){ btn.textContent = 'copy all'; }, 2500); };
    // Fallback path: select the block so the keyboard shortcut works immediately, and try
    // the legacy copy command. The clipboard API is refused in plenty of contexts.
    var fallback = function(){
      try{
        var r = document.createRange(); r.selectNodeContents(code);
        var sel = window.getSelection(); sel.removeAllRanges(); sel.addRange(r);
        done(document.execCommand && document.execCommand('copy'));
      }catch(err){ done(false); }
    };
    if(navigator.clipboard && navigator.clipboard.writeText){
      navigator.clipboard.writeText(code.textContent).then(function(){ done(true); }, fallback);
    } else { fallback(); }
  });

  document.getElementById('dr-back').addEventListener('click', close);
  document.getElementById('dr-x').addEventListener('click', close);
  document.addEventListener('keydown', function(e){ if(e.key === 'Escape' && !dr.hidden) close(); });
})();
</script>
</body></html>`
