---
tag: npm-phantom-dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:02:01Z
identity: npm-phantom-dependencies-20261106T180201-jay-flowers
tier: draft
---

A top-level `import { parseDocument } from "yaml"` in `.opencode/lib/instruction-lint.ts` resolves only because npm hoists effect's transitive yaml (via @opencode-ai/plugin) into `.opencode/node_modules`. Under `npm ci --ignore-scripts --omit=dev` — the uf init re-run install path (internal/scaffold/plugin_activation.go) — the dev-side carriers (vitest→vite peer) are pruned, so resolution hangs on a single runtime transitive and on hoisting, which is a resolution artifact rather than a contract. Any lockfile regeneration can silently drop or re-major the hoisted copy, breaking make lint-instructions, the advisory pre-flight check, and make plugin-test at import time with no manifest change. Direct imports must be declared in .opencode/package.json; the four coordinated pin surfaces (expectedRuntime in TestPluginPackageInputs_ExactVersionsAndLockIntegrity, embedded mirrors under internal/scaffold/assets/opencode/, provenance testdata + expectedPackages, regenerated lockfile) must be extended in the same change.
