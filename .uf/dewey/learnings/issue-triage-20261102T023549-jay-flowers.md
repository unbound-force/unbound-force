---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:35:49Z
identity: issue-triage-20261102T023549-jay-flowers
tier: draft
---

A docs-tracking issue that stays open past subsequent changes to the same surface keeps its validity but can carry a stale narrative. unbound-force issue #706 (document checkpoint/get_progress/clear_progress, filed 2026-10-12 for merged PR #705) frames the tools as replacing the fragile execution-checklist pattern with reliable JSON state, while the validated compiled/uf-workflow article records the surface's later evolution: filesystem markers (tasks.md checkboxes, dispatch artifacts) remain authoritative for resumability, checkpoint state is keyed by git branch at write time so a mid-workflow branch switch strands it, and implement-phase pass/fail truth lives in phases[N].verdict rather than completed_steps. Triage rule: corroborate a long-open docs issue's framing against the current validated state (compiled articles, merged follow-up changes to the same surface, such as the tool-args-schema-mismatch follow-up tracked by #740) and cite that source for the docs author. Staleness narrows the issue's framing; it does not invalidate the tracker.
