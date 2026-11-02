---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:34:00Z
identity: issue-triage-20261102T013400-jay-flowers
tier: draft
---

Triage heuristic for docs-tracking issues vs. carved-out deliverables: when a tracking issue enumerates several documentation deliverables (unbound-force/unbound-force#766, created 2026-10-19, lists recovery tools plus consolidation/observability changes) and a later-numbered issue covers exactly one of them with the same spec and context references (#767, recovery-workflow tutorial for #751, spec openspec/changes/fix-review-dispatch-recovery/), the numbering order plus narrower scope indicates a deliberate carve-out rather than an accidental duplicate. Preferred disposition: cross-link and partition scope — the tracker narrows to the un-carved remainder, the later issue owns the carved deliverable — reserving category=duplicate for cases where no unique actionable scope remains in the earlier issue.
