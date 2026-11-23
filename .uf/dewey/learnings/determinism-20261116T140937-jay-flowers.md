---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T14:09:37Z
identity: determinism-20261116T140937-jay-flowers
tier: draft
as_of: "unbound-force@c9e35b82b9b5a4718abfc2f4bd320eb849aa60c1"
---

A DE-001 extraction (tool-derived field replacing an agent-supplied one) is best verified by a negative seam test that feeds a decoy agent-supplied value and asserts it never reaches durable state. Precedent: uf-workflow.test.ts "derives remaining_ms from the delegation-budget registry and round-trips it through get_progress" passes phase.remaining_ms: 1 in the call args and asserts the recorded state carries the tool-derived 440_000 — proving the contract no longer accepts agent-supplied values, not merely that the tool can derive one.
