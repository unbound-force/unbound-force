.PHONY: check build test lint install packages
.PHONY: coverage ensure-gaze crapload crapload-baseline crapload-check

check: lint test build

build:
	go build ./...

test:
	go test -race -count=1 ./...

lint:
	go vet ./...
	if command -v golangci-lint > /dev/null; then golangci-lint run; else echo "golangci-lint not installed, skipping"; fi

install:
	go build -o $(shell go env GOPATH)/bin/unbound-force ./cmd/unbound-force/
	ln -sf $(shell go env GOPATH)/bin/unbound-force $(shell go env GOPATH)/bin/uf

##@ OpenPackage Generation

packages: ## generate packages/ from .opencode/ sources
	./scripts/generate-packages.sh

##@ CRAP Load Monitoring

GAZE_VERSION ?= latest
GAZE_BASELINE := .gaze/baseline.json
GAZE_COVERPROFILE := coverage.out
GAZE_NEW_FUNC_THRESHOLD ?= 30

coverage: ## run tests with coverage profile
	go test -race -count=1 -coverprofile=$(GAZE_COVERPROFILE) ./...

ensure-gaze: ## install gaze if not present
	@command -v gaze >/dev/null 2>&1 || \
		(echo "Installing gaze..." && go install github.com/unbound-force/gaze/cmd/gaze@$(GAZE_VERSION))

crapload: ensure-gaze coverage ## run CRAP and GazeCRAP analysis (human-readable)
	gaze crap --format=text --coverprofile=$(GAZE_COVERPROFILE) ./...

crapload-baseline: ensure-gaze coverage ## generate baseline in .gaze/baseline.json
	@mkdir -p .gaze
	@REPO_ROOT=$$(pwd); \
	gaze crap --format=json --coverprofile=$(GAZE_COVERPROFILE) ./... | \
		jq --arg root "$$REPO_ROOT/" '(.scores[],.summary.worst_crap[]?,.summary.worst_gaze_crap[]?) |= (.file |= ltrimstr($$root))' > $(GAZE_BASELINE)
	@echo "Baseline written to $(GAZE_BASELINE)"

crapload-check: ensure-gaze coverage ## check for CRAP regressions against baseline
	@if [ ! -f $(GAZE_BASELINE) ]; then \
		echo "ERROR: Baseline file $(GAZE_BASELINE) not found. Run 'make crapload-baseline' first."; \
		exit 1; \
	fi
	@REPO_ROOT=$$(pwd); \
	gaze crap --format=json --coverprofile=$(GAZE_COVERPROFILE) ./... | \
		jq --arg root "$$REPO_ROOT/" '(.scores[],.summary.worst_crap[]?,.summary.worst_gaze_crap[]?) |= (.file |= ltrimstr($$root))' > /tmp/crapload-current.json
	@echo "Comparing against baseline..."
	@jq -r '.scores[] | "\(.file):\(.function) \(.crap) \(.gaze_crap // 0)"' $(GAZE_BASELINE) | sort > /tmp/crapload-baseline.txt
	@jq -r '.scores[] | "\(.file):\(.function) \(.crap) \(.gaze_crap // 0)"' /tmp/crapload-current.json | sort > /tmp/crapload-current.txt
	@REGRESSIONS=0; \
	while IFS=' ' read -r func crap gaze_crap; do \
		baseline_crap=$$(grep -F "$$func " /tmp/crapload-baseline.txt | head -1 | awk '{print $$2}'); \
		baseline_gaze=$$(grep -F "$$func " /tmp/crapload-baseline.txt | head -1 | awk '{print $$3}'); \
		if [ -z "$$baseline_crap" ]; then \
			if [ "$$(echo "$$crap > $(GAZE_NEW_FUNC_THRESHOLD)" | bc -l)" = "1" ]; then \
				echo "NEW FUNCTION VIOLATION: $$func CRAP=$$crap (threshold=$(GAZE_NEW_FUNC_THRESHOLD))"; \
				REGRESSIONS=$$((REGRESSIONS + 1)); \
			fi; \
		else \
			if [ "$$(echo "$$crap > $$baseline_crap" | bc -l)" = "1" ]; then \
				echo "REGRESSION: $$func CRAP $$baseline_crap -> $$crap"; \
				REGRESSIONS=$$((REGRESSIONS + 1)); \
			fi; \
			if [ "$$(echo "$$gaze_crap > $$baseline_gaze" | bc -l)" = "1" ]; then \
				echo "REGRESSION: $$func GazeCRAP $$baseline_gaze -> $$gaze_crap"; \
				REGRESSIONS=$$((REGRESSIONS + 1)); \
			fi; \
		fi; \
	done < /tmp/crapload-current.txt; \
	if [ $$REGRESSIONS -gt 0 ]; then \
		echo "FAIL: $$REGRESSIONS regression(s) detected"; \
		exit 1; \
	else \
		echo "PASS: No regressions detected"; \
	fi
