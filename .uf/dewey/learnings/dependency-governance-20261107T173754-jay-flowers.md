---
tag: dependency-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:37:54Z
identity: dependency-governance-20261107T173754-jay-flowers
tier: draft
---

A testdata artifact-hash record enforced only by comparison against a test constant updated in the same commit is a circular pin: the test verifies copy-consistency, never artifact correspondence. Whole-artifact hash records for current-state baselines (e.g. plugin-package-rollback.json's baseline_lock_sha256, the tracked .opencode/package-lock.json SHA mandated by SC-FR-003 "Reproducible Plugin Dependencies") must be anchored by a test that hashes the real tracked artifact and compares — the same pattern TestPluginPackageInputs_ExactVersionsAndLockIntegrity already applies per-package for versions and sha512 integrities. Hash the canonical tracked file, not the scaffold mirror or embedded copy: drift tests already enforce byte-identity between canonical, mirror, and embedded copies, so one canonical anchor covers all three. Historical rollback-target hashes (the pre-hardening lock SHA fef6f7f1…) have no in-tree artifact and legitimately stay declarative. Discovered assessing PR 948 feedback thread PRRT_kwDORVoNbc604LsB (uf-divisor, LOW assertion-depth).
