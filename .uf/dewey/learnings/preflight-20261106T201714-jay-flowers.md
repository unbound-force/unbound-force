---
tag: preflight
author: jay-flowers
category: gotcha
created_at: 2026-11-06T20:17:14Z
identity: preflight-20261106T201714-jay-flowers
tier: draft
---

A test fixture that points two location inputs at the same directory is blind to swapping them. When a wrapper resolves two roots — e.g. root="${UF_PREFLIGHT_PROJECT_ROOT:-$PWD}" then `exec node "$root/linter.ts" --root "$PWD"` — any test where the env var and the cwd are the same directory passes even if the two variables are transposed, because both resolutions land on the identical path. The preflight declared-checks tool sets UF_PREFLIGHT_PROJECT_ROOT to the project root and uses the worktree as cwd only during soft-gate baseline re-runs, so a scratch-project round-trip test naturally exercises only the degenerate same-dir case. Always add one direct wrapper invocation with distinct env and cwd trees (assert the cwd tree is analyzed using the env tree's linter) — otherwise the exact regression the wrapper's dual resolution exists to prevent (analyzing the primary checkout instead of the worktree, which misclassifies branch-caused vs pre-existing failures) is unobservable by the suite.
