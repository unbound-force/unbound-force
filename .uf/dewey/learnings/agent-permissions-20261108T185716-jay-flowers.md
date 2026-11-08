---
tag: agent-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:57:16Z
identity: agent-permissions-20261108T185716-jay-flowers
tier: draft
---

A security-regression test that reads one named file enforces its invariant only for that file; every new file carrying the same shape re-opens the hole the test closed. PR 959 shipped divisor-determinism.md with trailing-star bash allows and no terminal metacharacter denies while the bash-hardening test stayed green, because the test hardcoded the divisor-linguist.md path. Tests guarding class-wide security invariants (persona bash permission blocks, terminal deny placement) should enumerate the class — glob .opencode/agents/*.md and filter to files whose bash map contains allow entries — rather than name one member. Extending the glob also surfaced that divisor-curator.md (gh-prefixed allows) lacks the denies, confirming the per-file shape hides sibling regressions; fix all members of the class in the same change so the class-wide test lands green.
