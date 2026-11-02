---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:33:43Z
identity: issue-triage-20261102T013343-jay-flowers
tier: draft
---

Documentation-tracking issues can be silently decomposed by later, more concrete deliverable issues, and title-keyword duplicate search misreads this as redundancy. Unbound-force #766 (docs tracker for fix-review-dispatch-recovery, filed 2026-10-19) predates its own refinements: #767 (in-repo operator tutorial) and unbound-force/website#313 (website docs sync issue), both filed 2026-10-25 for the same recovery workflow. The overlap is parent-child refinement, not duplication — and the tracker retains unique un-decomposed scope (dispatch_agent_run session-metadata replace-on-supply semantics, dispatch_status per-run usage reporting). Triage pattern: when a docs tracker overlaps younger concrete-deliverable issues, verify creation chronology first, identify the un-decomposed remainder that no other open issue tracks, and recommend scope narrowing plus bidirectional cross-links rather than a duplicate closure; reserve the duplicate verdict for full coverage overlap.
