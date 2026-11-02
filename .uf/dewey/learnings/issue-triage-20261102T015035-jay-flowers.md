---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:50:35Z
identity: issue-triage-20261102T015035-jay-flowers
tier: draft
---

Deferred code-review-council refactor follow-ups triage objectively when the issue carries full provenance — architect run id, dispatch artifact correlation id, maintainer deferral decision, and named symbols with line references; verify symbols by grep, the extraction precedent in-repo (.opencode/lib/ pattern), and current file size by direct read. For unbound-force plugin-suite module extractions, triage completeness requires the issue to acknowledge the byte-identical scaffold mirror under internal/scaffold/assets/opencode/ and registration in scaffold_test.go / plugin_activation.go — an extraction proposal omitting mirror awareness is incomplete. Unenforced deferrals compound fast: .opencode/plugins/uf-workflow/index.ts was 3,700+ lines when issue #741 was filed (2026-10-16), 4,902 lines nine days later (compiled/plugin-architecture learning), and 5,402 lines at triage time — deferred structure findings need scheduled implementation, not just a VALID verdict.
