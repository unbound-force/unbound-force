---
tag: address-feedback
author: jay-flowers
category: decision
created_at: 2026-11-07T16:25:09Z
identity: address-feedback-20261107T162509-jay-flowers
tier: draft
---

When feedback on a docs-only chore PR requests a new mechanical gate (a CI completeness check over learning files), deferring to a separate spec-governed change is the tier-correct response — AGENTS.md requires a spec workflow for CI changes and the chore branch carries no spec artifacts, so bundling the gate would cross the tier boundary (meta-prompt.md MP-007/MP-008) and violate commit scope — but the deferral must be anchored or it evaporates: PR #949's truncation fix (cd603598) shipped without the requested gate and the identical defect recurred on PR #968. Anchored means a tracked follow-up issue or a retrospective entry naming the gate change; when that change lands, the check must be an executable tool with tests per determinism.md DE-001 (parse .uf/dewey/learnings/*.md, assert every body ends in terminal punctuation and no body sits at the ~512-character cap), not a prose instruction. Adding the gate strengthens governance and weakens nothing, so the AGENTS.md gatekeeping rule is not implicated.
