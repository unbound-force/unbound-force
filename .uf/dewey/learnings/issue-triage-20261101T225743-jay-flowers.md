---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:57:43Z
identity: issue-triage-20261101T225743-jay-flowers
tier: draft
---

Structural validation for tutorial issues in unbound-force repos: a tutorial issue's section outline is mechanically checkable against the command contract it cites. A sound outline maps one-to-one onto the contract's phases plus an audience-entry section for symptom recognition; sections inventing phases absent from the contract signal drift, and omitted contract phases signal incompleteness. Verified on /uf.rebase-pr tutorial issue #855: its seven sections (symptom recognition + argument pre-validation, gating checks, local checkout/rebase, conflict resolution with git_rebase_abort backout, push with git_push_force_with_lease and the head-branch-only rule, gh mergeable-state verification) map onto the six-phase contract in .opencode/commands/uf.rebase-pr.md, and every tool named in the outline exists in the toolchain. This converts "is the tutorial well-scoped" from an impression into an outline-to-contract cross-check at triage time, complementing the tutorial-vs-docs-gate duplicate test (learning/issue-triage-20261101T221541-jay-flowers).
