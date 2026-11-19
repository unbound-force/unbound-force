---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T06:02:30Z
identity: determinism-20261115T060230-jay-flowers
tier: draft
---

A "wording-only" instruction change still moves token counts and DE-004 still applies: tightening uf.address-feedback's question-tool phrasing to "Call the `question` tool — do not write the call as text" added +172 tokens (16566→16738) at zero line delta, on a file already over its 8000-token file budget and 20000-token load budget. Repeated guard clauses across 9 gate sites compound. When reviewing or authoring wording-only instruction changes, paste the instruction-lint delta.totals and per-file rows into the change record so "the repetition is cheap" claims rest on counts and the token-budget direction is visible.
