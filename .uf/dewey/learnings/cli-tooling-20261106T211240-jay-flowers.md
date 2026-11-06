---
tag: cli-tooling
author: jay-flowers
category: pattern
created_at: 2026-11-06T21:12:40Z
identity: cli-tooling-20261106T211240-jay-flowers
tier: draft
---

Node CLI exit-code contracts need a catch-all at the entry seam, not per-callsite try/catch. An uncaught throw in a node process exits with code 1 — the same code most lint-style tools document for "findings found" — so any uncovered IO path (a git spawn, a directory walk) silently violates the documented contract and makes consumers misclassify a crash as a findings result. The fix that scales across sibling CLIs is one guard in the shared entry wrapper: cli-common.ts runMain wraps its main() call in try/catch, prints the error to the injected err writer, and sets process.exitCode to the documented usage/IO-error code (2), so every CLI sharing the helper inherits the guarantee and per-callsite wrapping becomes message polish rather than contract enforcement.
