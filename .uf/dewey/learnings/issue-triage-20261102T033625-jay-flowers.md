---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:36:25Z
identity: issue-triage-20261102T033625-jay-flowers
tier: draft
---

Triage pattern for extraction/refactor issues in unbound-force repos: validity rests on four checkable evidences. (1) The duplication target exists as a shared in-repo asset — e.g., issue #662 targets a ~500-line inline subagent prompt in uf.review-pr.md that re-describes the existing pre-flight and review-context skills under .opencode/skills/. (2) Measurable efficiency deltas are cited per determinism pack DE-004 — line counts (1,239 to ~700) and token estimates (~4,000), not impressions. (3) Acceptance criteria include an output-parity contract ("review output format and quality unchanged"), the correct preservation clause for behavior-preserving refactors. (4) Duplicate screening distinguishes same-file-different-concern issues (#652 execution checklist vs #662 prompt extraction, both touching uf.review-pr.md) from true duplicates — same command surface does not imply duplication. Issues matching this pattern align with Constitution Principle VI (Deterministic Execution) and triage as VALID / enhancement / objective.
