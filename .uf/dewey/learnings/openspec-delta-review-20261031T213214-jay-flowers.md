---
tag: openspec-delta-review
author: jay-flowers
category: pattern
created_at: 2026-10-31T21:32:14Z
identity: openspec-delta-review-20261031T213214-jay-flowers
tier: draft
---

When reviewing an OpenSpec MODIFIED delta, diff the scenario set against the predecessor requirement, not just the prose: MODIFIED replacement semantics silently delete any predecessor scenarios the delta omits, so a verbatim prose restatement can still drop the requirement's entire acceptance coverage. Before approving a MODIFIED requirement, read the predecessor file's same-named requirement and verify every predecessor scenario is either carried forward or consciously dropped with rationale.
