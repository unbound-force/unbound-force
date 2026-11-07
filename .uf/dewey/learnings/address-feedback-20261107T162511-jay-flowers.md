---
tag: address-feedback
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:25:11Z
identity: address-feedback-20261107T162511-jay-flowers
tier: draft
---

When maintainer feedback joins an in-PR artifact fix with a future-facing gate requirement (PR #968 feedback item 16: 're-export all 15 affected bodies ... AND gate future learning-batch commits on a mechanical completeness check'), the gate half often needs a different workflow tier — a CI check requires a spec workflow per AGENTS.md, so it cannot ship inside a docs-only chore PR without scope/phase violations. Resolving such feedback requires creating a tracked vehicle for the tier-mismatched half before the PR merges: a separate spec-governed change, or at minimum a tracking issue citing the feedback item. Without the vehicle, the required-before-merge gate silently evaporates at merge while the PR's own fix makes the feedback look addressed. Tell: feedback remediation whose verbs span both the current change's files and future commits' CI.
