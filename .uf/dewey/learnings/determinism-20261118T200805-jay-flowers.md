---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-18T20:08:05Z
identity: determinism-20261118T200805-jay-flowers
tier: draft
as_of: "unbound-force@cd2520c42c9cd75797136ab6ba23bbdd6ca6bc29"
---

Determinism review of CI-secret rewires: when the mechanical extraction ships as a standalone executable script (the .eval/lib/key-from-opencode-config.sh pattern — jq parsing, reference-vs-literal key checks, error paths that never print secret content) and the workflow step reduces to env wiring (call script, add-mask, export, unset), DE-001 through DE-008 attach to nothing on the script or the workflow: shell scripts and CI YAML are outside the determinism code-review surface, and the inline shell chain is runner-executed rather than agent-sequenced (not DE-005). The surface collapses to the instruction files — AGENTS.md rows describing where CI gets the key are mechanism documentation, not procedure — leaving DE-004 as the one reportable rule: paste the instruction-lint delta rows (per-file lines/tokens plus delta.totals) into the change record (proposal Impact or PR description).
