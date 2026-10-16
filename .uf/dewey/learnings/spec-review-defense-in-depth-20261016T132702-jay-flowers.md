---
tag: spec-review-defense-in-depth
author: jay-flowers
category: pattern
created_at: 2026-10-16T13:27:02Z
identity: spec-review-defense-in-depth-20261016T132702-jay-flowers
tier: draft
---

When a spec introduces a new validation layer (e.g., zod schema validation) in front of existing runtime validation (e.g., INVALID_PR_NUMBER guards), the acceptance criteria should explicitly document which layer is responsible for rejecting which class of invalid input. Even when the outer layer is SDK-owned and not under the spec's control, a brief note clarifying the boundary prevents future confusion about defense-in-depth responsibilities. The fix-tool-args-schema-mismatch spec handles this well in design.md D4 but could be more explicit in the FR-003 acceptance criteria.
