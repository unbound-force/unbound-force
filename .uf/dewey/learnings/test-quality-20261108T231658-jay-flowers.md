---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-08T23:16:58Z
identity: test-quality-20261108T231658-jay-flowers
tier: draft
---

When a command-prose change is pinned by contract clause tests, derive the pins from the spec delta's full scenario text, not from the diff's most visible spans (table rows). Three obligations are routinely missed: (a) every AND-clause inside a Given/When/Then scenario is a separate pinning obligation — advisory-approval-event scenario 1's 'AND the advisories MUST remain in the review' maps to the sentence 'Advisories stay in the review as inline comments whichever event is used.' (uf.review-pr.md:1083-1084), which landed unpinned in PR #990; (b) a scenario implemented by a sentence appended to a neighboring paragraph guarded by a DIFFERENT pre-existing test (scenario 4's 'This includes an advisory verdict that maps to `APPROVE`.' added to the self-authored paragraph covered by TestCommandContracts_ReviewPrSelfAuthoredMapping) still needs its own fragment pin — the older test's generic pin does not cover the new disambiguating sentence; (c) retired conflicting prose (the old '| APPROVE WITH ADVISORIES | `COMMENT` |' table row) needs an absence assertion, which is precise when the retired fragment does not substring-match the replacement rows after normalizeCommandText. Mechanical review check: extract every scenario bullet (including AND clauses) from the spec delta, grep the command-contract tests for a pin per bullet, and treat any unpinned bullet as an open silent-regression window — the same enumeration discipline as learning/test-quality-20261105T143503, extended to sub-clauses and cross-paragraph implementing sentences.
