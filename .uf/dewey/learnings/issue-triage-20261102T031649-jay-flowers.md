---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:49Z
identity: issue-triage-20261102T031649-jay-flowers
tier: draft
---

Triage nuance for prose→tool extraction issues with stale citations after sibling migrations: verify the residual target, not the citation. Issue #668 (resolve_conflict tool for uf.finale conflict recovery) cited uf.finale.md lines 625-953 (~146 lines), but current HEAD is 399 lines — steps 1-8 already migrated to uf_finale_* plugin tools and the 5-option conflict flow sits at §7b lines 302-348. The verdict keys on two checks: (1) the mechanical prose still exists (it does), and (2) a DE-005 intent-level collapse remains available — distinguish primitive-level tooling (the prose already calls git_fetch/git_merge/git_rebase individually) from sequence-level orchestration (the option→command-sequence mapping is still prose; once the human picks an option, execution is judgment-free). Primitive tools being present does not satisfy DE-005; the sequencing is the extraction target. Stale line ranges and token-savings estimates are re-baselined at spec time (DE-004 actuals), not treated as NEEDS-CLARIFICATION blockers. Corollary: openspec/changes/finale-conflict-resolution is the earlier option-5 sub-agent addition (refs #325, pre-rename finale.md) — a scope-mate in the same flow, not a duplicate of the extraction issue.
