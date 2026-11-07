---
tag: error-handling
author: jay-flowers
category: pattern
created_at: 2026-11-07T17:13:27Z
identity: error-handling-20261107T171327-jay-flowers
tier: draft
---

Byte-anchored preservation splices need both a failure channel and an anchor drift test. When a merge function like preservePluginDevDependencies (re-inserting a target's devDependencies before the manifest's overrides anchor) returns only the merged bytes, anchor misses and unparseable inputs become observationally identical to success — the silent-failure class, and the one input path where a spec preservation MUST goes unmet with zero signal. Two guards close it: (1) return a preservation-status signal (bool or error) so the call site can log or fail loudly, naming the file and reason; and (2) bind the anchor to the real embedded asset in a drift test (the TestPreservePluginDevDependencies_RestoresRepositoryManifest pattern: run the merge over the tool's own repository manifest and assert a byte-identical round-trip), so a shape change to the asset fails CI at authoring time instead of silently disabling preservation in every downstream target.
