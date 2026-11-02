---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:55:52Z
identity: issue-triage-20261102T035552-jay-flowers
tier: draft
---

Unbound-force command parity triage: verify "command lacks X" issues mechanically — grep .opencode/commands/ for "TodoWrite" and "checkpoint|get_progress", then read the named command in full. Gotcha: reference patterns drift between issue filing and triage. Issue #652 (filed 2026-10-05) cited uf.unleash.md's in-file execution checklist plus a mandatory TodoWrite section as the pattern to copy; by 2026-11-01 that pattern was superseded — uf.unleash.md now uses the checkpoint tool as the durable record of truth (SESSION-RESUME GUARD re-anchoring via get_progress + parse_tasks) with TodoWrite as a RECOMMENDED display aid, a pattern shared by uf.review-council, uf.fix-ci, and uf.address-feedback. uf.review-pr.md still has no progress mechanism at all (its session-resume guards cover only the fix-branch commit and review-posting human-confirmation gates). Triage outcome: the problem claim was valid, but the requested remedy was re-targeted to the current checkpoint-based pattern rather than the retired in-file checklist. Assess problem claims against current files AND re-target cited remedies to the current canonical pattern.
