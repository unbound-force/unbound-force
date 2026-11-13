---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:08:02Z
identity: issue-triage-20261110T190802-jay-flowers
tier: draft
---

In unbound-force/unbound-force the documentation-issue label namespace is split between "docs" and "documentation" (tracked by open issue #825, "label namespace unstandardized"), so a label-scoped duplicate check (gh issue list --label docs) can miss the tracking issue for the very change under review: issue #1034 (effect security bump documentation, filed 2026-11-10 per the documentation gate after PR #1024's council HIGH finding) carries only the "documentation" label and is invisible to a "docs"-label search. Documentation-gate duplicate checks in this repo must combine the label-scoped listing with a keyword search across all open issues (--search "<package or change name>"), otherwise a satisfied documentation gate gets re-flagged as missing and a duplicate issue gets filed.
