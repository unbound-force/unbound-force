---
tag: review-process
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:24:57Z
identity: review-process-20261107T162457-jay-flowers
tier: draft
---

When reviewer feedback requires a new CI gate but the current PR's tier or commit-scope rules forbid bundling CI changes (AGENTS.md: CI changes require a spec workflow), the compliant response is a split — in-PR symptom fix plus a separate spec-governed gate change — and the split discharges the requirement only if the follow-up is filed and tracked at deferral time. An untracked 'should be a separate change' is indistinguishable from 'dropped' at merge. On PR #968 feedback item 16, the reviewer required a learning-completeness gate before merge; the sound approach deferred it to a spec-governed change but named no issue, change proposal, or owner. Required elements for the split to hold: file the follow-up before or with the PR's merge, reference it from the PR description or fix commit, scope it against the already-fixed root cause (cite issue #941 / PR #949's payload.proposals persistence so the gate is scoped as a detection backstop, not a redundant capture fix), and when the original item was marked required-before-merge, post the split rationale back to the reviewer thread for concurrence rather than assuming the split discharges it.
