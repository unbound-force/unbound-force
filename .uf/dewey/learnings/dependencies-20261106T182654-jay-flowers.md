---
tag: dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:26:54Z
identity: dependencies-20261106T182654-jay-flowers
tier: draft
---

An undeclared import resolved via transitive hoisting (e.g. `yaml` arriving via @opencode-ai/plugin → effect and vitest → vite when the manifest pins exactly two runtime dependencies) makes a tool's availability hostage to dependency trees it does not own: any transitive update can remove the module at runtime with no install-time signal. Tools wired into pre-flight or CI must declare every runtime import in the manifest — and in the scaffold mirror where drift tests pin the dependency set. An in-code comment acknowledging the transitive reliance documents the defect; it does not fix it.
