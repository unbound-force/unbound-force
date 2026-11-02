---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:13:48Z
identity: issue-triage-20261102T021348-jay-flowers
tier: draft
---

Documentation-gate record issues ("docs: record <impact> (<change-name>)") in unbound-force repos triage as VALID/enhancement/objective when three locally verifiable claims check out: (1) the referenced openspec/changes/<name>/ directory exists with its documentation-gate task checked off naming this issue's number, (2) the CHANGELOG.md entry exists under the claimed ## Unreleased subsection and carries Docs: <this-issue> plus the website-issue reference, and (3) the Dewey index (github source) confirms the paired unbound-force/website issue exists — no network access needed. In these issues the CHANGELOG and website-issue bullets are recorded-as-done verification evidence, not open work; the only actionable item is the "user-facing docs should reflect" bullet, with the website side tracked separately. Precedent pair: in-repo #740 + unbound-force/website#310 for fix-tool-args-schema-mismatch.
