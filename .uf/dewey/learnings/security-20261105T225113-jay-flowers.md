---
tag: security
author: jay-flowers
category: pattern
created_at: 2026-11-05T22:51:13Z
identity: security-20261105T225113-jay-flowers
tier: draft
---

An input invariant enforced at one acquisition site does not cover the same identifier acquired elsewhere. PR #955's digits-only PR_NUMBER gate sits in argument parsing (the user-provided form), but the omitted-argument path auto-detects the number via `gh pr view --json number --jq '.number'` (structured output — practically safe) and the new Section 1.2.1 `gh pr checkout <PR_NUMBER>` template consumes it without re-validation, while the command's own charset-gate paragraph promises validation for "every PR-derived value before it reaches any shell template." State input invariants at the point of first security-sensitive use, or repeat them after every acquisition site; otherwise the invariant is source-scoped rather than value-scoped, and the next acquisition path added re-opens the gap the gate was written to close.
