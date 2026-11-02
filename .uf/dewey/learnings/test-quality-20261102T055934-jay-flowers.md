---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:59:34Z
identity: test-quality-20261102T055934-jay-flowers
tier: draft
---

Exact-content tests that compare seeded output against the same Go constant the code writes are tautological content guards: they pin the write path, not the constant's required properties. In the unbound-force scaffold, TestRun_OverrideStubs_FreshRunCreatesExactContent asserts file == modelMatrixOverrideStub, so it cannot catch a constant edit that breaks the comment-only/YAML-null invariant SC-FR-011 depends on — the property that lets seeded per-developer overrides load as no-ops with zero loader changes. Loader tests elsewhere (model-matrix.test.ts) verify comment-only semantics only with their own fixtures, so nothing pins the shipped constants. Rule of thumb: whenever a seeded constant must satisfy an invariant (comment-only/YAML-null, no duplicate entries, must-not-contain-X, must-contain-Y), add a property test over the constant itself — a line-prefix check or a parser round-trip — separate from the exact-content wiring test. The in-repo precedent is TestEnsureGitignore_FreshDir asserting .gaze/ absence and pattern presence directly against the written block rather than relying only on the repo-drift equality test.
