---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-03T16:46:17Z
identity: review-dispatch-20261103T164617-jay-flowers
tier: draft
---

Before changing a plugin tool's observable error code, grep the unarchived OpenSpec change deltas (openspec/changes/*/specs/) for the existing code literal — those deltas are live contracts for the capability even before archiving. If the new code refines a failure case the spec scenarios do not cover (PR #937: diff_not_found splitting off from path_validation_failed for missing-in-allowed-directory diff paths), the existing scenarios usually remain accurate — FR-011 in build-review-prompts governs locations where those codes surface, so grep both the deltas and the consuming requirement text before the rename.
