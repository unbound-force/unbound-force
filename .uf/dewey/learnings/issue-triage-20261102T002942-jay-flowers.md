---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:29:42Z
identity: issue-triage-20261102T002942-jay-flowers
tier: draft
---

Umbrella-child issue triage: an "Umbrella: #N" header marks intentional phase decomposition, not duplication — the referenced umbrella is lineage, not a duplicate candidate. A valid phase child scopes to a single repo, enumerates its verification targets, defers adjacent work explicitly (e.g., docs updates to a later cutover phase), and states a binary done-when criterion (same committed tree works under both old and new tool versions). Issue #805 (website OpenCode v2 verification pass under umbrella #797) exemplifies the pattern.
