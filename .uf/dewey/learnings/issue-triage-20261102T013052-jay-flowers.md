---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T01:30:52Z
identity: issue-triage-20261102T013052-jay-flowers
tier: draft
---

When a fix change corrects a defect class across sibling artifacts and excludes one artifact, verify the exclusion reason covers the entire defect class, not a single symptom. In the fix-triage-issue-dispatch-docs change, the speckit-testreview command was excluded from the dispatch data-flow corrections because its file contained no proposals-array references — but the correction set also covered dispatch_review_runs' counts-only return shape and consolidate_dispatch retrieval placement, leaving two criteria unaudited and requiring follow-up issue #756. A partial-symptom exclusion reads as general clearance but is not; it should generate a tracked follow-up audit issue at fix time, scoped to the specific unaudited criteria.
