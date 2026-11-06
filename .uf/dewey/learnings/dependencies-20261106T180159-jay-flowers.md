---
tag: dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:01:59Z
identity: dependencies-20261106T180159-jay-flowers
tier: draft
---

In the unbound-force plugin suite, an undeclared npm import that resolves only through transitive hoisting is latent breakage by construction: `uf init` targets and doctor remediation install with `npm ci --ignore-scripts --omit=dev` (internal/scaffold/plugin_activation.go, internal/doctor/checks.go), so dev-chain hoisting (vitest→vite) is absent there, and runtime-chain hoisting is unreviewed and unpinned — any re-resolution can drop the package or hoist a different major. The manifest (dependencies + overrides) plus package-lock.json is the only pinning and review surface, and TestPluginPackageInputs_ExactVersionsAndLockIntegrity is the deliberate mechanism for changing the pinned dependency set — it is a gate to update through a reviewed change, never a constraint to route around by importing undeclared packages. Precedent: PR #958's instruction linter imported `yaml` undeclared with an in-code comment explaining the hoisting; maintainer feedback (PRRT_kwDORVoNbc60LQpa) required declaring it. Rule of thumb: every package a `.opencode/` source or test file imports must appear in `.opencode/package.json` (and its embedded scaffold mirror, internal/scaffold/assets/opencode/package.json), with the exact-versions test maps updated in the same change — note that a `dependencies` entry propagates to every hero repo's --omit=dev install, so dev-only tooling belongs in devDependencies.
