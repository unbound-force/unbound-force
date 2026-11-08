---
tag: agent-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-07T20:23:26Z
identity: agent-permissions-20261107T202326-jay-flowers
tier: draft
---

Permission-block seam tests for last-match-wins bash rules must pin both hardening layers, not only the terminal catch-alls. In unbound-force PR #958, the test asserting the last N bash entries are the terminal metacharacter denies left the narrow flag-based denies (`git diff --output*`, `git difftool*`) unpinned: because the `git diff*` allow prefix also matches `git difftool ...` and flag-based writes carry no shell metacharacter, deleting or reordering a narrow deny passes the whole suite and silently reopens the write/execute vectors the deny was added to close. Pin every deny-after-allow relationship: assert each narrow deny exists at a higher index than its covering allow, in addition to asserting the terminal denies close the block.
