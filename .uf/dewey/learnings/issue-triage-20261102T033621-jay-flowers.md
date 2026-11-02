---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:36:21Z
identity: issue-triage-20261102T033621-jay-flowers
tier: draft
---

Structural-refactor issues that cite exact line counts and inline-block ranges in command or agent files go stale when later architecture changes land. Before validating such an issue, verify its premise against current HEAD (actual file length, presence of the described inline block, whether the referenced skills are now loaded via the skill tool) and check openspec/changes/ history for superseding refactors. Precedent: unbound-force issue #662 proposed extracting a ~500-line inline subagent prompt from uf.review-pr.md (claimed 1,239 lines) into a review-pr-worker.md agent, but the review-pr-divisor-fan-out change had already removed the single-general-subagent pattern: HEAD is 1,075 lines with no inline Task-tool prompt, pre-flight and review-context are invoked via the skill tool, and divisor-* child agents each read their own .opencode/agents/ definition with prompts assembled by dispatch_review_runs. The issue's goal (stop re-tokenizing an inline prompt) was achieved by a stronger mechanism, and its proposed single-worker agent would have reintroduced the removed pattern.
