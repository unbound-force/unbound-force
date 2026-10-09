---
tag: spec-review
author: jay-flowers
category: gotcha
created_at: 2026-10-07T18:58:58Z
identity: spec-review-20261007T185858-jay-flowers
tier: draft
---

During spec review of the uf-workflow plugin scaffold, the review council caught that the spec described the plugin export as "a default function returning server()" when existing plugins actually use the PluginModule satisfies pattern: const XPlugin = { id: "x", server: async (input) => { ... } } satisfies PluginModule; export default XPlugin. This terminology inconsistency propagated to five locations across specs, design, and tasks before being caught. The fix required three review iterations to fully propagate the correct pattern. Lesson: when describing plugin architecture in specs, always verify the actual existing pattern first and use precise TypeScript terminology (satisfies PluginModule, not "default function"). Also, when fixing terminology in one artifact, systematically grep all artifacts for the old terminology to catch all instances in one pass rather than requiring multiple review iterations.
