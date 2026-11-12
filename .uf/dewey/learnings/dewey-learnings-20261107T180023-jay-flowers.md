---
tag: dewey-learnings
author: jay-flowers
category: gotcha
created_at: 2026-11-07T18:00:23Z
identity: dewey-learnings-20261107T180023-jay-flowers
tier: draft
---

A retrospective learning that documents a defect in the change currently shipping is review evidence, not just knowledge capture. When the retrospective store step runs inside a change's own workflow, any learning whose text names a live, unresolved defect in the current PR must be triaged into a fix, a tasks.md item, or a tracked follow-up issue before the PR merges. Observed in unbound-force PR #957 (stop-shipping-plugin-lock): three retrospective learnings each documented a live defect in the PR — ci-cd-20261106T200302 (the throwaway-lock OSV compensation covered push/pull_request only, leaving the daily scheduled scan without npm transitive coverage), dewey-learnings-20261106T200258 (nine bulk-stored learning files truncated mid-payload, with re-storage prescribed but not applied), and spec-authoring-20261106T200300 (an SC-FR-005 requirement sentence broader than the implemented runtime-packages-only doctor check) — and all three shipped unactioned because the knowledge store has no defect-triage gate between capture and commit. Reviewers should grep a change's newly added learnings for phrases like "Seen in PR", "Observed in", and "still asserts" as a cheap defect-discovery pass over the change's own retrospective.
