---
tag: dewey-learnings
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:35:29Z
identity: dewey-learnings-20261107T163529-jay-flowers
tier: draft
---

Extends dewey-learnings-20261016T161248 (same-session overlap scan): file-based retrospective learning batches have NO mechanical dedup at all — the dispatch consolidation's near-identical proposal merge (duplicate_proposal_merged) sees only proposals inside one dispatch, so an insight recurring across capture sessions, review iterations, or personas lands once per occurrence. PR #968's 50-file batch carried seven clusters totalling 22 files restating seven insights, several written seconds apart under different tags in one session, plus a cross-session re-capture (scaffold-mirrors-20261105T174838 vs learning/scaffold-mirrors-20261025T145313 from PR #815 — the same insight's third capture cycle after the 2026-10-16 scaffold-mirrors/scaffold-parity pair). Practice: before persisting a learning batch, dedupe both within the batch AND against the existing graph (search the insight's distinctive phrases, not just its tag); when collapsing a cluster, retain the fullest complete statement; and where a redundant copy is also a mid-sentence truncated body, delete it rather than complete it — PR #968's code-review-20261103T211847/48 and test-quality-20261104T225837 are simultaneously dedup redundancies and truncation victims, so deletion resolves both findings at once and shrinks the body-completion workload from 15 files to 12.
