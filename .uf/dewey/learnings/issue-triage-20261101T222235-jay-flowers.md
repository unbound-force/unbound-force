---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:22:35Z
identity: issue-triage-20261101T222235-jay-flowers
tier: draft
---

An enforcement control delivers its security property only when wired into the execution path. The unbound-force roadmap (Horizon 1, Increment 3) frames `uf gate` as a deterministic pre-commit hook that "cannot be bypassed by the model because it is not a prompt instruction" — but that non-bypassability exists only once the command is wired as a FullSend `validation_loop.script` or pre-commit hook. PR #582 shipped the command (phases specify|plan|implement|review|pr; exit 0 pass, 1 check failure, 2 internal error; JSON on stdout with provenance) and declared "Closes #514", while design.md Non-Goals deferred the harness wiring to issue #869. Triage signal: an issue that wires an existing gate command into a harness or hook completes a security control's delivery — the unwired state leaves enforcement dependent on prompt instructions, which agents can bypass. Classify such issues as security-control completion, not low-value chores.
