---
tag: scaffold
author: jay-flowers
category: gotcha
created_at: 2026-11-15T22:16:36Z
identity: scaffold-20261115T221636-jay-flowers
tier: draft
as_of: "unbound-force@6a553043168ccaffc4677672b76b30f5453aa1b5"
evidence:
  - "cmd/unbound-force/main_test.go"
  - "internal/scaffold/scaffold.go"
  - "openspec/changes/review-dispatch-errors-cap-loses-unrecov/tasks.md"
---

Adding any tool-owned embedded asset under internal/scaffold/assets/ requires bumping the uf init file-count window in cmd/unbound-force/main_test.go TestRunInit_FreshDir: the assertion counts files created in a fresh directory within an exact numeric window ([84-88] after RD-FR-066 on branch opsx/review-dispatch-errors-cap-loses-unrecov, 2026-11-15). Every new embedded asset shifts the count by +1 and fails make check with a count mismatch until the window (and its lagging comment arithmetic, which claims to derive the count) is updated. Check this window first when a scaffold-asset change fails the gate — it was the only change-caused failure in the RD-FR-066 change (fix: a62fde4f, window [83-87] to [84-88]).
