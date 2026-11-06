---
tag: dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-06T21:12:20Z
identity: dependencies-20261106T211220-jay-flowers
tier: draft
---

A resolution guard for a phantom (transitively-hoisted) import that runs inside the dev install cannot detect loss of the runtime dependency chain when a dev-only chain also provides the package: the guard resolves against the combined dev graph and passes while a production-only (npm ci --omit=dev) install breaks. Observed on unbound-force PR #959: .opencode/lib/instruction-lint.ts imports yaml, provided by both @opencode-ai/plugin→effect (runtime chain) and vitest→vite (dev-only chain); the shipped "runtime dependencies" vitest test guards only the combined graph. Guard the runtime graph independently — resolve against an --omit=dev install in CI, or (preferred) declare the dependency so the manifest, not the hoist, is the contract.
