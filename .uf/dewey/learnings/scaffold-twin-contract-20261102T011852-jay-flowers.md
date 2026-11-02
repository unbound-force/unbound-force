---
tag: scaffold-twin-contract
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:18:52Z
identity: scaffold-twin-contract-20261102T011852-jay-flowers
tier: draft
---

Module extraction of a dual-maintained byte-identical file pair (canonical .opencode/plugins/review-dispatch/index.ts ↔ embedded scaffold twin internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts) must keep the drift-detection test (TestEmbeddedAssets_MatchSource, internal/scaffold/scaffold_test.go:96) as the parity gate of record across the new module layout. Two viable strategies exist: lockstep module extraction in both copies, or build-time twin generation with the drift test converted to a staleness/determinism check of the generation output — the choice belongs in an OpenSpec design doc, because the second option re-points a governance gate. Scale compounds the risk: the file grew from 5,992 lines at issue filing (2026-10-19) to 8,634 lines by 2026-11-01, with each change requiring a manual lockstep copy (the cp-command Dual-Copy Sync task pattern) verified only after the fact by CI drift tests. Large hand-synced twins make partial updates likely and lockstep diffs impractical to review, so decomposition is resilience-positive exactly as long as the drift gate survives the restructuring. Related prior learning: scaffold-parity-20261016T151133-jay-flowers (git diff index pairs as zero-cost parity verification).
