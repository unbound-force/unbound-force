---
tag: content-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-04T20:08:39Z
identity: content-pipeline-20261104T200839-jay-flowers
tier: draft
---

Verifying a change's documentation-gate issues requires live gh issue list searches in both target repos, not tasks.md checkbox state. PR #957's task 4.3 (one docs-labeled issue in unbound-force/unbound-force plus one in unbound-force/website) was unchecked at the head commit, and label-scoped plus keyword searches (package-lock, plugin-lock) in both repos confirmed neither issue existed. Checkboxes lag reality in both directions — issues can be filed after the last commit, or the box checked with the issue  [… truncated at capture; final clause unrecoverable — recorded 2026-11-07].
