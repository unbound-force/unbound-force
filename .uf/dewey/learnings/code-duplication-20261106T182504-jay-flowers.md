---
tag: code-duplication
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:25:04Z
identity: code-duplication-20261106T182504-jay-flowers
tier: draft
---

Duplication born inside a single change drifts before the change merges — review it as active divergence, not future risk. In unbound-force PR 959, findRoot was copy-pasted into two new .opencode/lib CLIs (instruction-lint.ts, tool-inventory.ts) within one PR, and the copies already disagreed at review time (existsSync vs isDir for the .opencode root check), letting the two chained CLIs select different workspace roots when .opencode exists as a file; no test on either copy pinned the semantics, so the drift survived authoring. Reviewer rule of thumb: when one change introduces two or more executables sharing plumbing (root discovery, IO injection, subprocess wrappers, error formatting), treat extraction to a shared lib module as part of the same increment — born-here duplication has zero external consumers, co-located new tests, and minimum lifetime extraction cost, while deferral inherits a demonstrated drift rate. When the change is itself governance tooling for a one-source-of-truth principle (DE-001, IL-050), shipping hand-copied plumbing costs one severity notch above the naive read.
