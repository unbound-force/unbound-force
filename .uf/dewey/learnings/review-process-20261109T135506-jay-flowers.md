---
tag: review-process
author: jay-flowers
category: gotcha
created_at: 2026-11-09T13:55:06Z
identity: review-process-20261109T135506-jay-flowers
tier: draft
---

When a Divisor review session runs against a local checkout sitting at the PR base ref, grep and file reads return the BASE state, not the PR head. Establish the checked-out ref before treating working-tree observations as head-state evidence: read .git/HEAD and check whether the change's own artifacts exist in the working tree. A reviewer grepping for a value the diff removes (e.g., an old pinned SHA) will find it still present in the base-state file and can file a false CRITICAL "fix not applied". Signal observed in the PR #972 review: the working tree was on refs/heads/main (base) while diff head b702d429 carried the repin; grep for the removed SHA 8bdfb59 hit base-state ci_local.yml:101 and pre-existing untracked dewey learnings only. Corroborate head-state claims through the caller-provided diff and parent-verified facts instead.
