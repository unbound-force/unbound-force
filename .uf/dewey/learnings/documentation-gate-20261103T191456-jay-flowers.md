---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-03T19:14:56Z
identity: documentation-gate-20261103T191456-jay-flowers
tier: draft
---

A CHANGELOG 'Added' entry is the project's own user-visibility signal: when a change records one, the AGENTS.md documentation-issue gate (file an issue against the current repo for user-facing changes before PR merge) applies by the same classification. A PR body that defers the decision — "file one if the docs gate requires it" — leaves the gate undecided at review time, so the reviewer must either see the filed issue reference or a recorded exemption rationale (e.g., report-format-only change with no stale documentation to update).
