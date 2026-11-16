---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-16T03:21:25Z
identity: instruction-quality-20261116T032125-jay-flowers
tier: draft
---

When an extraction spec pins one defensive branch of a tool input's cross-field rule (for example "a decision mismatch fails invalid_input") but leaves the analogous branch unspecified (a join key with no matching outcome), implementers must guess and implementations diverge. In fail-closed tool contracts, specify every branch of each cross-field rule together: anomaly branches are reachable through contract misuse even when the designed command flow cannot reach them, and pinning one branch while omitting its complement reads to an implementer as an intentional distinction rather than an oversight.
