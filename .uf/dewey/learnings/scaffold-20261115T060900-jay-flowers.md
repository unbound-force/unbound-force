---
tag: scaffold
author: jay-flowers
category: gotcha
created_at: 2026-11-15T06:09:00Z
identity: scaffold-20261115T060900-jay-flowers
tier: draft
as_of: "unbound-force@e5e077b35f0c7984517024431e696737e330b080"
---

TestEmbeddedAssets_MatchSource in internal/scaffold/scaffold_test.go enumerates the embedded-assets side only: a NEW canonical file under .opencode/ with no scaffold mirror passes the drift test silently — nothing walks the canonical side. The real safety nets for a new canonical lib file are TestStageReviewPluginSources_CoversImportedLibModules (the staging list in plugin_activation.go must name the file), TestAssetPaths_MatchExpected (expectedInitiallyDeployedAssetPaths), and the derived count tests (TestRun_CreatesFiles / SkipsExisting / ForceOverwrites read their counts from that same list, so they self-correct). When adding a file under .opencode/lib/, wire every registration point at once: the staging list, isToolOwned and isDivisorAsset in scaffold.go (NOT isActivationGatedAsset — it gates only the 4 plugin index files), the mirror copy under internal/scaffold/assets/opencode/lib/, the expected-assets "Shared lib modules" block, task42CanonicalAssets, and the isToolOwned/isDivisorAsset test-table rows. Verified on opsx/harden-spec-review-gate (2026-11-15): progress-state.ts initially passed the drift test with no mirror; the staging test caught it.
