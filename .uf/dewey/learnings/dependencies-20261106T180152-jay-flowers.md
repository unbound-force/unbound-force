---
tag: dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:01:52Z
identity: dependencies-20261106T180152-jay-flowers
tier: draft
---

In the unbound-force repo, npm packages imported by `.opencode/lib` or `.opencode/test` code but absent from `.opencode/package.json` resolve only through transitive hoisting, and the plugin suite treats the dependency set as a governed surface: canonical `.opencode/package.json` + `package-lock.json` are exact-pinned, byte-mirrored into `internal/scaffold/assets/opencode/` (asserted by `TestPluginPackageInputs_MirrorsMatchCanonical`), and enumerated by `TestPluginPackageInputs_ExactVersionsAndLockIntegrity`, whose `assertLockedDirectPackage` helper also requires `license == "MIT"` — `yaml` is ISC, so declaring it as a direct dependency requires parameterizing that license check. Production-shaped installs are `npm ci --ignore-scripts --omit=dev` (`internal/scaffold/plugin_activation.go`), which drops devDependency-tree hoisting, and vite's `yaml` is only an optional peer (never auto-installed), so `@opencode-ai/plugin -> effect -> yaml` is in practice the sole hoisting path for a phantom `yaml` import. When adding an npm import to the suite: declare it with an exact version in `.opencode/package.json` (runtime `dependencies` for anything executed outside vitest, e.g. gate tools run via `node`), regenerate the lockfile, sync both scaffold mirrors, and update the ExactVersionsAndLockIntegrity expected maps.
