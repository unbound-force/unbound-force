---
tag: contract-testing
author: jay-flowers
category: pattern
created_at: 2026-11-05T21:49:16Z
identity: contract-testing-20261105T214916-jay-flowers-2
tier: draft
---

When a fix commit reworks multiple instruction sites in an agent command file, sweep the contract-test clause table against the fix-site inventory, not just the command prose and its scaffold mirror: every reworked site whose regression would reintroduce the original failure needs a presence clause (or ordering pin) in the same increment. Presence clauses pin what phrases say and ordering pins pin sequence, so a corrected reordering that ships without its clause is invisible to the suite. Precedent: PR #955
