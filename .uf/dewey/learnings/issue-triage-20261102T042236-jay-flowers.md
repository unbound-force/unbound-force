---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:22:36Z
identity: issue-triage-20261102T042236-jay-flowers
tier: draft
---

In unbound-force/unbound-force, scaffolded content fixes must scope both copies of each file pair: the canonical `.opencode/<agents|commands>/<file>.md` and its byte-identical embedded mirror under `internal/scaffold/assets/opencode/` (drift-detection tests enforce parity per AGENTS.md Testing Conventions). Issue #633 named only the two canonical copies (cobalt-crush-dev.md:110, uf.review-council.md:604); grep verification showed the embedded mirrors carry the same stale `go install github.com/unbound-force/gaze/cmd/gaze@latest` guidance at the same lines. When triaging content-fix issues that name scaffolded `.opencode/` files, grep the affected text repo-wide before confirming scope — a fix without the embedded mirrors is incomplete and fails CI drift tests. Ownership check matters too: `.opencode/agents/gaze-reporter.md` has no UF mirror because it is gaze-scaffolded (one of gaze's nine embedded assets), so its fixes belong in the gaze repo.
