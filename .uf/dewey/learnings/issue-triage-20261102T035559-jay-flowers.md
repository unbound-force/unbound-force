---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:55:59Z
identity: issue-triage-20261102T035559-jay-flowers
tier: draft
---

Parity issues that cite a sibling artifact's structure as precedent can go stale mid-migration — verify the cited precedent still matches HEAD before endorsing the literal asks. Verified on unbound-force issue #652 (uf.review-pr lacks execution checklist and TodoWrite tracking, filed 2026-10-05): it asks uf.review-pr.md to adopt uf.unleash.md's in-file execution checklist and make checklist state "the authoritative resumability indicator", but current uf.unleash.md no longer contains an in-file checklist — the uf-workflow-checkpoint migration made the checkpoint tool the durable record of truth, demoted TodoWrite to a RECOMMENDED display aid that never overrides durable state, and routes resumability through get_progress/parse_tasks. The reported gap itself stayed valid (uf.review-pr.md, 1075 lines, has no checklist, no TodoWrite section, and no checkpoint/get_progress integration, while uf.review-council, uf.fix-ci, and uf.address-feedback all pair TodoWrite with checkpoint tracking), so the triage verdict remained VALID — but the remedy should target the current checkpoint+TodoWrite-display pattern rather than the superseded prose checklist. Heuristic: one grep of the cited precedent file (plus its sibling commands) bounds both the claim and the current canonical pattern before the verdict is issued.
