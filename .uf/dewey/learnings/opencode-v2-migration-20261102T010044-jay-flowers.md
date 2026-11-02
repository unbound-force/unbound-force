---
tag: opencode-v2-migration
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:00:44Z
identity: opencode-v2-migration-20261102T010044-jay-flowers
tier: draft
---

Triaging unbound-force#798 (OpenCode v2 Phase 0 spike) against the repo verified the issue's named Bun risk but revealed it is under-scoped: beyond createBunYamlParser (.opencode/lib/reviewer-manifest.ts:166, consumed by review-dispatch, uf-workflow, and invoke-agent), production code in .opencode/plugins/review-dispatch/index.ts invokes Bun.spawn directly for every gh-CLI subprocess call (runGhCommand and its siblings, ~line 7614 onward — 45 Bun.* matches suite-wide). If the OpenCode v2 plugin runtime is node rather than bun, both Bun.YAML and Bun.spawn break, so a v2 runtime audit keyed only to "replace YAML parsing" would miss the subprocess surface. The "confirm the v2 plugin runtime (bun vs node)" check should enumerate all Bun.* call sites, not just the parser. Mitigations already in the tree to extend: the Node-side vitest suite ships Bun shims (.opencode/test/helpers/bun-runtime.ts, gh-spawn-mock.ts, yaml-subset.ts) and createBunYamlParser fails closed with an explicit "Bun.YAML.parse is unavailable; reviewer manifest parsing requires Bun 1.3.x or newer" error — a Node-compat seam pattern that already exists.
