---
tag: regression-protection
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:53:42Z
identity: regression-protection-20261105T195342-jay-flowers
tier: draft
---

A "keep the two in sync" invariant between duplicated constants in build files (e.g. Makefile PLUGIN_TEST_NPM = npx -y npm@11.6.2 and the osv_plugin_transitive job in .github/workflows/ci_security.yml) enforced only by reciprocal comments has no drift detection — grep confirmed no test references either location. The unbound-force repo already ships the mechanism for this class of invariant: canonical↔mirror drift tests (TestEmbeddedAssetsMatchSource, task42CanonicalAssets, and the .opencode/.gitignore↔opencodeGitignore test). Reuse the pattern: one test parses both files and asserts the pinned values match.
