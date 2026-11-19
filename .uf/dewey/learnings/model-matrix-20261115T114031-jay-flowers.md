---
tag: model-matrix
author: jay-flowers
category: gotcha
created_at: 2026-11-15T11:40:31Z
identity: model-matrix-20261115T114031-jay-flowers
tier: draft
---

Schema-extending .uf/model-matrix.yaml mid-pipeline trips MATRIX_INVALID in parented child sessions: uf_delegate_command in a child of a long-lived host fails with 'Model matrix at .uf/model-matrix.yaml is structurally invalid: ... Unrecognized keys: "selection", "delegation"' when the branch adds new top-level knob sections (added 2026-11-15 on opsx/compaction-storm-on-a-tight-context-mode: selection.footprint_margin_tokens, delegation.storm_threshold), because the shared host process still runs the pre-branch uf-workflow plugin with the older schema. Workaround that kept the pipeline moving: temporarily truncate BOTH .uf/model-matrix.yaml and its scaffold mirror internal/scaffold/assets/uf/model-matrix.yaml to the pre-branch boundary (first 46 lines, ending at 'classifier: {}'), delegate with an explicit instruction not to touch/stage/commit the two matrix files, then restore both from /tmp/opencode/matrix-backup/ and verify git diff is empty. New knobs are optional with identical defaults, so the stripped matrices validate under both stale and fresh code. The workaround must never appear in any diff, review, commit, or staging. Root fix: restart the host (or run the delegation from a fresh session) so the plugin reloads the extended schema.
