---
tag: plugin-scaffold
author: jay-flowers
category: pattern
created_at: 2026-10-07T18:58:50Z
identity: plugin-scaffold-20261007T185850-jay-flowers
tier: draft
---

When adding a new plugin to the unbound-force scaffold system, the activation gating architecture requires updates in six distinct locations: (1) scaffold.go constant definition, (2) isActivationGatedAsset() predicate, (3) plugin JSON registration loop and hasPlugins check, (4) isToolOwned() and isDivisorAsset() ownership functions, (5) plugin_activation.go staging and activation lists, and (6) comprehensive test updates across scaffold_test.go and plugin_activation_test.go. The TestAtomicallyActivateReviewPlugins_Branches test has three subtests (Force refresh, Content diff refresh, Identical skip) that each need staged directory creation, installed directory creation, file writes, and post-activation content assertions. Missing any of these locations causes test failures that surface as SCFR003 violations. The scaffold_test.go also requires expectedAssetPaths and task42CanonicalAssets updates for drift detection. The main_test.go TestRunInit_FreshDir file count range must be bumped by the number of new files added.
