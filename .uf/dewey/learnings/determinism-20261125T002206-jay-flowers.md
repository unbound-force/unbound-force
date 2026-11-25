---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-25T00:22:06Z
identity: determinism-20261125T002206-jay-flowers
tier: draft
as_of: "unbound-force@c9241969995f590083f3f56f0a0347168c4d56f9"
---

tool-inventory --seams classifies test layers by filename and it()/describe() title patterns (a test must pair the tool name with "round-trip" phrasing to count as the round-trip layer). A DE-007 seam test whose title describes the round-trip in other words — "carry a resolved state from write to persisted artifact" — classifies as layer "other" and the tool reports the layer as missing. The Determinist must read the referenced test files before reporting DE-007 on a missing layer: the missing list is evidence, not a verdict. Verified in unbound-force intent-satisfaction-check: acquire_intent_context, dispatch_review_runs, and finalize_review_dispatch all reported missing: ["round-trip"] while .opencode/test/intent-satisfaction.test.ts:767-825 ("Dispatch contract seams (DE-007)") verifies registration surface, input-schema parity, documented output shape, and the write→read round-trip through session-metadata.json, reviewer prompts, the persisted artifact, and the consolidation result.
