---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-15T15:53:06Z
identity: instruction-quality-20261115T155306-jay-flowers
tier: draft
---

When command prose names a structured tool argument, state the argument's type and derivation. A placeholder like `N` reads as a JSON number; when the tool's schema wants a string (e.g. harden-spec-review-gate's spec_review_decisions finding ids are the 1-based index into the artifact's findings array, rendered as "1".."N" by relay_spec_review_gate), prose-following agents fail zod input validation on the first call and must recover from the tool error. Name the id scheme and the tool output where the agent can see it.
