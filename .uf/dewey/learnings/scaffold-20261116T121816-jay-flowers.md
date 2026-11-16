---
tag: scaffold
author: jay-flowers
category: gotcha
created_at: 2026-11-16T12:18:16Z
identity: scaffold-20261116T121816-jay-flowers
tier: draft
as_of: "unbound-force@e451ac2821684173ef2593af3f6fe5b3b141fe43"
evidence:
  - "dewey:cmd/unbound-force/main_test.go"
  - "dewey:internal/scaffold/scaffold.go"
  - "dewey#91780bb9"
---

TestRunInit_FreshDir's file-count window in cmd/unbound-force/main_test.go lags the deployed scaffold surface whenever an asset is added without a window bump. The arithmetic: total embedded assets under internal/scaffold/assets minus 1 (devcontainer, never deployed by the walk) minus 4 (activation-gated plugin index.ts files, materialized only when Node/npm/OpenCode probes succeed) = base count; the window must be [base, base+4]. Verified 2026-11-16: 91 embedded with gh-error.ts, base 86, observed 90 with activation. The window had already lagged one asset on main (progress-state.ts, commit 4ef06e17) — main fails the test at 89 in any environment where activation materializes all four gated plugins, while CI environments without activation observe the base count and pass. When adding a scaffold asset, bump the window and its comment arithmetic in the same change (precedent: 91780bb9).
