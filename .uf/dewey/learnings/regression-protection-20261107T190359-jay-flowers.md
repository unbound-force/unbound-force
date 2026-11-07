---
tag: regression-protection
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:03:59Z
identity: regression-protection-20261107T190359-jay-flowers
tier: draft
---

A parity test that hardcodes the pinned value as a literal guards a duplicated pair's twin-ness but not its currency against the canonical pin. When sibling surfaces in the same pin graph are guarded by tests that read their files dynamically (Makefile-to-workflow, workflow-to-docs), a future one-sided bump updates every dynamically-read surface, passes those tests, and leaves the hardcoded-literal pair — and the test asserting it — silently stale at the old version. Precedent: TestRemediationConstants_CarryNpm11EscapeHatch (internal/scaffold/scaffold_test.go, PR #957 feedback fixes) asserts the literal "npx -y npm@11.6.2" while TestPluginTestNpmPin_TwinsCISecurityPin and TestOsvRecipeDoc_PinsMatchCanonicalJob derive their expectations from the files they sweep, so the four consumer-facing remediation constants are the only npm-pin surface that can drift from PLUGIN_TEST_NPM without a failing test. Rule: in a pin-parity graph, every guarded edge derives its expected value from the canonical pin source (parse the Makefile variable that owns the pin), never from a literal copy of the value under test.
