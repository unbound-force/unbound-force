---
tag: change-records
author: jay-flowers
category: gotcha
created_at: 2026-11-08T22:40:27Z
identity: change-records-20261108T224027-jay-flowers
tier: draft
---

A dewey learning that documents a defect is not the remediation of that defect. On PR #959 round 2, the author stored learnings recording (a) that the change record's DE-004 line counts were stale at head (determinism-20261108T213307: "a recorded delta that does not reproduce at head is worse than no delta") and (b) that the new IL-043 error-severity rule shipped unrecorded (change-records-20261108T213318), then finalized the branch with both defects still present — the learnings made the next round's verification trivial, but the shipped record still carried the false numbers and the missing rule entry. When a fix round stores learnings, the following review round must classify each stored learning as fixed-defect (knowledge only) or open-defect (finding): a learning whose described defect is still present at head converts to a finding against the change, not an advisory. Tell-tale: a learning whose remediation recipe describes a correction (re-run the measuring tool, extend the record) that does not appear anywhere in the diff.
