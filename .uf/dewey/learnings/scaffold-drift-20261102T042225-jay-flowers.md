---
tag: scaffold-drift
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:22:25Z
identity: scaffold-drift-20261102T042225-jay-flowers
tier: draft
---

When a sibling Go repo migrates its module path for a major version (e.g., unbound-force/gaze → github.com/unbound-force/gaze/v2, gaze#289), every `go install` command embedded in scaffold-owning repos goes stale silently: `go install <unsuffixed-path>@latest` can never resolve the new major version and keeps resolving the last v1 tag (v1.9.0). The fix must land in the scaffold owner's source templates — both the canonical `.opencode/` files AND their byte-parity embedded mirrors under `internal/scaffold/assets/` (drift-detection tests enforce the pair) — not in downstream patched copies that re-scaffolding overwrites. Verified in unbound-force/unbound-force#633: `.opencode/agents/cobalt-crush-dev.md:110` and `.opencode/commands/uf.review-council.md:604` plus their `internal/scaffold/assets/opencode/` mirrors still carried the unsuffixed command after gaze PR #290 migrated the module and updated gaze's own copies; gaze PR #296's known-issues note explicitly deferred the UF-scaffolded files to the UF issue.
