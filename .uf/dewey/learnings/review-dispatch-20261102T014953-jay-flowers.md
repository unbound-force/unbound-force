---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:49:53Z
identity: review-dispatch-20261102T014953-jay-flowers
tier: draft
---

Triage heuristic for dispatch-session contract conflicts (verified on unbound-force #746 → #794): when per-item dispatch+consolidate cycles fail deterministically on items 2+ within one conversation (INCONCLUSIVE / tier2_unavailable, "run has no included plan entry" or verdict mismatch at consolidation), suspect the dispatch tool's session-identity scoping before assessing individual items — a correlation ID keyed to the parent conversation plus first-call-only session_metadata persistence co-mingles every cycle into the first cycle's session directory. A recorded session-architecture constraint in workflow state (e.g. session_architecture_constraint in .uf/feedback/<pr>/state.json) is the tell. Verified resolution shape: mint a fresh correlation per aggregate dispatch invocation and re-point the session map before any session-creating side effect (fix-review-dispatch-correlation-release D1); persist session metadata replace-with-latest, never first-call-only (fix-review-dispatch-recovery D5); report the minted correlation in machine-readable output; and bind the command contract to capture and forward that correlation_id per cycle, scoped per item. General rule: an input schema must never silently accept a field the implementation ignores after first call — honor it or reject it.
