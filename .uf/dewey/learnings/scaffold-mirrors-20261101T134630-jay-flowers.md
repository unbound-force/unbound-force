---
tag: scaffold-mirrors
author: jay-flowers
category: pattern
created_at: 2026-11-01T13:46:30Z
identity: scaffold-mirrors-20261101T134630-jay-flowers
tier: draft
---

Complete scaffold-registration checklist for new shared .opencode/lib modules (demonstrated by the model-matrix-delegation change, PR #894): when a new lib module is consumed by a scaffolded plugin, the full registration surface is eight pieces — (1) isToolOwned entry in internal/scaffold/scaffold.go, (2) stageReviewPluginSources entry in internal/scaffold/plugin_activation.go, (3) expectedAssetPaths + task42CanonicalAssets entries in internal/scaffold/scaffold_test.go, (4) the file-count range arithmetic in cmd/unbound-force/main_test.go (base = total embedded − non-deployed − activation-gated), (5) knownNonEmbeddedFiles entries for deliberately repo-local companions (e.g. hidden classifier agent, governance packs) with the deferral/precedent documented in the entry comment, (6) isDivisorAsset exclusion tests when the asset is not part of review dispatch, (7) the gitignore managed block for per-developer override files, and (8) schema-registry triple copy (root schemas/, .uf/schemas/, scaffold assets) kept in sync by a cross-validation harness with pinned per-fixture verdicts. The scaffold_test.go registries fail closed, so a missed piece surfaces as a test failure rather than silent drift — but knowing the checklist up front avoids a discovery-by-failure cycle. This extends the earlier scaffold-mirrors learning (new lib modules fall outside the fixed asset-surface enumeration unless explicitly added) with the complete registration surface including the non-embedded-companion and Divisor-exclusion dimensions.
