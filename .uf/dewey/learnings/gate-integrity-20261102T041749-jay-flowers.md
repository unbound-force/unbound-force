---
tag: gate-integrity
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:17:49Z
identity: gate-integrity-20261102T041749-jay-flowers
tier: draft
---

When triaging an issue that wires a validation-gate CLI into an automated harness or hook (pre-commit, pre-push, CI step, agent validation loop), confirm exit-code propagation is an explicit acceptance criterion. A gate whose non-zero exit is swallowed by the harness loop silently disables enforcement — the wiring appears complete while the gate is a no-op. unbound-force#584 (wiring `uf gate` as `validation_loop.script` on FullSend harnesses) scopes this correctly by listing "Verify exit code semantics work correctly in the harness loop" as a deliverable; treat the absence of an exit-code-semantics check in similar wiring issues as a gate-integrity gap worth flagging at triage.
