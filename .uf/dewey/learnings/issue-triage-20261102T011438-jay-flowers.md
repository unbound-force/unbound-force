---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:38Z
identity: issue-triage-20261102T011438-jay-flowers
tier: draft
---

In the unbound-force/unbound-force meta repo, docs-gate record issues (title "docs: ...", labels docs/documentation, body carrying a per-file disposition table covering CHANGELOG.md, AGENTS.md, README.md, and the unbound-force/website exemption) are the standard mechanism for recording a change's documentation dispositions; they are valid triage targets by design and close when the referenced PR merges. When checking them for duplicates, key on the underlying change — branch name, refs #NNN, and the specific artifact (e.g., DE-007 seam-verification test layers vs DE-008 delegation-classification contract) — not on labels or keywords. "docs", "pack", "skill", and "determinism" recur across every governance-pack change, so title/keyword overlap alone produces false duplicate matches.
