---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-16T11:57:24Z
identity: instruction-quality-20261116T115724-jay-flowers
tier: draft
as_of: "unbound-force@474f5aee31d7e0bc4583635cccfeec6030277b33"
---

When a spec mandates tool-side derivation of a value at consumption time (PR #1157, PTB-FR-004: remaining budget "MUST be derived from resolveInheritedDeadline on the current session"), threading it as an agent-supplied command-template parameter with a prose selection rule over prior delegation envelopes fails twice: the envelope value is a result-time snapshot (stale at checkpoint time by whatever the pre-flight gate and fix loops consume), and the selection inputs exist only in conversation context — exactly what the consuming command's own session-resume guard forbids relying on after compression. The spec's registry-based mandate is precisely the compression-surviving derivation; the prose rule is precisely the compression-fragile one. Pattern: derive the value inside the consuming tool from the registry the spec names, and keep command templates free of envelope-selection prose. Observed at .opencode/commands/uf.unleash.md ~1501 (fill rule) vs the openspec per-task-delegation-budgets delta; the checkpoint tool's own parameter comment carried the same tension ("at checkpoint time, taken from the delegation envelope").
