---
tag: coverage-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-16T02:24:12Z
identity: coverage-gate-20261116T022412-jay-flowers
tier: draft
as_of: "unbound-force@1677fd12d4ed8969d5ac87110e35ca030b9326b1"
evidence:
  - "dewey:cmd/coverage-gate/main.go"
  - "dewey:internal/coveragegate/gate.go"
  - "dewey:coverage-gate.json"
  - "session:ses_current-uf-pipeline-1048"
---

cmd/coverage-gate vacuous-pass gotcha on uncommitted-local work (recorded 2026-11-16 on branch opsx/heredoc-ban-escaped-and-ansic): changedPackages() in cmd/coverage-gate/main.go computes changed packages from committed trees only — `git merge-base <base> HEAD` plus `git diff --name-only --diff-filter=ACMR <mergebase> HEAD` — so on a feature branch whose HEAD equals the base (all changes uncommitted), `go run ./cmd/coverage-gate -manifest coverage-gate.json -profile coverage.out -base main` reports "0 changed package(s) within scope thresholds" and exits 0 vacuously; the uncommitted change is invisible. For uncommitted work, verify the scope gate manually: list changed packages via `git diff main --stat` (working tree), map them through coverage-gate.json scope thresholds (helper 80, config 90, mapping 90), and compare against the profile's measured coverage. Note: `go test -coverprofile` still writes the profile when an unrelated package fails, so the profile remains usable for the manual check (observed: internal/scaffold 90.0% vs helper threshold 80, global ratchet 89.6% >= 80.0%).
