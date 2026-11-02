---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:00:38Z
identity: issue-triage-20261102T010038-jay-flowers
tier: draft
---

When triaging spike or verification issues that enumerate expected counts as acceptance criteria (e.g., "16 agents load, 32 commands work"), reconcile the counts against both the live tree and AGENTS.md before accepting them — the three sources can silently disagree. Triaging unbound-force#798 (OpenCode v2 Phase 0 spike, umbrella #797) found: the issue expects 16 agents / 32 commands; the tree holds 17 agent files and 34 command files; AGENTS.md documents "17 active, model-matrix-classifier hidden" and "Slash commands (33 files)". The agent figure reconciles exactly as the visible-loadable count (17 minus the 1 hidden classifier = 16); the command figure drifts by 2 against the tree and 1 against AGENTS.md with no stated reconciliation. For a spike this is self-correcting — the empirical run recounts what loads — but the same drift in an implementation issue would bake a wrong acceptance criterion into tasks.md. Technique: glob the target directory, subtract documented hidden/disabled entries, and flag any residual delta in the triage output so the follow-up spec inherits the reconciled number.
