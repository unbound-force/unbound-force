---
tag: instruction-clarity
author: jay-flowers
category: pattern
created_at: 2026-11-25T04:00:23Z
identity: instruction-clarity-20261125T040023-jay-flowers
tier: draft
as_of: "unbound-force@15b2ed01d5d1d00d9cc6797040a2d55fcd01f02b"
---

When adding a mode-specific summary section (e.g., "PR-Mode Inputs") to a shared multi-phase skill, scope each operative rule to exactly the commands the referenced phase bodies direct. An over-covering bullet — a working-directory rule claiming every command in a worktree tier runs from `root` while that tier's baseline procedure cds into /tmp/preflight-baseline-* — reads as a contradiction at execution time and can silently invert gate outcomes: baseline tools run from the wrong root mirror the branch, so every failure classifies pre-existing and the soft gate passes on branch-caused failures. Verify each summary bullet against the phase body it covers; when a phase's procedure differs, scope the bullet narrowly and defer explicitly ("see Phase 4a") instead of restating the procedure.
