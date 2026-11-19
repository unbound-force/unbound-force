---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-10T21:26:04Z
identity: spec-authoring-20261110T212604-jay-flowers
tier: draft
---

When a change re-bounds one field of a tool result, audit the spec delta's "unchanged" enumeration against every other field of that result the code touches. The full-lesson-proposals delta enumerated findings, advisories, warnings, and verdict_reason as unchanged but silently omitted the errors array — which the same change newly capped (5 strings of 256 characters, errors_truncated flag) as an enabling detail for the byte budget. The cap shipped, was tested, and was documented in the CHANGELOG and tool description, but the spec of record (once archived) would misstate the contract, and the consuming commands' prose (uf.review-council §5, uf.review-pr §5, which instruct parents to read errors directly from the compact result) was never told about the cap. Rule of thumb: enumerate the full field list of the changed result when writing an "unchanged caps" clause, and diff that list against the fields the implementation actually touches.
