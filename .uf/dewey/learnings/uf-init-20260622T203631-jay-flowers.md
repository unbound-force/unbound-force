---
tag: uf-init
author: jay-flowers
category: gotcha
created_at: 2026-06-22T20:36:31Z
identity: uf-init-20260622T203631-jay-flowers
tier: draft
---

When modifying the /uf-init slash command (.opencode/commands/uf-init.md), the scaffold asset copy at internal/scaffold/assets/opencode/commands/uf-init.md must be kept byte-identical. This was caught during spec review by multiple Divisor agents referencing prior learnings. The drift detection test TestEmbeddedAssets_MatchSource in internal/scaffold/ will fail if the copies diverge. Always run `cp .opencode/commands/uf-init.md internal/scaffold/assets/opencode/commands/uf-init.md` followed by `go test -race -count=1 -run TestEmbeddedAssets_MatchSource ./internal/scaffold/` after any changes to uf-init.md.
