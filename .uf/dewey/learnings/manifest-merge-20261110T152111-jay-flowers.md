---
tag: manifest-merge
author: jay-flowers
category: pattern
created_at: 2026-11-10T15:21:11Z
identity: manifest-merge-20261110T152111-jay-flowers
tier: draft
---

Byte-stable merges of tool-owned JSON manifests (uf init's .opencode/package.json) work as anchored splices, not full re-serialization: parse both manifests to compute the merged map, then replace only the target block's body between a structural anchor (e.g. "\n  \"dependencies\": {\n") and its closing line, rendering entries sorted through a no-HTML-escape JSON string encoder (jsonString). This keeps an unchanged target byte-identical so the tool-owned skip fires, where json.Marshal would reformat the whole file and defeat the skip. Contain the anchor fragility with three test layers: a repository-manifest round-trip (merge(shipped, repo) == repo), anchor-miss passthrough cases, and an idempotence test over the merge's own output. Precedent: preservePluginDevDependencies (stop-shipping-plugin-lock D2a) and preservePluginRuntimeDependencies (merge-plugin-runtime-deps). Gotcha when adding the second splice beside the first: the containment comment adapted from the sibling function can retain the sibling's test-file reference — verify cited test files actually contain the new function's tests.
