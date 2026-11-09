---
tag: error-handling
author: jay-flowers
category: pattern
created_at: 2026-11-08T22:40:12Z
identity: error-handling-20261108T224012-jay-flowers
tier: draft
---

When a CLI establishes "escape control characters at every print site" as its untrusted-content invariant (terminal control-sequence injection), the error paths are print sites too. Branches that interpolate spawned-process stderr (`io.err(`cannot read base ${ref}: ${spawnDetail(e)}`)`) bypass a `print` helper that escapes, because the helper is wired only to stdout. Route every human-facing write — out and err, success and failure — through the escaping writer, or give the CLI a single escaping err function next to `print`. PR #959 round 3: both instruction-lint.ts and tool-inventory.ts missed the `cannot read base` branches (Adversary LOW finding).
