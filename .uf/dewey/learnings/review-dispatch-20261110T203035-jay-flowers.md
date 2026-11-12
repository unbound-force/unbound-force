---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:30:35Z
identity: review-dispatch-20261110T203035-jay-flowers
tier: draft
---

When a tool-result byte budget is enforced on the UTF-8 serialization but the input schema caps string length in UTF-16 code units, a single schema-maximum entry can exceed the budget on its own: 16,384 CJK characters satisfy z.string().max(16384) yet serialize to 49,152 UTF-8 bytes — over a 40,960-byte budget. Budget-drop loops need a test at the all-entries-dropped boundary (kept.length === 0, entries key absent, truncation flag set, result under budget), not only the some-dropped boundary; an ASCII-only over-budget fixture always keeps at least one entry and stays green if the loop is later changed to always keep one.
