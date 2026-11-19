---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T09:19:30Z
identity: determinism-20261116T091930-jay-flowers
tier: draft
---

Submitted by divisor-determinism during the opsx/per-task-delegation-budgets code review iteration 1 (correlation bd99036f-177a-4bb0-8743-ede5f63db5ce); stranded because dewey_store_learning was unavailable; written to disk at retrospective. When command prose branches on a delegation failure envelope, branch on the classification fields the tool already computes (e.g. `error.retryable` on every DELEGATE_FAILED envelope from uf_delegate_command — the executor sets it true for wall_clock/idle/storm and false for budget_exhausted/cancelled) instead of re-deriving a kind→action mapping from `timeout.kind` values in prose. Observed in the per-task-delegation-budgets review: the new uf.unleash.md retry branch enumerated "kind wall_clock or idle (retryable), kind budget (non-retryable)" while the envelope's `retryable` field already rendered that routing — and the two tables diverged on the storm kind (envelope retryable: true; the prose enumeration stops on it), creating a second source of truth for the same decision. The same discipline applies to new prescribed output lines: the repo's mechanism of record for any fixed-format terminal line is a tool-rendered field the command emits verbatim (format_unleash_exit/format_unleash_demo {text}, uf_preflight_declared_checks `rendered`, post_feedback_replies summary line), not a prose template the agent reconstructs — even for a single status line, and even when a Go content gate pins the prose, because pinning prose entrenches the prose as the mechanism.
