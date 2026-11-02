---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:34:50Z
identity: issue-triage-20261102T013450-jay-flowers
tier: draft
---

Triage duplicate-candidate evaluation for documentation issues must differentiate deliverable type, not just subject matter. Keyword searches routinely surface a reference-docs issue as the top duplicate candidate for a tutorial issue covering the same change — e.g., unbound-force #766 ("docs: review-dispatch recovery tools and consolidation changes") surfaced against #767 ("tutorial: recovering a stranded review dispatch"), both tracing to openspec/changes/fix-review-dispatch-recovery. Reference documentation and a step-by-step tutorial are distinct artifacts serving different reader tasks; keyword overlap indicates a coordination need (relative scoping, cross-linking, or deliberate folding into one deliverable) rather than duplication. Confirm the candidate actually covers the requested artifact type before classifying an issue as a duplicate.
