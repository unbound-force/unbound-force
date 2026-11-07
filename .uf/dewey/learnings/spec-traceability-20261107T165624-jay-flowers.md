---
tag: spec-traceability
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:56:24Z
identity: spec-traceability-20261107T165624-jay-flowers
tier: draft
---

A MUST clause implemented as a wiring call inside an orchestrating function needs composed (orchestrator-level) coverage, not just unit coverage of the extracted helper: the guard condition, the call's ordering relative to skip/idempotency comparisons, and the write-group reporting are the clause's observable behavior, and all three can regress while helper unit tests stay green. Review heuristic: enumerate a requirement's MUST clauses and check each maps to a composed test; the clause whose spec delta carries no Given/When/Then scenario is typically the clause without one, and unequal test depth between clauses of the same requirement inside one PR is the smell. Precedent: PR 957 SC-FR-003 — the foreign-lock and fresh-target clauses received Run()-level tests (TestSCFR003_ForeignLockSurvivesRun, TestSCFR003_FreshTargetInputsDoNotActivatePlugins) while the devDependencies-preservation clause had only direct unit tests of preservePluginDevDependencies, leaving the scaffold.Run deploy-loop merge unverified.
