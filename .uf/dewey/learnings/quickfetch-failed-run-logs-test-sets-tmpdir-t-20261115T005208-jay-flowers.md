---
tag: quickfetch-failed-run-logs-test-sets-tmpdir-t
author: jay-flowers
category: pattern
created_at: 2026-11-15T00:52:08Z
identity: quickfetch-failed-run-logs-test-sets-tmpdir-t-20261115T005208-jay-flowers
tier: draft
---

Learning from branch quick/fetch-failed-run-logs-test-sets-tmpdir-t (issue #1051), 2026-11-14, category pattern, topic pre-flight causality classification. When a pre-flight hard gate fails on a check whose inputs the branch never touched, the failure is provably pre-existing and must stay out of the build/test fix loop. The method that worked here: first establish input identity — git diff <merge-base> --name-only restricted to the failing tool's input file classes (for golangci-lint: '*.go', go.mod, go.sum, .golangci.yml) returned empty, so the branch's lint inputs are byte-identical to the baseline; then confirm empirically with the pre-flight skill's Tier 2 method — a detached git worktree at the merge-base running only the failing tool reproduced the identical errcheck finding (internal/gateway/gateway_test.go:4383, exit 1). The gate then ran the remaining make check components (test, build, coverage-gate, plugin-test) individually to complete in-scope verification, recorded the failure as pre-existing in the quick-pre-flight checkpoint inventory, and treated the gate as green without scope-creeping a TypeScript-only fix into Go lint repair.
