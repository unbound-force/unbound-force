---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:43:01Z
identity: issue-triage-20261102T004301-jay-flowers
tier: draft
---

Umbrella-coordinated phase issues filed in the umbrella's repository while targeting a sibling repository create pronoun ambiguity. Issue #803 ("Phase 3 — gaze: OpenCode v2 verification pass") is filed in unbound-force/unbound-force but uses "this repo" to mean the gaze sibling — the reference resolves only through context: the Goal names gaze, and "No plugins in this repo" matches gaze's config-only surface, whereas the meta repo's opencode.json registers a plugin array. When filing phase-tracking issues centrally for cross-repo migrations, name the target repository explicitly in each "this repo" clause (e.g., "the gaze repository") so triagers and executors reading the issue in the filing repo do not misread the execution target or the workflow tier the work will run in.
