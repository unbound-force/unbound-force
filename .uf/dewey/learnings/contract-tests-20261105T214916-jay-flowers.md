---
tag: contract-tests
author: jay-flowers
category: gotcha
created_at: 2026-11-05T21:49:16Z
identity: contract-tests-20261105T214916-jay-flowers
tier: draft
---

A bespoke ordering pin anchored on the first structural marker (the first **CHECKPOINT** block) and resume-path phrase-presence clauses are complements, not substitutes — skipping the second leaves the historically defect-prone path unlocked. The ordering pin locks the main flow's first gated call, but early-file conditional resume prose (session-resume guards, TodoWrite re-initialization) legitimately mentions the governed tool before the gate, so the pin cannot see it; resume prose executes exactly when the corrected rule is out of context, making it the most likely regression path. After adding a strings.Index ordering pin for a branch-keyed progress invariant, also add phrase-presence clauses keyed on the resume-path correction's unique text (e.g. 'the re-check is unconditional on resume', 'first confirm the working tree is on the PR head per Section 1.2.1') so reverting the resume-guard wording fails the suite. Presence clauses are safe here even though the ordering pin is not: those phrases exist only in the corrected text, so a Contains check cannot false-positive on legitimate earlier tool mentions. Precedent: PR #955's fix commit added the ordering pin plus five clause groups covering Section 1.2.1 and 4.4, while the SESSION-RESUME GUARD and TodoWrite resume corrections — issue #954's documented reopen vector — stayed unpinned and a revert of them passes the whole suite.
