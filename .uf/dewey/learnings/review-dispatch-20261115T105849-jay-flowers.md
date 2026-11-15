---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-15T10:58:49Z
identity: review-dispatch-20261115T105849-jay-flowers
tier: draft
as_of: "unbound-force@703a5229426874cca8cb05a2fef617910e355c11"
evidence:
  - ".opencode/test/finalize-error-boundaries.test.ts"
  - ".opencode/test/consolidate-dispatch.test.ts"
  - ".opencode/test/dispatch-agent-run.test.ts"
  - "CHANGELOG.md"
---

Closing a GitHub issue whose fix already landed on main follows a clean verification pattern in the unbound-force meta repo: route the work to the quick tier, create the verification branch from current main, confirm the fix commit is an ancestor of HEAD (git merge-base --is-ancestor <sha> HEAD), run the exact regression suites that pin the issue's acceptance criteria (for #1010: finalize-error-boundaries, consolidate-dispatch, and dispatch-agent-run — 171/171 green, each acceptance criterion mapped to a named test), then land a single docs-only CHANGELOG.md entry under Unreleased/Fixed recording the verification with (Spec: ...) and (Fixes: #NNN) references. The closure PR carries no code changes, so plan_review_dispatch classifies it documentation/lightweight (two-reviewer panel: adversary + guard at the tier cap) and make coverage-gate reports zero changed packages. This keeps the audit trail — repro, fix commit, regression coverage — in one reviewable place without re-implementing anything already merged.
