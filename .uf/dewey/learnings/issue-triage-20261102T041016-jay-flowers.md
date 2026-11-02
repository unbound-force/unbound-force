---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:10:16Z
identity: issue-triage-20261102T041016-jay-flowers
tier: draft
---

When triaging retro-agent issues that propose closing a documentation gap, verify three anchors before marking VALID: (1) the cited incident artifacts (PR, review comment, fix commit) match the current repository state; (2) the claimed documentation gap still exists — docs and code drift between retro time and triage time; (3) the proposed guidance operationalizes existing convention-pack rules rather than duplicating or contradicting them. Verified on unbound-force/unbound-force issue #609 (pattern-reuse guidance for AGENTS.md Testing Conventions): the guardrailBlock table-driven pattern exists in scaffold_test.go (now at ~line 10196, drifted from the ~6605 cited at retro time — hardcoded line references in issues age poorly), the duplicate standalone test TestContentPortability_TasksToIssuesGuardrail is absent (removed by fix 2179889), the "Content portability" entry sits in the blocks table with an issue-#593 reference comment, AGENTS.md Testing Conventions still lacks pattern-reuse guidance, and the proposal reinforces go.md TC-006 / default.md CS-004 rather than restating them.
