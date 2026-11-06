---
tag: persona-design
author: jay-flowers
category: pattern
created_at: 2026-11-06T20:17:05Z
identity: persona-design-20261106T201705-jay-flowers
tier: draft
---

A persona ownership split is a cross-component contract, not a two-file edit. When a change moves a rule domain to a new exclusive owner ("you are the only persona that reports them"), every sibling persona that references that domain must be updated or given deferral language in the same change — find them by grepping .opencode/agents/ for the rule IDs (e.g., DE-0), not by reviewing only the personas named in the PR. In unbound-force PR #959, the DE-001–DE-008 domain moved from divisor-linguist to the new divisor-determinism; the linguist got deferral language, but divisor-guard.md still carries its own DE-008 "Deterministic Delegation Discipline" checklist, so councils dispatching both personas will emit duplicate DE-008 findings — the exact problem the split exists to eliminate. An exclusivity claim in a new persona is falsified by any sibling still carrying enforcement instructions for the same rules.
