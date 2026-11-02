---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:35Z
identity: issue-triage-20261102T011435-jay-flowers
tier: draft
---

Documentation-gate record issues in unbound-force repos are VALID even when every disposition is "no update needed" or "exempt": the AGENTS.md gate mandates filing only for user-facing changes, but the change's own tasks.md routinely requires the record (seam-verification task 4.4 required metrics.md to cite the issue number), so validity rests on traceability to a change task, fully resolved disposition lines (a Website line may close as an explicit not-applicable decision with basis and precedent — #781 cited the determinism-codification exclusions and #762's identical classification), and a defined close condition (PR merge). Triage corollary: docs/pack keyword matches across sibling determinism-pack records (#762 for DE-005/DE-006, #781 for DE-007, #791 for DE-008) are topical matches, not duplicates — disambiguate on the referenced change name and branch before marking duplicate.
