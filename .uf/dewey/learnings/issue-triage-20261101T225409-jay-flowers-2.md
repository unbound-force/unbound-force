---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:54:09Z
identity: issue-triage-20261101T225409-jay-flowers-2
tier: draft
---

Unbound Force triage pattern for determinism trade-offs: when a merged change keeps mechanical steps as command prose under a documented Non-Goal, the follow-up tracking issue must enumerate each deferred step as a checkbox, cite the governing rule by identifier (determinism.md DE-001), and name the target plugin and test suite — this satisfies the zero-waste rule's requirement that documented trade-offs map to actionable work and gives triage a verifiable citation instead of prompt archaeology. Security corollary from the Adversary domain: the deferred steps typically named in such issues (parsing delegation return text, git status checks, gh pr view/edit choreography) are dual liabilities — DE-001 violations AND unvalidated-input/resilience risks — because prose-instructed parsing drifts after context compression and carries no schema validation or fail-closed error handling; tool extraction remediates both at once. Exemplar: issue #842 tracking PR #839's update-run-pipeline-plan deferral (four extractions into the uf-workflow plugin's TypeScript test suite).
