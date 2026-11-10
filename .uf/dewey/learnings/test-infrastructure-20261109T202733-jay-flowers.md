---
tag: test-infrastructure
author: jay-flowers
category: gotcha
created_at: 2026-11-09T20:27:33Z
identity: test-infrastructure-20261109T202733-jay-flowers
tier: draft
---

Golden output-shape test sites in this repo are tracked against the seam-verification change's site registry (openspec/changes/seam-verification/metrics.md, "45 of the 49 documented sites"). A tool added after that registry lands sits outside the count unless the registry is extended — and editing another change's metrics.md from a feature PR conflicts with commit-scope discipline. The checkout-pr-head change resolved this by placing its golden block in a separate describe with an in-file comment recording the placement decision; fold such sites into the registry when the seam-verification change is finalized or next extended so the audit stays authoritative.
