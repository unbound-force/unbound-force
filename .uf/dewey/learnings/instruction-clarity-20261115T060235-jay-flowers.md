---
tag: instruction-clarity
author: jay-flowers
category: pattern
created_at: 2026-11-15T06:02:35Z
identity: instruction-clarity-20261115T060235-jay-flowers
tier: draft
---

When tightening tool-invocation prose ("Present the X tool" → "Call the X tool"), append negative guards ("do not write the call as text") after the complete positive instruction rather than interpolating them between the verb and its arguments, and re-anchor carried-over positional phrases ("below the printed text") to the new verb. In unbound-force PR #1092 (uf.address-feedback), inserting a ~20-word em-dash guard between "Call the `question` tool" and "with the same options, the recommended option FIRST" at six gates left both the vestigial phrase "below the printed text" and the options complement ambiguously attached — the guard fixed the tool-call-as-text stall but introduced a new parse ambiguity at the same sites.
