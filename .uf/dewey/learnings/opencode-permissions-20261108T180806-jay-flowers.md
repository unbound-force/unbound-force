---
tag: opencode-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:08:06Z
identity: opencode-permissions-20261108T180806-jay-flowers
tier: draft
---

A permission-hardening test that hardcodes one persona file stops at that file: the next persona shipped with bash allows passes every gate while missing the hardening. Observed on unbound-force PR #959: instruction-lint-preflight.test.ts's "keeps the persona's metacharacter denies after every bash allow" reads only divisor-linguist.md, and the new divisor-determinism.md shipped trailing-star node allows with no *;*, *|*, *>*, *$(* terminal denies — the test stayed green because it never read the new file. Hardening tests must enumerate the whole population (glob .opencode/agents/*.md and assert the terminal deny suffix on every frontmatter carrying a bash allow), not a named exemplar; otherwise each new persona re-opens the vector the test was written to close.
