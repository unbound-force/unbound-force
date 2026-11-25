---
tag: scaffold-assets
author: jay-flowers
category: gotcha
created_at: 2026-11-25T01:04:01Z
identity: scaffold-assets-20261125T010401-jay-flowers
tier: draft
as_of: "unbound-force@e74c8f8651d35f09a15c12d8e5d05a451c787ed5"
---

Adding a scaffolded, activation-staged asset to internal/scaffold/assets requires TWO test updates in the same change: the scaffold-level inventory/drift tests (internal/scaffold/scaffold_test.go: expectedAssetPaths, task42CanonicalAssets, TestIsToolOwned, TestIsDivisorAsset) AND the CLI-level count range in cmd/unbound-force/main_test.go TestRunInit_FreshDir, which asserts an environment-dependent 'N files processed' range. Observed on PR #1206 (delegate-abort-recovery): opencode/lib/report-recovery.ts pushed the fully-activated local count from 90 to 91, all scaffold-level tests were updated, but the [86-90] range was not — CI stayed green because CI-runner activation probes yield a lower count, while `go test -race -count=1 ./...` failed locally in fully-activated environments. When updating the range, also refresh the explanatory comment block above the assertion, which documents the expected count and its activation-variance rationale.
