---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:15:39Z
identity: issue-triage-20261101T221539-jay-flowers
tier: draft
---

In unbound-force repos, duplicate candidates surfaced by keyword search on a command name frequently mix two distinct issue tracks: documentation-gate issues (title pattern "docs: <command> ... (<change>)", filed per the Documentation gate for each user-facing change) and content-pipeline tutorial issues (label: tutorial, gated by the tutorial significance threshold — new command, new workflow phases, human-decision gates, taxonomy, or iteration budget). They are complementary deliverables, not duplicates: docs issues track reference documentation for a specific change; tutorial issues track structured end-to-end workflow walkthroughs for the command itself. Before flagging a tutorial issue as a duplicate of a docs issue for the same command, verify the deliverable type and check the sibling-command precedent. Precedent: /uf.rebase-pr tutorial issue #855 shipped alongside that command's docs issue; /uf.fix-ci tutorial issue #870 ships alongside docs issues #862 (uf-fix-ci change) and #873 (extend-fetch-pr-review-context re-route). Verified in triage of #870 (2026-11-01).
