---
tag: ci-causality-soft-gate
author: yvonne-devlin
category: pattern
created_at: 2026-07-01T10:06:19Z
identity: ci-causality-soft-gate-20260701T100619-yvonne-devlin
tier: draft
---

When adding a new execution mode to a shared skill with multiple consumers, the cleanest approach is a new mode rather than extending an existing mode with optional parameters. The soft-gate addition to the pre-flight skill demonstrated this: adding a third mode alongside hard-gate and ci-aware preserved the principle that each mode has one unambiguous behavior. Extending hard-gate with optional baseline parameters would have made it behave differently depending on context, breaking the clean contract that existing consumers (unleash) depend on. The key design insight was that consumer inheritance works for free: /unleash's code review step delegates to /review-council, which now uses soft-gate, so /unleash inherits the improvement without any direct changes. Only /unleash's phase checkpoints remain hard-gate, which is the correct behavior for mid-implementation stops.
