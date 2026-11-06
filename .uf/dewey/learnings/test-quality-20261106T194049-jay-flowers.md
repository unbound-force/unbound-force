---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-06T19:40:49Z
identity: test-quality-20261106T194049-jay-flowers
tier: draft
---

When a spec MUST clause is stated at the command level (e.g. "uf init refreshes an existing manifest and MUST keep X"), an acceptance test must drive the command's Run() entry point with the pre-existing state, not only the extracted helper. Function-level tests of the helper leave its wiring site — the guard condition, the relPath match, and the placement before the content-diff classification — unprotected: removing the three-line wiring keeps every test green while breaking the scenario. Precedent: stop-shipping-plugin-lock tested preservePluginDevDependencies deeply at the function level but left the scaffold.go deploy-loop invocation uncovered for the devDependencies-bearing case, while the sibling foreign-lock scenario in the same file got a full Run()-level composition test (TestSCFR003_ForeignLockSurvivesRun) — the asymmetry is the gap.
