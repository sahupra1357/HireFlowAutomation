# AI Job Application workspace
#
# The /job skill (and its task files in jobs/tasks/) runs inside Claude Code.
# This Makefile only covers the Go dashboard that reads jobs/output/jobs.md.

PORT ?= 8080
JOBS ?= jobs/output/jobs.md
IX   ?= interviews/output/interviews.md
BIN  ?= bin/jobs-dashboard

.DEFAULT_GOAL := help
.PHONY: help web pdf refill daily morning submitted reset build install run-bin fmt vet check clean stats ix-stats history snapshot stop status

help: ## Show this help
	@echo "AI Job Application — make targets"
	@echo
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
	  | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
	@echo
	@echo "  dashboard : http://localhost:$(PORT)"
	@echo "  jobs      : $(JOBS)"
	@echo "  interviews: $(IX)"
	@echo "  override  : make web PORT=9000 JOBS=/other/jobs.md IX=/other/interviews.md
	@echo "  layout    : jobs/{input,output,dashboard}  interviews/{input,output,bin}""

web: build ## Run the dashboard, reloading jobs.md on every request (Ctrl+C to stop)
	@test -f "$(JOBS)" || { echo "error: $(JOBS) not found — run /job first"; exit 1; }
	@echo "  dashboard → http://localhost:$(PORT)     Ctrl+C to stop  ·  or: make stop"
	@exec "./$(BIN)" -addr ":$(PORT)" -f "$(JOBS)" -i "$(IX)"

pdf: build ## Render tailored resumes to PDF in the source-resume format (JOB=<job-id>, default all)
	@"./$(BIN)" -f "$(JOBS)" -pdf "$(or $(JOB),all)"

refill: ## Regenerate the replay scripts from form-fill.json (JOB=<job-id>, default all)
	@python3 jobs/bin/make-refill.py $(or $(JOB),--all)

daily: ## The whole loop: search → JDs → tailored resumes → filled forms on screen
	@bash jobs/bin/daily-run.sh

morning: ## Open every mapped application, filled, in a visible browser (JOB=<job-id> for one)
	@bash jobs/bin/morning-run.sh $(JOB)

submitted: ## Record that YOU submitted an application (make submitted JOB=<job-id>)
	@test -n "$(JOB)" || { echo "usage: make submitted JOB=<job-id>"; exit 1; }
	@bash jobs/bin/mark-submitted.sh $(JOB)

stop: ## Stop any running dashboard
	@n=$$(pgrep -f "jobs-dashboard -addr|dashboard -addr :" 2>/dev/null | wc -l | tr -d ' '); \
	if [ "$$n" = "0" ]; then echo "  no dashboard running"; else \
	  pkill -f "jobs-dashboard -addr" 2>/dev/null; \
	  pkill -f "dashboard -addr :" 2>/dev/null; \
	  sleep 1; echo "  stopped $$n dashboard process(es)"; fi

status: ## Show whether the dashboard is running, and where
	@out=$$(pgrep -fl "jobs-dashboard -addr|dashboard -addr :" 2>/dev/null); \
	if [ -n "$$out" ]; then echo "$$out" | sed 's/^/  /'; else echo "  not running"; fi
	@out=$$(lsof -nP -iTCP:$(PORT) -sTCP:LISTEN 2>/dev/null | tail -n +2); \
	if [ -n "$$out" ]; then echo "  port $(PORT): in use"; else echo "  port $(PORT): free"; fi

build: ## Compile the dashboard to bin/
	@mkdir -p $(dir $(BIN))
	go build -o "$(BIN)" ./jobs/dashboard
	@echo "built $(BIN)"

run-bin: build ## Build, then run the compiled binary
	"./$(BIN)" -addr ":$(PORT)" -f "$(JOBS)" -i "$(IX)"

install: ## Install the dashboard to your GOBIN
	go install ./jobs/dashboard

fmt: ## Format Go sources
	gofmt -w jobs/dashboard/

vet: ## Run go vet
	go vet ./jobs/dashboard/

check: fmt vet build ## fmt + vet + build
	@echo "all checks passed"

stats: ## Print job counts without starting the server
	@test -f "$(JOBS)" || { echo "error: $(JOBS) not found"; exit 1; }
	@awk '/^## Summary/{s=1;next} /^## Unverified backlog/{s=2;next} /^## Excluded/{s=3;next} \
	      /^## /{s=0} /^\|/{if($$0 !~ /^\|[- :|]+\|$$/ && s){c[s]++}} \
	      END{printf "  on the list: %d\n  backlog:     %d\n  excluded:    %d\n", c[1]-1, c[2]-1, c[3]-1}' "$(JOBS)"

snapshot: ## Back up jobs.md to jobs/output/history/
	@jobs/bin/snapshot.sh

history: ## List jobs.md snapshots, newest first
	@ls -1t jobs/output/history/jobs-*.md 2>/dev/null \
	  | awk '{n=$$0; sub(/.*jobs-/,"",n); sub(/\.md$$/,"",n); printf "  %-22s %s\n", n, $$0}' \
	  || echo "  no snapshots yet"

reset: ## Wipe everything the agent generated and start fresh (--dry-run first!)
	@bash jobs/bin/reset.sh $(ARGS)

clean: ## Remove build artifacts
	rm -rf bin/
	@echo "cleaned"

ix-stats: ## Print interview-experience counts without starting the server
	@test -f "$(IX)" || { echo "error: $(IX) not found — run /interview-search first"; exit 1; }
	@awk '/^## Reports/{s=1;next} /^## Unverified backlog/{s=2;next} /^## Excluded/{s=3;next} \
	      /^## /{s=0} /^\|/{if($$0 !~ /^\|[- :|]+\|$$/ && s){c[s]++}} \
	      END{printf "  reports:  %d\n  backlog:  %d\n  excluded: %d\n", c[1]-1, c[2]-1, c[3]-1}' "$(IX)"
