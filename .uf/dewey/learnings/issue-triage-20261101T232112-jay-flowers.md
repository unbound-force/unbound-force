---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:21:12Z
identity: issue-triage-20261101T232112-jay-flowers
tier: draft
---

Docs-issue triage verification pattern for interaction-model changes: when a change alters an autonomous command's interaction model (e.g., /uf.unleash Speckit clarify moving from exit-on-ambiguity to an interactive ask gate with defer fallback), every docs passage enumerating that command's exit conditions goes stale simultaneously — the same superseded sentence appeared verbatim in both docs/usage.md ("Go Fully Autonomous") and docs/architecture.md ("Autonomous Mode"), as surfaced by issue #840. Verify by (1) grepping docs/ for the quoted superseded sentence rather than trusting cited line numbers (line citations drifted ~20 lines within days of filing; verbatim sentence quotes are the durable reference), (2) confirming the source change's tasks.md shows the behavior actually landed (all checkboxes complete with recorded CI-parity runs), and (3) checking whether the docs issue was filed by the change's own documentation gate with a consolidate-with-overlapping-issue strategy (#840 was filed by unleash-interactive-clarify task 6.2, which deliberately made no in-change docs edits and directed consolidation with #739's rewrite of the same sections to avoid competing edits and merge conflicts).
