---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-07T19:55:26Z
identity: supply-chain-20261107T195526-jay-flowers
tier: draft
---

An npm `overrides` security bump that crosses a major version over a transitive consumer's declared range (PR 948: toml 4.2.0 → 5.0.1 while effect 4.0.0-beta.83 declares ^4.1.1) voids the consumer's compatibility contract by construction. Unit and integration tiers exercise the plugin's own code paths, not the transitive consumer's internal usage of the overridden package, so validation of the range-crossing rests on the smoke tier exercising the real import chain plus OSV/CI corroboration (--fail-on-vuln, Trivy). Record the intentional range-crossing next to the override — consumer, declared range, GHSA list — so a later dependency refresh re-evaluates the override instead of silently inheriting a major skew between what a transitive package declares and what it runs against.
