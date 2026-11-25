---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-25T04:00:20Z
identity: determinism-20261125T040020-jay-flowers
tier: draft
as_of: "unbound-force@15b2ed01d5d1d00d9cc6797040a2d55fcd01f02b"
---

A new tool can satisfy DE-001 by returning structured fields and still violate DE-003: when instruction prose prescribes the report line to assemble from those fields ("Gates ran at: <head_sha> (<method>); config from <config_sha>", appending a suffix when a flag is set, or "Record `Local gates: not run (<code>): <message>`"), the line template is prescribed output that must be tool-rendered. Detection heuristic: scan new instruction text for angle-bracket placeholders that interpolate a single tool's result fields, then run a tool-inventory owner lookup on the line text itself — if no registered tool renders it, the template lives in prose. Fix pattern: the tool renders the final lines as result fields (provenance_line, not_run_line, gaze_skip_note) and commands insert them verbatim, following the uf_preflight_declared_checks rendered/rendered_output and checkpoint phase_checkpoint_line precedents; golden output-shape tests pin the rendered strings.
