---
tag: drift-detection
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:43:35Z
identity: drift-detection-20261110T204335-jay-flowers
tier: draft
---

Plugin-pin lockstep changes in unbound-force (canonical .opencode/package.json + scaffold mirror + TestPluginPackageInputs_ExactVersions + plugin-package-provenance.json) are guarded by independent exact-pin tests rather than one cross-surface check: TestPluginPackageInputs_ExactVersions pins the canonical overrides map bidirectionally, TestEmbeddedAssetsMatchSource pins the scaffold asset to canonical-minus-devDependencies bytes via shippedPluginManifest(), and TestMergePluginManifest_RestoresRepositoryManifest byte-compares mergePluginManifest(shipped, repo) against repo, which transitively forces the overrides blocks to match. One asymmetry to close opportunistically: the provenance test's packages loop cross-checks provenance versions against manifest pins at runtime, but its security-overrides loop compares only against inline expected values — adding the same provenance-to-manifest cross-check for overrides would mirror the repo's own packages-loop standard and catch coordinated half-updates that also edit inline test expectations.
