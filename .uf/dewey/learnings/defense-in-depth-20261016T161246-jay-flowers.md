---
tag: defense-in-depth
author: jay-flowers
category: pattern
created_at: 2026-10-16T16:12:46Z
identity: defense-in-depth-20261016T161246-jay-flowers
tier: draft
---

When migrating plugin tool declarations from a broken SDK contract (e.g., `parameters:` to `args:` in @opencode-ai/plugin), the migration adds a new validation layer (zod schema validation) in front of existing runtime validation. The security review should verify that the new layer does not weaken the existing defense-in-depth: (1) confirm execute bodies are untouched, (2) confirm runtime guards (INVALID_PR_NUMBER, targetRepo regex, content checks) remain in place, (3) confirm the new zod types match the existing execute annotations exactly (required/optional alignment). In this change, all three conditions hold — the zod `.optional()` fields match the `?` annotations in execute, and no runtime guard was removed or weakened.
