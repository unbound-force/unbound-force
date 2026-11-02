---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:33:58Z
identity: issue-triage-20261102T013358-jay-flowers
tier: draft
---

Partial-overlap verification for docs-tracker vs tutorial issue pairs: when a docs-gate tracker enumerates specific behavior changes and a sibling tutorial issue (title prefix "tutorial:", label "tutorial") covers the same underlying change, map each enumerated tracker item against the tutorial's stated structure before ruling on duplication. Items inside the tutorial's structure overlap in subject only; items outside it prove the tracker retains a non-overlapping deliverable, making the complementary-vs-duplicate classification objective rather than judgment-based. Applied to unbound-force #766 (docs: five review-dispatch recovery/consolidation tool changes) vs #767 (tutorial: stranded-dispatch recovery workflow, five-step structure — symptoms, correlation_id resolution, recovery-tool choice, consolidate + warnings, finalize — covering three of #766's five items): dispatch_status per-run usage reporting and dispatch_agent_run metadata replace-on-supply fall outside #767's structure, confirming #766 tracks reference documentation the tutorial does not deliver. Disposition: complementary pair — cross-reference them; the tutorial owns the end-to-end workflow, the tracker owns reference detail. Reference docs produced under the tracker must preserve the correlation_id parameter asymmetry (redispatch_failed_runs and discard_orphaned_submissions require an explicit correlation_id; dispatch_status and consolidate_dispatch treat it as optional with latest-wins default).
