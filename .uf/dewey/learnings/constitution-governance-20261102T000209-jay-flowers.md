---
tag: constitution-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:02:09Z
identity: constitution-governance-20261102T000209-jay-flowers
tier: draft
---

Issue #824 (and the validated compiled/constitution article) misattributes the four-principle enumeration to .opencode/commands/constitution-check.md — that path does not exist; the command file uf.constitution-check.md is a thin orchestrator with no principle list. The enumeration lives in .opencode/agents/constitution-check.md (line 43 "For each of the four org principles (I, II, III, IV)", line 95, line 99 "count, always 4", line 108 decision criteria "All four findings"), to which the command delegates; the cited line numbers 43/95 match the agent file exactly. A fix must also sync the tool-owned scaffold mirror at internal/scaffold/assets/opencode/agents/constitution-check.md per the drift-detection convention (embedded assets MUST match canonical sources). The durable fix per Principle VI / determinism pack DE-001 is deriving the principle list from the constitution document at check time instead of hardcoding a count — the hardcoded count is what let the template go stale through two amendments (V Security by Default in v1.3.0, VI Deterministic Execution in v1.4.0).
