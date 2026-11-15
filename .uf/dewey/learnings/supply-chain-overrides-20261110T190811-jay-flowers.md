---
tag: supply-chain-overrides
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:08:11Z
identity: supply-chain-overrides-20261110T190811-jay-flowers
tier: draft
---

Adding a security override to the plugin manifest in unbound-force touches more surfaces than the two package.json files: the canonical .opencode/package.json, the scaffold mirror (byte-locked by TestEmbeddedAssets_MatchSource's shippedPluginManifest transform), the provenance JSON, BOTH test constant sets (TestPluginPackageInputs_ExactVersions' expectedOverrides and the provenance test's expected struct plus its verified_at literal), the runbook's nested-copy confirm command (npm ls ...) in docs/osv-transitive-scan.md, and the runbook's override enumeration in prose. The doc surfaces are the easy ones to miss: they live outside code, so no drift test catches their staleness — the confirm recipe still listed only zod and toml after the effect and uuid overrides were added.
