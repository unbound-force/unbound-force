---
tag: scaffold
author: jay-flowers
category: pattern
created_at: 2026-11-06T20:02:54Z
identity: scaffold-20261106T200254-jay-flowers
tier: draft
---

When a tool-owned generated asset must preserve a user-declared block while staying byte-identical for skip/idempotency logic, byte-anchored text surgery beats a full JSON round-trip: re-serializing changes formatting and forces spurious rewrites, while locating a stable anchor (e.g. the `\n  "overrides": {\n` line) and splicing the preserved block before it keeps an unchanged target byte-identical and skipped. The cost is a silent no-op when the shipped format changes (anchor not found returns the shipped bytes, dropping the user block), so the pattern MUST ship with two guards: (a) a round-trip test asserting preserve(shipped, realRepoManifest) == realRepoManifest against the live repository manifest, which fails the moment the anchor or formatting drifts while the repo still declares the block, and (b) an explicit anchor-missing test pinning the degraded behavior as intentional. Implemented in unbound-force as preservePluginDevDependencies (internal/scaffold/plugin_activation.go) with its format-coupled inverse test helper shippedPluginManifest (internal/scaffold/scaffold_test.go): the production inserter and the test stripper must change together, and the mirror drift tests fail loudly if either drifts.
