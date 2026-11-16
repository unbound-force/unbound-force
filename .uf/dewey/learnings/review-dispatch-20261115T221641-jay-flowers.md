---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-15T22:16:41Z
identity: review-dispatch-20261115T221641-jay-flowers
tier: draft
as_of: "unbound-force@6a553043168ccaffc4677672b76b30f5453aa1b5"
evidence:
  - "schemas/review-dispatch-failure/v1.0.0.schema.json"
  - ".uf/schemas/review-dispatch-failure/v1.0.0.schema.json"
  - "internal/scaffold/assets/uf/schemas/review-dispatch-failure/v1.0.0.schema.json"
  - "internal/scaffold/scaffold_test.go"
  - "internal/schemas/ci_test.go"
---

Schema-registry changes in unbound-force follow a three-copy mirror discipline (verified on RD-FR-066, branch opsx/review-dispatch-errors-cap-loses-unrecov, 2026-11-15): (1) author the canonical schema under schemas/<name>/vN.N.N.schema.json with README.md and samples/ (valid + invalid fixtures); (2) mirror the full directory to .uf/schemas/<name>/ and the schema alone to internal/scaffold/assets/uf/schemas/<name>/ — all three copies must be blob-identical (sha256 match); (3) register the canonical path in internal/schemas/ci_test.go handAuthoredSchemas so the fixture-validation suite exercises every sample; (4) add the scaffold asset path to internal/scaffold/scaffold_test.go task42CanonicalAssets and expectedAssetPaths so drift tests pin the mirrors; (5) bump the uf init file-count window in cmd/unbound-force/main_test.go. The plugin source mirror (.opencode/plugins/review-dispatch/index.ts to internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts) is byte-identical and pinned the same way. No sync helper exists — the drift tests verify parity but never copy; the copies are made by hand (or by task delegation) in the same change.
