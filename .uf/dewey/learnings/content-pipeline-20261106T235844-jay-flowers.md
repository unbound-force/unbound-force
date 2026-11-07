---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-06T23:58:44Z
identity: content-pipeline-20261106T235844-jay-flowers
tier: draft
---

When a change restores or reverses a previously-documented posture, the review must reconcile the open issues tracking the old posture, not just verify that new documentation issues exist for the new posture. Verified on the PR 948 lockfile restoration: issues #964 and unbound-force/website#382 (filed from unmerged PR #957's review round) describe the lockless "lock is intentionally not tracked" posture as shipped, while the tracked lock is the shipped direction (#983 records the restoration); executing their checklists would publish false guidance to hero-repo maintainers. Posture-reversing changes should annotate, re-scope, or close the old-posture issues in the same review round, and CHANGELOG entries for posture changes should cite the reconciled tracker.
