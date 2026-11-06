---
tag: dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:01:55Z
identity: dependencies-20261106T180155-jay-flowers
tier: draft
---

The .opencode plugin suite's dependency governance is a five-part surface: .opencode/package.json (exact pins), .opencode/package-lock.json, the embedded mirror internal/scaffold/assets/opencode/package.json, the pinning/provenance tests in internal/scaffold/scaffold_test.go (TestPluginPackageInputs_ExactVersionsAndLockIntegrity, TestPluginPackageInputs_MirrorsMatchCanonical, TestPluginPackageInputs_ApprovedMaintenanceProvenance), and the AGENTS.md Technology Stack list. Repo-local tooling placed under .opencode/lib/ that needs a package beyond the two production deps must declare it in devDependencies and update all five surfaces — transitive hoisting (e.g. effect→yaml ^2.9.0, vite→yaml) satisfies the resolver today but is invisible to every governance check and breaks on the next dependency refresh. Placement matters: scaffolded targets install via `npm ci --ignore-scripts --omit=dev` (internal/scaffold/plugin_activation.go), so a `dependencies` declaration for repo-local tooling costs every generated target an unused package, while `devDependencies` is skipped there and matches the linter's actual execution contexts (make target, pre-flight, persona invocation, vitest — all full `npm ci` installs in the source repo).
