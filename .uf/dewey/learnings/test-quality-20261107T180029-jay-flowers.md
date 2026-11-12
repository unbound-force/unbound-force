---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-07T18:00:29Z
identity: test-quality-20261107T180029-jay-flowers
tier: draft
---

When a shipped asset is a transformation of a canonical source (e.g., the scaffolded .opencode/package.json is the repository manifest minus its devDependencies block), pin the regression lock at two layers: (1) a drift test that derives the expected bytes independently of production code (shippedPluginManifest line-stripping in scaffold_test.go) plus a direct negative assertion (the mirror must not contain devDependencies), and (2) pin the transformation's preconditions — TestSCFR003_RunPreservesTargetDevDependencies fatals when the shipped manifest lacks the overrides anchor, so preservePluginDevDependencies's silent return-shipped fallback cannot fire unnoticed. Derived-asset drift tests that only compare against a static copy miss both silent transformation changes and silent fallback paths.
