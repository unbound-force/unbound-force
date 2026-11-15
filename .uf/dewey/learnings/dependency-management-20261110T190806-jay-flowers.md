---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:08:06Z
identity: dependency-management-20261110T190806-jay-flowers
tier: draft
---

Adding an npm `overrides` entry does not move an already-installed node_modules tree onto the overridden version. Observed 2026-11-10 in unbound-force (quick/merge-plugin-runtime-deps-docs): the effect 4.0.0-beta.103 security override landed in .opencode/package.json and the scaffold mirror, the throwaway-lock OSV scan re-resolved the manifest clean, and make check stayed green — while .opencode/node_modules/effect still held the vulnerable 4.0.0-beta.83. No gate sees the gap: doctor's manifest-install compares only shipped direct pins (never `overrides`), the throwaway-lock OSV scan validates manifest resolution rather than the installed tree, and vitest asserts no transitive versions. After landing a security override, check the installed copy with `npm ls <pkg>` and rebuild node_modules when it is stale.
