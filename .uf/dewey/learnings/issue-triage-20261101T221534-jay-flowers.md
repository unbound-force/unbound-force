---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:15:34Z
identity: issue-triage-20261101T221534-jay-flowers
tier: draft
---

Tutorial issues filed to correct a sibling-disposition asymmetry triage objectively when the body carries the full precedent chain: the mirrored sibling tutorial issue number, the significance-threshold justification, and the re-disposition provenance (originating PR, review round, thread). Verified on unbound-force/unbound-force#870 (/uf.fix-ci tutorial, triaged 2026-11-01): prior learnings content-pipeline-20261029T115034-jay-flowers (asymmetry detected — opsx/uf-fix-ci declined the tutorial with rationale anchored in docs issue #862 while sibling /uf.rebase-pr carried open tutorial issue #855, one day earlier) and content-pipeline-20261029T154738-jay-flowers (tutorial disposition re-anchored to live issue #870 after PR #868 round-1 review) independently corroborate the body's origin claim, converting the provenance from assertion to verified record. Duplicate-analysis corollary: the keyword-search candidates that surface for such issues are typically the docs-gate deliverables that anchored the original decline (#862 command reference docs, #873 tooling/re-route docs) — by construction a different artifact type than the structured walkthrough the tutorial issue tracks, so they are complementary related issues to cross-reference, not duplicates. Triage recipe for disposition-correction issues: verify the precedent sibling issue exists, verify the re-disposition record (PR/review thread or stored learning), then classify on the artifact type the correction delivers rather than re-litigating the original decline.
