---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-05T14:19:54Z
identity: supply-chain-20261105T141954-jay-flowers
tier: draft
---

When a change removes a supply-chain control (a tracked lockfile with sha512 verification), audit every install path that relied on it, not just the primary one. PR 957 (stop-shipping-plugin-lock) applied the design's replacement mitigation (--ignore-scripts) to the narrow `uf init` activation graph (--omit=dev: zod, @opencode-ai/plugin) but omitted it on `make plugin-test`, the widest transitive surface (the vitest 5 dev tree), which runs in CI on every PR and on dev machines via `make check` — inverting the control: strictest on the smallest surface, none on the largest. Root pattern: when a mitigation is named in a design doc to offset a removed control, enumerate every command the change rewrites that loses that control, and require each rewritten command to carry the mitigation or a documented exception. Spec deltas should enumerate the rewritten install commands explicitly so an under-specified command in tasks.md cannot silently drop the mitigation.
