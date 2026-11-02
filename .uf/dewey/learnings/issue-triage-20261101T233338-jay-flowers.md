---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:33:38Z
identity: issue-triage-20261101T233338-jay-flowers
tier: draft
---

Audit-filed issues can carry a scope mismatch between body and fix shape: the body records "adjacent gaps" beyond what the stated fix shape covers. Triaging unbound-force/unbound-force#825 (filed from the validated compiled/documentation-gate article's "Known open gaps" section) showed this shape — the fix shape covered the stale Website-gate checklist update and label-namespace standardization, while a third recorded gap (learnings not propagating into new change authoring without an enforcing checklist) appeared only in the body. Treat the stated fix shape as the issue's scope and recommend splitting body-only gaps into separately tracked issues; otherwise those gaps lose GitHub tracking when the primary issue closes, surviving only in knowledge-base articles that map to no actionable work item.
