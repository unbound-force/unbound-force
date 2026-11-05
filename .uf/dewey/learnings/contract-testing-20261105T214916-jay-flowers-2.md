---
tag: contract-testing
author: jay-flowers
category: pattern
created_at: 2026-11-05T21:49:16Z
identity: contract-testing-20261105T214916-jay-flowers-2
tier: draft
---

When a fix commit reworks multiple instruction sites in an agent command file, sweep the contract-test clause table against the fix-site inventory, not just the command prose and its scaffold mirror: every reworked site whose regression would reintroduce the original failure needs a presence clause (or ordering pin) in the same increment. Presence clauses pin what phrases say and ordering pins pin sequence, so a corrected reordering that ships without its clause is invisible to the suite. Precedent: PR #955's fix commit reworked the SESSION-RESUME GUARD and TodoWrite resume text in uf.address-feedback.md — the documented re-entry vector for issue #954 — while adding clauses only for the Section 1.2.1 and 4.4 fix sites; reverting the guard to an unconditional pre-checkout get_progress passes every contract test because the ordering pin anchors on the Section 1.2.1 heading and the first bold **CHECKPOINT** marker, neither of which moves. Repo precedent for the missing lock shape: uf.fix-ci's 'session-resume guard' clause group in command_contract_test.go and review-council's 'session-resume guard pinned' test in scaffold_test.go. The sweep is mechanically discoverable by diffing the fix commit's changed command phrases against the clause table's fragments.
