---
tag: risk-documentation
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:09:40Z
identity: risk-documentation-20261105T160940-jay-flowers
tier: draft
---

An OpenSpec design.md risk bullet must carry a mitigation record or an explicit acceptance/scope decision, not just the consequence. A bullet that states only the loss leaves the constitution's "risk trade-offs recorded with their compensating controls or accepted explicitly" standard unmet and gives feedback triage no "already documented as accepted risk" defense. In PR #957 (stop-shipping-plugin-lock), the "Transitive drift" bullet recorded mitigations and the "No byte-for-byte reproducible install" bullet recorded an explicit out-of-scope line, while the "Lock-based scanners" bullet recorded only the loss plus residual direct-dependency coverage — the single bullet the Divisor council re-confirmed as a finding across two reviews. Assessment pattern for accepted-risk defenses in feedback triage: diff each risk bullet individually for a mitigation-or-acceptance record; the Risks section's presence is not the acceptance itself. Sibling bullets that close their loop make the open one visible by contrast. Related: compensating controls that generate ephemeral state in CI (for example a throwaway package-lock.json for OSV-Scanner, discarded after the scan) can restore lost detective coverage without violating a MUST-NOT-track spec requirement, because an ephemeral CI artifact is neither tracked nor shipped.
