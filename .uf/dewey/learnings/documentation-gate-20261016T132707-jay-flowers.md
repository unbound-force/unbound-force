---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-10-16T13:27:07Z
identity: documentation-gate-20261016T132707-jay-flowers
tier: draft
---

An OpenSpec documentation-gate task that names "this repository" as the sole documentation-issue target for a self-acknowledged user-facing workflow change satisfies the repo AGENTS.md gate while skipping the org constitution's Cross-Repo Documentation gate (unbound-force/website issue, required before the implementing PR merges). Naming a target is necessary but not sufficient: for user-facing changes the spec must name the constitutional target as well, or record a human-authorized sole-target reconciliation in the change artifacts (the review-voice-pack precedent). The fix-tool-args-schema-mismatch spec (tasks.md 4.2, design.md D7, FR-006) named the current repo only — it passes a superficial "a target is named" check while omitting the constitutional gate entirely. Reviewers verifying documentation tasks must check that the named target set covers both gates, not merely that a target is named.
