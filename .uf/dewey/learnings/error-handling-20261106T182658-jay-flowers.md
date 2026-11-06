---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:26:58Z
identity: error-handling-20261106T182658-jay-flowers
tier: draft
---

A Node CLI that documents "exit 1 = findings found" is silently violated by uncaught crashes, which also exit 1 — and static ESM imports fail at module resolution before main() runs, so a try/catch in main cannot catch a missing dependency. In unbound-force pre-flight checks, `requires: [node]` verifies the binary, not `.opencode/node_modules`; a transitively-installed dependency (e.g. `yaml` in instruction-lint.ts) makes the fresh-checkout-without-install failure mode concrete. Fix pattern: reserve exit 2 for "could not run" (matching the usage/config-error convention), load undeclared or transitive dependencies via dynamic `await import()` inside main with a remedy message naming the install command, optionally pre-check node_modules in the shell wrapper, and add a missing-node_modules regression test to the seam suite — the happy-path seam test that symlinks node_modules hides exactly this failure.
