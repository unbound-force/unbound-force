---
tag: error-handling
author: jay-flowers
category: decision
created_at: 2026-11-07T17:13:30Z
identity: error-handling-20261107T171330-jay-flowers
tier: draft
---

When reviewer feedback asks to make an intentional silent fallback loud, locate where the risk is contained before accepting a new runtime signal channel. For preservePluginDevDependencies (internal/scaffold/plugin_activation.go), the anchor-miss passthrough can only trigger when this repo's own embedded manifest changes shape, and the round-trip test TestPreservePluginDevDependencies_RestoresRepositoryManifest (plugin_activation_test.go:1038, asserting preserve(shipped, repoManifest) == repoManifest byte-identical) fails loudly in CI the moment that happens — the silence is contained at the origin, and the two shipped guards (round-trip test plus an anchor-missing test case pinning the degraded behavior as intentional, plugin_activation_test.go:1081) are the pattern's documented risk acceptance. The loud alternatives each carry structural cost: an error return routes into the fail-fast per-file deploy loop and turns a cosmetic upstream reformat into a total uf init failure, a log line introduces the first side channel into a deliberately pure package (internal/scaffold ships zero production log calls), and a warnings field expands the public Result struct outside the change's scoped delta. The convention-compliant minimum (default pack CS-006) is a justifying comment at the decision point naming the guard tests. Reserve the loud-signal redesign for fallbacks that can trigger from inputs the origin's guard tests cannot see; one known residual is the vacuous-pass hole where running uf init after a format change rewrites the tracked manifest before the round-trip test can fail, leaving the tracked-file diff and make plugin-test as the remaining signals — a test-robustness matter, not a runtime-signal one.
