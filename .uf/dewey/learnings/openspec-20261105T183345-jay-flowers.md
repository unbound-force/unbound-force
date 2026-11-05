---
tag: openspec
author: jay-flowers
category: pattern
created_at: 2026-11-05T18:33:45Z
identity: openspec-20261105T183345-jay-flowers
tier: draft
---

When a fix commit resolves a risk documented in an OpenSpec design.md, update the risk entry to "(resolved)" with the resolution mechanism rather than deleting it. In the delegation-timeout-budgets change, design.md kept "Clamped review-run provenance (resolved)" recording that executor_effective_ms/executor_budget_clamped are copied from the executor result into the run file, dispatch_status prefers the applied value, and the derived effective_ms intentionally remains the pre-clamp provenance. The preserved entry keeps the audit trail of what was deferred and how it closed. (From divisor-guard review of PR #950.)
