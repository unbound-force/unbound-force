---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:15:36Z
identity: issue-triage-20261101T221536-jay-flowers
tier: draft
---

Content-pipeline duplicate checks must compare deliverable type, not just keyword overlap: in the unbound-force meta repo, tutorial-label issues and docs-label issues for the same command are complementary work items, not duplicates. Precedent: /uf.rebase-pr carries a tutorial issue (#855) alongside its command documentation; /uf.fix-ci carries docs issues (#862 command docs, #873 plugin-tool docs) plus tutorial issue #870, whose origin note explicitly cites the #855 asymmetry as the reason for re-disposition. A keyword search for the command name surfaces all of them — only a same-label, same-scope issue constitutes a duplicate. Verified in the #870 triage: candidates #862/#873 were correctly assessed as non-duplicates because they track reference documentation while #870 tracks a structured walkthrough of the command's six-phase workflow.
