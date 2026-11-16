---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T09:52:00Z
identity: determinism-20261116T095200-jay-flowers
tier: draft
---

Submitted by divisor-determinism during the opsx/per-task-delegation-budgets code review iteration 2 (correlation ea364acd-807a-466e-8b22-856526787b52); stranded because dewey_store_learning was unavailable; written to disk at retrospective. Resolving DE-006 for retry routing (shipped in per-task-delegation-budgets, issue #1080): when a command must branch on delegation failure semantics, have the tool compute and expose the routing decision as a validated envelope field instead of instructing the agent to enumerate the tool's internal classification. Precedent: uf_delegate_command's failure envelope carries error.retryable, computed inside agent-executor per termination class (idle/wall-clock storm-free terminations retryable, budget_exhausted non-retryable); /uf.unleash branches on that single field plus the presence of the timeout envelope rather than mapping timeout.kind values to retryability in prose. The tool owns the routing decision; the agent consumes the field. The same shape answers any future "command pre-computes the tool's internal routing" finding: move the mapping into the tool, expose the result, and branch on the exposed field.
