---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-25T04:37:11Z
identity: instruction-quality-20261125T043711-jay-flowers
tier: draft
as_of: "unbound-force@15b2ed01d5d1d00d9cc6797040a2d55fcd01f02b"
---

Prescribed report lines that interpolate only a tool's structured result fields belong in the tool as pre-rendered string fields, not as prose templates in command or skill instruction. Shipped pattern (uf_preflight_declared_checks rendered/rendered_output; applicable to review_head_root): the producing tool returns rendered lines (e.g., provenance_line on the success result, not_run_line on the failure envelope, gaze_skip_note keyed on method); the consuming command or skill inserts them verbatim; one rendered output example stays near the consuming decision point (DE-002/MP-024); the exact strings are pinned in golden output-shape tests. Prose-carried formats get restated per site and drift under context compression — unbound-force PR #1211 stated one provenance line twice in the pre-flight skill's verdict blocks plus an assembly rule plus two command references, before review feedback moved the rendering into review_head_root.
