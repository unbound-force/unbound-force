---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:47:39Z
identity: issue-triage-20261102T014739-jay-flowers
tier: draft
---

Implementation follow-up issues become self-verifying at triage time when the predecessor design decision explicitly designated the follow-up. D4 of openspec/changes/fix-triage-issue-dispatch-docs declared the consolidate-payload assessment-fields extension a Non-Goal "(follow-up, task 5.4)" and pre-specified that "the tool's backend can switch from persisted-text parsing to payload fields without any command change"; issue #755 (unbound-force/unbound-force) quotes that lineage, and a Dewey search of the indexed design doc corroborates every factual claim (persisted-text backend via PersistedRunData.text, preserved validation bounds, no-command-change decoupling). Triage pattern: (1) verify follow-up issues against the cited design doc's follow-up designation before assessing validity; (2) when the issue states the command-layer contract is preserved by design, the change is internal for documentation-gate purposes — exempt from new docs/blog/tutorial issues under the internal-refactoring exemption, with the predecessor's docs-gate issue cross-referenced rather than re-filed; (3) categorize as enhancement, not feature, because no user-visible capability changes; and (4) implementation-location options in the sketch ("run record OR submit_review_findings/consolidation projection") belong to the spec/design phase and do not warrant a split recommendation.
