---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:15:55Z
identity: issue-triage-20261102T001555-jay-flowers
tier: draft
---

Documentation-gate issues in unbound-force repos (filed per AGENTS.md's "MUST file a documentation issue against the current repo for user-facing changes before PR merge", typically the final task of an OpenSpec/Speckit change) can be triaged against four machine-verifiable evidence points: (1) the referenced change's tasks.md shows the documentation-issue task and its CHANGELOG predecessor both marked complete; (2) the CHANGELOG entry landed and its text matches the behavior changes the issue enumerates; (3) grep of every enumerated doc target shows zero coverage of the new behavior, confirming the gap is open rather than already closed; (4) for user-facing behavior, the complementary unbound-force/website issue exists and is explicitly scoped as complementary (constitution Cross-Repo Documentation rule), so the in-repo issue is not a cross-repo duplicate. Issue #808 (docs for fix-review-dispatch-correlation-release, task 5.5) passed all four checks. Additional nuance: conditional checklist items phrased "update any text implying X" are verify-then-maybe-skip items when the target file carries no related prose at all — the unbound-force root README had zero review-dispatch coverage, so that item resolves to a no-op or small addition rather than a correction; do not read such items as claims that stale text exists.
