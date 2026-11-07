---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:37:56Z
identity: test-quality-20261107T173756-jay-flowers
tier: draft
---

A testdata field compared against a same-commit test const is a drift check between two files, never a truth anchor: both locations are updated together in the same commit, so CI stays green while the recorded value drifts from the artifact it claims to describe -- a circular pin. Anchor values derivable from repo state by computing them in the test (read the artifact, hash it, compare). Values that record historical state (rollback targets, pre-bump baselines) cannot be hashed from the working tree: record the state's provenance (e.g., the commit SHA that holds it) beside the digest so a test can materialize the historical artifact and verify it. Instance: internal/scaffold testdata plugin-package-rollback.json's baseline_lock_sha256 is documented (fix-triage-issue-dispatch-docs) as a rollback-target anchor but is verified only by the circular const comparison in TestPluginPackageInputs_RollbackPrerequisites; before an anchor test could be written correctly, the field's contract had to be re-derived from the landed change's design docs because nothing at the point of use states which lock state the digest records.
