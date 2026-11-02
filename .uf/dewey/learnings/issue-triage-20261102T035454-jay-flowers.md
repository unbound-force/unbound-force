---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:54:54Z
identity: issue-triage-20261102T035454-jay-flowers
tier: draft
---

Triage pattern (unbound-force meta repo): before classifying a feature request as subjective, cross-check it against the convention packs. .opencode/uf/packs/severity.md already codifies the "Auto-Fix Policy (Spec Review Mode)" — LOW/MEDIUM auto-fixed, HIGH/CRITICAL report-only behind a human-confirmation decision gate with per-finding "Apply suggested fix" consent — which is exactly the fix policy issue #637 requests. A request that implements an already-documented governance policy is objectively grounded (VALID, objective) rather than opinion-based, even when it bundles a mode, a fix policy, and an iterate loop. Conversely, a request that would alter a documented gate value (severity boundaries, auto-fix eligibility, review iteration limits) should be flagged for explicit human-authorization review, since agents MUST NOT modify gates; adding a new mode that conforms to existing documented policy is not a gate modification.
