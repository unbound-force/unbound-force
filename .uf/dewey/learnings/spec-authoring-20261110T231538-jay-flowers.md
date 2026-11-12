---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-10T23:15:38Z
identity: spec-authoring-20261110T231538-jay-flowers
tier: draft
---

When an implementation must introduce an auxiliary cap to satisfy a spec'd byte budget — e.g., capping a compact finalization result's errors array (5 strings of 256 characters) so the result stays under 40,960 bytes once full-length proposals become the only droppable array — the spec delta's unchanged-caps enumeration must name every array the change touches. Silence reads as "no cap": the archived spec then contradicts the shipped tool description and CHANGELOG, and a reader implementing from the spec alone either violates the budget MUST or reinvents the cap undocumented. The tell: the enumeration lists findings, advisories, warnings, and verdict_reason but omits exactly the field the diff newly caps. Audit pattern: diff the spec delta's cap enumeration against every constant the implementation adds or removes from the limits table — any mismatch is a spec-accuracy finding even when the mechanism itself is necessary and correct.
