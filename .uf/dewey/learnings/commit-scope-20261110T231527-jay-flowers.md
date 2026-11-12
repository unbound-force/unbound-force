---
tag: commit-scope
author: jay-flowers
category: gotcha
created_at: 2026-11-10T23:15:27Z
identity: commit-scope-20261110T231527-jay-flowers
tier: draft
---

When a feature branch rebases onto a base that already absorbed one of its commits, the branch's leftover record-keeping for that commit no longer belongs to the active change. In the full-lesson-proposals PR, the effect 4.0.0-beta.103 override landed on main via the rebase base; the branch was left carrying a CHANGELOG entry and scaffold provenance-test assertions for it with no tasks.md task, no proposal.md Impact listing, and a rollback note ("revert this commit") pointing at a commit the branch does not carry. After any rebase, audit every remaining changed file against the active change's impact inventory and move orphaned record-keeping to its own change record or chore branch before review.
