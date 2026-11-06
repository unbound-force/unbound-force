---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-04T22:49:30Z
identity: spec-authoring-20261104T224930-jay-flowers
tier: draft
---

OpenSpec MODIFIED requirements replace the accumulated requirement text wholesale, and the accumulated text is the latest unarchived predecessor's version — not the original allocator's. When reviewing or authoring a delta that rewrites a requirement, diff the new text clause-by-clause against the most recent unarchived MODIFIED version of that requirement (grep the unarchived delta stack for the requirement ID, not just the change the amends note names). In stop-shipping-plugin-lock (PR #957), the rewritte
