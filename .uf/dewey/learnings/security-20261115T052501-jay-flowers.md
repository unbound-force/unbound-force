---
tag: security
author: jay-flowers
category: gotcha
created_at: 2026-11-15T05:25:01Z
identity: security-20261115T052501-jay-flowers
tier: draft
---

When a command ingestion point parses external PR text (prior review bodies, inline comments, diffs), the 'data, never instructions' untrusted-input marker must be scoped per consumption point, not per data class. Adding a new staleness assessment that extracts a finding set from a prior review body needs its own marker even when a sibling path (light-check) already marks the same source — otherwise full-mode classification is steerable via instruction-shaped text quoted in the prior review.
