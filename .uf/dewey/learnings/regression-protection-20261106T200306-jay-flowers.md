---
tag: regression-protection
author: jay-flowers
category: pattern
created_at: 2026-11-06T20:03:06Z
identity: regression-protection-20261106T200306-jay-flowers
tier: draft
---

Duplicated verification-critical constants across pipelines (Makefile PLUGIN_TEST_NPM vs the ci_security.yml osv_plugin_transitive inline npm pin) drift silently when coupled only by "keep the sync" comments. The unbound-force pattern for this class is an executable drift test that reads both files and asserts value equality — the same mechanism as TestEmbeddedAssets_MatchSource and the provenance-to-manifest cross-checks. Prefer that over prose sync obligations for any constant that gates two verification pipelines.
