---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:14:11Z
identity: issue-triage-20261102T021411-jay-flowers
tier: draft
---

Triage recipe for documentation-gate issues in unbound-force/unbound-force: (1) Verify provenance through the originating change's artifacts — determinism-codification task 1.11 and the PR #723 summary explicitly record "#720" and "unbound-force/website#305" as the filed docs-gate issues, confirming the issue discharges the AGENTS.md documentation gate for a merged change. (2) Verify the gap directly in the enumerated docs — docs/architecture.md's convention-pack listing (~lines 218-236) already includes non-coding governance packs (severity, content, ci, review-voice) while omitting determinism.md, so the acceptance criterion's "explicitly scope to coding packs" escape hatch is unavailable there and the list must be extended. (3) Distinguish duplicates by the pack's rule range at filing time, not by pack name — the determinism pack grew from DE-001..DE-004 (determinism-codification, tracked by #720) to DE-001..DE-008 through later changes each carrying its own docs issue (#762 DE-005/DE-006 router-pattern, #781 DE-007 seam-verification, #791 DE-008 delegation), so a smaller rule range in an issue body is a time-anchored snapshot, not a duplicate or conflict signal.
