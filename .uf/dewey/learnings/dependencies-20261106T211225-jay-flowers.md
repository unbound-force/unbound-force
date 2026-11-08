---
tag: dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-06T21:12:25Z
identity: dependencies-20261106T211225-jay-flowers
tier: draft
---

When a TypeScript module must import a package that is not a declared dependency (e.g. `yaml` in .opencode/lib/instruction-lint.ts, resolved through transitive hoisting from @opencode-ai/plugin→effect and vitest→vite because the scaffold pins the plugin package to exactly two runtime dependencies), pair the import with a resolution-guard test: createRequire from the importing module, require.resolve the package, and on failure throw an error naming the package and the cause ("transitive hoisting changed"). This converts a silent ERR_MODULE_NOT_FOUND at CLI startup — in a tool wired into preflight and make targets — into a named CI failure at the exact moment the pinned graph changes, and documents the coupling for the next maintainer. The guard is a mitigation, not a fix: the structural exit is declaring the dependency through whatever governance gate pins the dependency count, or vendoring a minimal parser.
