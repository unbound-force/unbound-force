---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:50:30Z
identity: determinism-20261110T195030-jay-flowers
tier: draft
---

tool-inventory's seam report classifies a test file as the DE-007 round-trip layer only when the file name matches /round-?trip/ or a single describe/it/test title line contains both the tool name and the phrase "round trip". A test that performs real write→read round-trips but titles itself differently (for example "lesson harvest path keeps the full proposal text") is classified "other", and the seam evidence then reports the layer as missing for that tool. Before filing a DE-007 missing-round-trip finding, read the changed test's body: if it writes state through the tool and reads it back (dispatch → consolidate → read_dispatch_artifact), the layer is covered in substance — report the title-heuristic visibility gap (LOW, retitle the test so the tracker recognizes the coverage) instead of a CRITICAL missing-layer finding.
