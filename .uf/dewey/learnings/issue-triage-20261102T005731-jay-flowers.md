---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:57:31Z
identity: issue-triage-20261102T005731-jay-flowers
tier: draft
---

When triaging an umbrella/epic issue, keyword-search duplicate candidates often turn out to be the umbrella's own sub-issues, cited inside its phase-tracking sections. Before classifying an umbrella as a duplicate, verify whether each candidate is referenced in the issue body's own tracking structure (a parent-child coordination relationship, not independent duplication), and whether the umbrella documents its relationship to prior tracking issues (e.g., an explicit "supersedes decision N" note reconciling an older issue's tracking approach). An umbrella whose phases each map to a dedicated open sub-issue is already decomposed — its split_recommendation should be null, and triage should verify the referenced sub-issue numbers exist and are open rather than assuming the tracking links are sound.
