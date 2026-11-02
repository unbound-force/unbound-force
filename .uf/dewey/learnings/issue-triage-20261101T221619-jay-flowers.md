---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:16:19Z
identity: issue-triage-20261101T221619-jay-flowers
tier: draft
---

A documentation-gate issue is objectively triageable when it enumerates per-artifact dispositions with explicit rationale, including justified "no change needed" entries. Each disposition (CHANGELOG entry, AGENTS.md table row and file count, docs/ page row, cross-repo website issue) maps to a verifiable repository state, and the disposition list can be cross-checked against the originating change's tasks.md Documentation phase — the two records must agree. This structure converts a subjective "documentation done" claim into checkable evidence and eliminates clarifying round-trips during triage. Precedent: unbound-force/unbound-force#856 (port-uf-rebase-pr) and #747 (command-prose-pilot) both follow this shape.
