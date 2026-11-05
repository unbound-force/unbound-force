---
tag: zero-waste
author: jay-flowers
category: gotcha
created_at: 2026-11-04T22:49:36Z
identity: zero-waste-20261104T224936-jay-flowers
tier: draft
---

When a change deletes an artifact that a test/testdata pair verifies, audit the retained test for self-reference — a test asserting fixture constants against identical constants hardcoded in the test body protects nothing. In PR 957 (stop-shipping-plugin-lock), design decision D5 retained TestPluginPackageInputs_RollbackPrerequisites and internal/scaffold/testdata/plugin-package-rollback.json as \"a historical record\": the test pins SHA256s of lock files the same change deleted from every location, so the ha
