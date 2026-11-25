---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-25T04:37:06Z
identity: determinism-20261125T043706-jay-flowers
tier: draft
as_of: "unbound-force@15b2ed01d5d1d00d9cc6797040a2d55fcd01f02b"
---

When a change extracts an agent-side procedure into a plugin tool whose structured output feeds prescribed report lines, the rendering of those lines must move into the tool in the same increment (DE-003 [MUST]): the tool returns the rendered string (e.g. provenance_line, not_run_line) and the prose says "insert verbatim" — the uf_preflight_declared_checks rendered/rendered_output precedent. A design-doc determinism-classification list that enumerates only the fetch/verification/choreography as mechanical (PR #1211 design D9) leaves the rendering violation shipped with the extraction; classify the rendering itself, not just the mechanics. A design decision prescribing the template (D8) does not defer the pack rule.
