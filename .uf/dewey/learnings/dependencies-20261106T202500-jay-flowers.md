---
tag: dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-06T20:25:01Z
identity: dependencies-20261106T202500-jay-flowers
tier: draft
---

In the unbound-force plugin suite, an npm package imported by `.opencode/lib` or `.opencode/test` code but absent from `.opencode/package.json` resolves only through transitive hoisting (e.g. `yaml` arriving via @opencode-ai/plugin → effect, and vitest → vite), which is a resolution artifact, not a contract: any lockfile regeneration can drop or re-major the hoisted copy and break the tool at import time with no manifest change. The dependency set is a five-part governed surface that must move together in one change: (1) `.opencode/package.json` (exact pins), (2) `.opencode/package-lock.json` regenerated, (3) the embedded mirrors under `internal/scaffold/assets/opencode/` (byte-identity asserted), (4) the pin/provenance tests in `internal/scaffold/scaffold_test.go` (`TestPluginPackageInputs_ExactVersionsAndLockIntegrity` expected maps — note its `assertLockedDirectPackage` helper defaults to `license == "MIT"`, so an ISC package like `yaml` needs that check parameterized — plus mirrors-match and provenance testdata), and (5) the AGENTS.md Technology Stack list. Placement matters: scaffolded targets and doctor remediation install with `npm ci --ignore-scripts --omit=dev` (internal/scaffold/plugin_activation.go), so a `dependencies` entry propagates an unused package to every generated hero repo, while `devDependencies` is skipped there and matches repo-local tooling's actual execution contexts (make target, pre-flight, persona invocation, vitest — all full installs in the source repo). An in-code comment acknowledging a transitive reliance documents the defect; it does not fix it. Precedent: PR #958's instruction linter imported `yaml` undeclared until review (PRRT_kwDORVoNbc60LQpa) required declaring it.
