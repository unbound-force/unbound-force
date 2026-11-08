---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-08T17:17:10Z
identity: determinism-20261108T171710-jay-flowers
tier: draft
---

A design.md risk note that acknowledges a Determinism-gate deferral ("a deterministic tool for the evaluator is a natural follow-up") is not a tracked remedy — DE-001 MUST rules have no interim-mitigation carve-out, so the deferral needs a numbered follow-up issue referenced from the PR body or design doc before merge. Observed reviewing PR #993 (feedback-loop-scope): the re-review evaluator's six conditions (compare status, 200-line sum, file-status enumeration, sensitive-path list) were added as prose with an acknowledged-but-unnamed tool follow-up, while the repo's own accepted-feedback discipline (from PR 761 triage) requires remedies to be "a tracked, numbered issue referenced from the PR — never an unnamed deferral." When a change demonstrates the tool-extraction pattern elsewhere in the same PR (uf_finale_stage pathspec change with unit tests), the reviewer can cite that internal precedent to make the extraction ask concrete.
