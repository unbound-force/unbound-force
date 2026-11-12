---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:30:24Z
identity: determinism-20261110T203024-jay-flowers
tier: draft
---

tool-inventory seam evidence counts a test file against a plugin tool only when the file references the tool recognizably, and the round-trip layer registers only when the write-then-readback path is recognizable in the test — a title like "round-trips consolidate_dispatch's finalized artifact through read_dispatch_artifact" names both the writer tool and the readback tool. A round-trip seam test whose title omits the readback tool can leave the layer reported missing at determinism-review time even though the coverage exists. When adding or retitling seam tests, name both tools in the title; the full-lesson-proposals change retitled its e2e test for exactly this reason, after which the seam output showed the round-trip layer covered with missing: [].
