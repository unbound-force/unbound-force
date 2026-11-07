---
tag: dewey-learnings
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:10:44Z
identity: dewey-learnings-20261107T161044-jay-flowers
tier: draft
---

A learning shipped inside a PR that prescribes an action applying to that same PR creates a verifiable obligation — reviewers should grep the PR's own new learnings for prescriptions and check whether each was executed before merge, because the learning's existence proves the author knew the bar and the gap it describes is usually still open at head. Observed on unbound-force PR #957 (stop-shipping-plugin-lock): the PR ships learning documentation-pipeline-20261106T200256, which documents that the zod 4.1.8→4.6.6 sweep left .uf/dewey/compiled/plugin-architecture.md line 75 asserting the old pin and prescribes either including .uf/dewey/compiled/ in the sweep path set or running an incremental dewey_compile — and at head 81912a90 neither happened (the compiled article still says zod 4.1.8, and dependency-governance.md still carries the pre-rename test name). The shipped learning converts what would be a speculative "derived docs may drift" concern into a data-driven finding with a pre-written recommendation. General rule: for any PR carrying new .uf/dewey/learnings files, read them for self-referential prescriptions (sweep path sets, compile/refresh steps, follow-up actions naming the PR's own artifacts) and verify execution against the frozen head; extends the existing data-driven-feedback pattern (a feedback item backed by a same-PR learning) from feedback triage to first-round review.
