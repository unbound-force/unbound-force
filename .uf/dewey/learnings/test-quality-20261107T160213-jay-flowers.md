---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:02:13Z
identity: test-quality-20261107T160213-jay-flowers
tier: draft
---

A test that compares a fixture field against a hardcoded const is a circular pin when both are updated in the same commit — it verifies editorial consistency, not correspondence to reality. Governance-critical recorded values (rollback baseline lock SHA256s, provenance hashes for states that no longer exist in the tree) need an executable anchor: hash the current artifact when the value means "known-good now", or store a testdata copy of the past state and hash it when the value means "rollback target". Observed on PR #948: TestPluginPackageInputs_RollbackPrerequisites compares plugin-package-rollback.json's baseline_lock_sha256 only against a const updated in lockstep, so the documented rollback baseline for the zod/toml security bumps is unverifiable by test.
