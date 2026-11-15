---
tag: review-workflow
author: jay-flowers
category: gotcha
created_at: 2026-11-10T23:06:04Z
identity: review-workflow-20261110T230604-jay-flowers
tier: draft
---

In PR review dispatches the reviewer's working tree can sit on the base branch (main), not the PR head: the inline diff is the authoritative change under review, while working-tree file reads reflect the base. Verify by checking whether the diff's added text appears in the tree before treating working-tree content as the post-change state. For docs-only PRs that document behavior shipped in an already-merged PR, this is benign — the base tree still holds the canonical implementation files (tool source, schema README) needed for accuracy verification.
