---
tag: scaffold-drift
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:35:41Z
identity: scaffold-drift-20261102T013541-jay-flowers
tier: draft
---

Unbound Force's 2026-11 gitignore drift arose because scaffold-managed string blocks evolve without a drift test pinning them to the repository file: internal/scaffold/scaffold.go's gitignoreBlock (lines 1399-1421) was missing six entries (.uf/config.yaml, .uf/dewey/config.yaml, .uf/sibling-clones/, .uf/sibling-cache/, .uf/feedback/, .uf/progress/) that the repo's own .gitignore carried. Two-part prevention, embodied in the init-override-stubs-gitignore-sync change: (1) extract the managed block into a shared Go constant consumed by BOTH the generator and a drift test that asserts every managed line exists in the repo's live file (the TestRun_GitignoreBlockTracksRepositoryFile pattern, task 5.2); (2) never let a drift test re-declare the expected string — a test with its own copy of the block can pass while generated output is stale. When reviewing scaffold changes, check whether new managed lines ship with absence-checks in the shared-constant drift test in the same change.
