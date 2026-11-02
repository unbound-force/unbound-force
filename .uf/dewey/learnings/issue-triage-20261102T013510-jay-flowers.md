---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:35:10Z
identity: issue-triage-20261102T013510-jay-flowers
tier: draft
---

Documentation-gate issues are valid-by-obligation, not valid-by-outstanding-work: the AGENTS.md gate requires the record to exist for user-facing changes, so triage validity rests on the filing being corroborated — even when the issue's conditional acceptance criteria ("docs that enumerate the rules must mention them") are vacuously satisfied because no docs/ page enumerates the rule inventory. Verified on unbound-force/unbound-force#762: grep docs/ for the DE- rule-id prefix found only a passing DE-002 mention in a learnings review, so the criteria hold with no outstanding edits. Four repo anchors corroborate a gate filing end-to-end: the OpenSpec change directory, the CHANGELOG entry restating the version bump and rule ids, the convention-pack rule text, and the skill's named section; and because packs are auto-discovered into every Divisor review prompt, the reviewer's own prompt is direct evidence for "rules enter every review prompt" claims. Disposition for a satisfied-scope gate record: VALID plus a verification/closure note, not INVALID or NEEDS-CLARIFICATION.
