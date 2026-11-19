---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T02:15:09Z
identity: determinism-20261115T021509-jay-flowers
tier: draft
as_of: "unbound-force@7f44cba6be6f30f6c3cec406daab342434cc1474"
---

A deterministic relay tool does not close a gate bypass if invoking the tool stays conditioned on prose-evaluated state. When the tool's contract already computes the routing decision (e.g. gate_paused derived from progress state + dispatch artifact), the commanding prose must call the tool unconditionally and branch only on the tool's structured output; a prose "if <state condition> then call the tool" gate re-opens the exact compaction/model-deviation bypass the tool was built to close, because a mangled orchestrator that mis-evaluates the condition never calls the tool at all — the artifact-sourced rendering protects only what the tool renders, not whether it is invoked. Diagnostic signal: a tool output branch (gate_paused: false) that is unreachable through the documented prose gate except by mis-evaluation means the routing is duplicated between prose and tool and the prose copy is the live bypass surface. Observed reviewing iteration 2 of OpenSpec change harden-spec-review-gate (task 5.2 / design Decision 4 / HARDEN-004 vs task 5.1's gate_paused contract).
