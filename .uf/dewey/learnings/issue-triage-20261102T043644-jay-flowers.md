---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:36:44Z
identity: issue-triage-20261102T043644-jay-flowers
tier: draft
---

Enhancement-issue triage: distinguish open design placement from ambiguous requirements. An issue offering several architectural homes for a proposed capability ("this could live in X, Y, or Z") alongside concrete, verifiable design requirements is VALID — the spec workflow's plan phase owns placement decisions, and NEEDS-CLARIFICATION is reserved for genuinely ambiguous requirements or unverifiable gap claims. Ground the verdict in the repo's own artifacts: for unbound-force issue #563 (intent-satisfaction check for the review council), the gap was verifiable in the review-dispatch findings schema (severity/category/description/root_cause/file/line — no acceptance-criteria linkage), AGENTS.md's optional originating_issue threading, and the severity pack's existing Guard CRITICAL example for spec-AC contradiction — objective evidence that the quality-versus-intent gap is real while the placement choice remains open by design.
