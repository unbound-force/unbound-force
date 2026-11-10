---
tag: scaffold-file-ownership
author: jay-flowers
category: pattern
created_at: 2026-11-10T15:21:21Z
identity: scaffold-file-ownership-20261110T152121-jay-flowers
tier: draft
---

Tool-owned scaffold files that must preserve user entries merge best with the pattern proven in preservePluginDevDependencies and now preservePluginRuntimeDependencies: parse the target, keep only the keys the shipped asset does not declare, return the shipped bytes unchanged when the target adds nothing (so pin-only targets stay byte-identical and report as skipped, keeping drift-detection tests and updated/skipped reporting stable), otherwise splice a sorted, jsonString-rendered block at the shipped asset's fixed anchor. Pin the contract with three tests: idempotence (merge over its own output is a no-op), the repository fixed point (merge(shipped, repo manifest) == repo manifest byte-for-byte), and anchor-miss passthrough. Unparsable targets pass through to the shipped bytes so uf init repairs rather than fails — internal/scaffold deliberately has no warnings channel, and the silence is contained at origin by those tests.
