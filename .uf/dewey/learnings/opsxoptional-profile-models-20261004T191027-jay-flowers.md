---
tag: opsxoptional-profile-models
author: jay-flowers
created_at: 2026-10-04T19:10:27Z
identity: opsxoptional-profile-models-20261004T191027-jay-flowers
tier: draft
---

Optional Zod schema fields (`ModelSchema.optional()`) combined with nullable TypeScript types (`string | null`) requires careful handling of the boundary between YAML/JSON null and undefined. In the `mergeOverride` function, the override YAML may contain `model: null` explicitly. When merging, null must be converted to undefined because `ModelSchema.optional()` allows undefined (absent key) but rejects null at the Zod validation level. The fix was `profileObj.model ?? undefined`. This is a pattern to apply whenever optional Zod schemas interact with YAML/JSON input that may contain explicit null values.
