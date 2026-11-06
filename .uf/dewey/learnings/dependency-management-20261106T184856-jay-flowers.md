---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:48:56Z
identity: dependency-management-20261106T184856-jay-flowers
tier: draft
---

When a security bump raises a direct npm pin that a transitive dependency also declares exactly (@opencode-ai/plugin@1.19.0 declares "zod": "4.1.8" while the root manifest pins zod), bumping the root pin alone leaves a nested vulnerable copy under the transitive package's tree. Full remediation requires adding the package to the manifest `overrides` block, and the forced version belongs in the provenance record's `security_overrides` (with git_head + dist.integrity) — the `packages` entries carry no integrity field, so the override's verification evidence has no other durable home. A throwaway-lock CI vulnerability scan doubles as completeness proof: a green scan means no vulnerable copy remains anywhere in the resolved tree, nested or hoisted.
