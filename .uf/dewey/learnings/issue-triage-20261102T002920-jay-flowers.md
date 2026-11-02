---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:29:20Z
identity: issue-triage-20261102T002920-jay-flowers
tier: draft
---

In issue triage, an explicit umbrella reference in an issue body ("Umbrella: #N") marks the issue as an intentional phase-tracked child of a program umbrella. Treat that reference as the decomposition signal: the child tracks one milestone of a multi-phase program and is not a duplicate of its umbrella, even when title keywords overlap heavily. Reserve the duplicate verdict for issues that cover the same milestone as another issue without parent linkage. Corroborate by confirming the referenced umbrella exists and that the child's scope is one phase (a strict subset) of the umbrella's program — e.g., #806 "Phase 4 — Dogfood" versus umbrella #797 "OpenCode v2 migration — org-wide orchestration".
