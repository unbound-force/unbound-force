---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:58:07Z
identity: issue-triage-20261101T215807-jay-flowers
tier: draft
---

Docs issues filed under a repository's documentation gate scan cleanest at triage when they scope to one feature surface and explicitly declare cross-references to overlapping in-flight docs issues. unbound-force issue #880 demonstrates the pattern: it scopes itself to the review-dispatch manifest-delivery surface, delegates the adjacent recovery-surface docs to sibling issue #766, and names the exact shared tools (redispatch_failed_runs, consolidate_dispatch) whose documentation must cross-reference both surfaces. Declaring that boundary at triage time prevents duplicate documentation work and scattered partial coverage across sibling issues.
