---
tag: dependency-governance
author: jay-flowers
category: pattern
created_at: 2026-11-07T20:50:42Z
identity: dependency-governance-20261107T205042-jay-flowers
tier: draft
---

A testdata hash record protected by both an executable artifact anchor and a same-commit test const is not a circular pin — the two checks verify different contracts, and the const should survive the anchoring. The artifact anchor (read the canonical tracked file, hash it, compare against the testdata field) verifies correspondence: the record describes the real artifact. The retained const verifies identity: the artifact is the specific known-good state, so moving the pin requires two visible same-commit edits (testdata field + test const) instead of one quiet pair. Removing the const as 'redundant after anchoring' weakens governance visibility: a lock change plus a testdata update would then pass without touching test source. Instance: PR 948 round-4 TestPluginPackageInputs_RollbackPrerequisites anchors plugin-package-rollback.json's baseline_lock_sha256 by hashing .opencode/package-lock.json while retaining const baselineLockSHA256 — both checks belong in the gate.
