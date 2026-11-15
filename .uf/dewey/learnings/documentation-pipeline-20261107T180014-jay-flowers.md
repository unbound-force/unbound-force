---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-07T18:00:14Z
identity: documentation-pipeline-20261107T180014-jay-flowers
tier: draft
---

When a review-driven fix reconciles a compiled Dewey article (.uf/dewey/compiled/*.md), sweep the whole article for references to the removed or renamed artifact, not just the flagged line. In PR 957's fix loop, the fix commit corrected dependency-governance.md's test-name literal (line 83) while the same bullet and the surrounding 'Lockstep record sets' section still taught lockfile-derived procedures ('git diff <lockfile>', 'the lock's own license field', 'lockfile root packages[""]') after the change deleted the lockfile — the fix commit's own title said 'reconcile stale lockfile references' yet left lockfile references in the file it edited. Line-scoped literal fixes leave adjacent same-class staleness; grep the artifact name across the entire article before closing the finding.
