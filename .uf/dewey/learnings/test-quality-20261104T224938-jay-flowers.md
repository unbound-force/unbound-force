---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-04T22:49:38Z
identity: test-quality-20261104T224938-jay-flowers
tier: draft
---

When a change deletes a shipped artifact, grep the test suite for tests that pin that artifact's content — SHA256 fixtures, testdata mirrors, inline duplicated constants. Tests that compare a testdata file against constants duplicated in the test body become tautological once the artifact is gone: they can only detect copy drift between the two duplicates, never a real regression. Observed in unbound-force PR #957 (stop-shipping-plugin-lock): TestPluginPackageInputs_RollbackPrerequisites kept pinning SHA256

> Note: payload truncated at store time (2026-11-04T22:49); the
> full text is unrecoverable — the Dewey graph page, git history
> (single commit 24b20b55), and dispatch artifacts all carry the
> same cut. Treat the surviving fragment as the record.
