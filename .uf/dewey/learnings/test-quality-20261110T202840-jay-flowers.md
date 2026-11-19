---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:28:40Z
identity: test-quality-20261110T202840-jay-flowers
tier: draft
---

Derive expected twin values from the canonical source at test runtime instead of hardcoding a third copy. TestSCFR004_OfflineInstallFailure extracts the PLUGIN_TEST_NPM pin from the Makefile via regex and asserts the remediation output contains that exact string, so the test enforces twin synchronization (Makefile pin == remediation constant) without duplicating the pin value; a bump on either surface fails the test until both are reconciled. This complements a hardcoded-expectation twin test (TestRemediationConstants_CarryNpm11EscapeHatch pins the literal): the derived form catches Makefile-vs-remediation drift, the literal form catches remediation-text regressions, and together they leave no copy that can silently go stale.
