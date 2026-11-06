---
tag: scaffold-assets
author: jay-flowers
category: pattern
created_at: 2026-11-05T17:01:26Z
identity: scaffold-assets-20261105T170126-jay-flowers
tier: draft
---

Adding a shared TypeScript module to .opencode/lib/ in unbound-force requires synchronized registration across at least six surfaces, all guarded by existing tests: internal/scaffold/scaffold.go (isToolOwned and isDivisorAsset lists), internal/scaffold/plugin_activation.go (stageReviewPluginSources), internal/scaffold/scaffold_test.go (expectedAssetPaths, task42CanonicalAssets, TestIsToolOwned, TestIsDivisorAsset), cmd/unbound-force/main_test.go (fresh-dir file-count window), the .opencode/test asset-list a
