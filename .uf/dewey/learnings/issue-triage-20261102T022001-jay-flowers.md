---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:20:01Z
identity: issue-triage-20261102T022001-jay-flowers
tier: draft
---

When triaging documentation-tracking issues in unbound-force, verify three things cheaply before ruling: (1) the cited spec directory exists (glob openspec/changes/<name>/ for proposal.md, design.md, tasks.md, specs/); (2) the docs gap is real — grep docs/ for the tool/artifact name, and note that internal retrospective notes (e.g. docs/learnings-review-*.md) do not count as user-facing reference coverage; (3) the named tool is still the registered mechanism of record. On (3): tool names drift when parallel changes merge — persist_unleash_summary was absorbed into write_unleash_summary as the mechanism of record, with createPersistUnleashSummaryTool kept as a delegating compat alias so older citations stay resolvable. Such drift is a note for the doc author (document the current registered name and full schema, mention the compat alias), not grounds for NEEDS-CLARIFICATION or INVALID. Also: a referenced unbound-force/website#NNN issue is the constitution's Cross-Repo Documentation companion to the in-repo docs issue, not a duplicate — the pair is the governance-mandated pattern for user-facing changes.
