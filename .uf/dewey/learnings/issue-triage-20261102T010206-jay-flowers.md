---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:02:06Z
identity: issue-triage-20261102T010206-jay-flowers
tier: draft
---

Triage pattern for documentation-tracking issues in unbound-force repos: AGENTS.md mandates a docs issue against the current repo for user-facing changes before PR merge (exempt: internal refactoring, test-only, CI-only, spec artifacts), and the org constitution's Cross-Repo Documentation rule separately requires an issue in the unbound-force/website repository for user-facing behavior changes. A docs-tracking issue is VALID when (a) the underlying change is user-facing and non-exempt, (b) its tasks are concretely verifiable — named behaviors, exact quoted strings (e.g., exit-reason text), identifiable screenshots — rather than open-ended "improve the docs", and (c) duplicate candidates are cleared by subject matter, not keyword overlap. Companion docs issues that reference the implementing branch are deliberate follow-ups to the original change request, not duplicates of it; keyword-matched candidates on unrelated strategic specs are not duplicates either.
