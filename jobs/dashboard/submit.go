package main

import (
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// submittable are the statuses the Summary offers a "Mark submitted" button on — the same
// set jobs/bin/mark-submitted.sh will move to `submitted`.
var submittable = map[string]bool{
	"found": true, "shortlisted": true, "jd-captured": true,
	"tailored": true, "filled-awaiting-user": true,
}

// submitHandler is the dashboard's "Mark submitted" button: the user's own confirmation
// that they clicked Submit on the employer's form. It is the screen version of
// `make submitted JOB=<id>` and runs the same script — the only thing allowed to set
// `submitted` — after a snapshot, so the change is one file away from undone.
//
// The server listens on localhost, so any web page the user visits could try to POST
// here. A custom header can't be sent cross-origin without a CORS preflight this server
// never answers, and the Origin check rejects the rest.
func submitHandler(liveDir string) http.HandlerFunc {
	root := filepath.Clean(filepath.Join(liveDir, "..", ".."))
	script := filepath.Join(root, "jobs", "bin", "mark-submitted.sh")
	snap := filepath.Join(root, "jobs", "bin", "snapshot.sh")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("X-Dashboard") != "1" || !sameOrigin(r) {
			http.Error(w, "cross-origin request refused", http.StatusForbidden)
			return
		}
		id := strings.TrimSpace(r.FormValue("job"))
		if !jobIDRe.MatchString(id) {
			http.Error(w, "bad job id", http.StatusBadRequest)
			return
		}
		if _, err := os.Stat(script); err != nil {
			http.Error(w, "jobs/bin/mark-submitted.sh not found next to this jobs.md — use make submitted", http.StatusNotImplemented)
			return
		}
		if out, err := runIn(root, "bash", snap); err != nil {
			http.Error(w, "snapshot failed, nothing changed: "+out, http.StatusInternalServerError)
			return
		}
		out, err := runIn(root, "bash", script, id)
		if err != nil || !strings.Contains(out, "✓") && !strings.Contains(out, "already") {
			http.Error(w, strings.TrimSpace(out), http.StatusUnprocessableEntity)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(strings.TrimSpace(out)))
	}
}

func runIn(dir string, name string, args ...string) (string, error) {
	c := exec.Command(name, args...)
	c.Dir = dir
	b, err := c.CombinedOutput()
	return string(b), err
}

// sameOrigin accepts a request whose Origin (or, failing that, Referer) is this server.
func sameOrigin(r *http.Request) bool {
	o := r.Header.Get("Origin")
	if o == "" {
		o = r.Header.Get("Referer")
	}
	if o == "" {
		return false
	}
	u, err := url.Parse(o)
	return err == nil && u.Host == r.Host
}
