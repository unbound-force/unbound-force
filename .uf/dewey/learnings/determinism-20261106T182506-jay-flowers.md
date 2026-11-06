---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:25:06Z
identity: determinism-20261106T182506-jay-flowers
tier: draft
---

The DE twice-test applies across tools, not just across runs: when one increment adds two CLIs to the same lib directory, every shared mechanical procedure (repo-root discovery, IO-injection interface, git wrappers, error formatting) needs exactly one implementation, or the copies drift independently and every fix lands twice. PR 959 (unbound-force) added instruction-lint.ts and tool-inventory.ts with parallel findRoot helpers that diverged at birth — existsSync accepts a file named .opencode, isDir requires a directory — so the two CLIs could resolve different repo roots in one workspace, one of them silently reporting an empty inventory at exit 0. Extract shared CLI plumbing into a shared module in the same increment that introduces the second consumer; when unifying drifted copies, unify on the stricter predicate and pin it with a unit test.
