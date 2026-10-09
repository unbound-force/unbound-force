---
tag: spec-review
author: jay-flowers
category: pattern
created_at: 2026-10-08T22:37:21Z
identity: spec-review-20261008T223721-jay-flowers
tier: draft
---

The OpenSpec spec review auto-fix policy (LOW and MEDIUM findings auto-fixed, HIGH and CRITICAL block) works well in practice. During the fix-stale-base-ref change, 10 findings (5 MEDIUM, 5 LOW) were auto-fixed in a single pass, covering missing constitution alignment sections, incomplete impact documentation, and missing documentation gate tasks. The auto-fixes were verified by the code review council which found no issues introduced by the spec-level fixes. Key auto-fixes included adding Principle V (Security by Default) assessment to proposal.md and adding documentation gate task 3.3 to tasks.md.
