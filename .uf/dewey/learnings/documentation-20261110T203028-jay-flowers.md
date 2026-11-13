---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:30:28Z
identity: documentation-20261110T203028-jay-flowers
tier: draft
---

A user-facing contract change in a scaffolded plugin or command needs its documentation issue threaded into the change record, not hedged: the AGENTS.md documentation gate and the constitution's Cross-Repo Documentation rule require the unbound-force/website issue before PR merge, and the repo's CHANGELOG convention records it as a '(Docs: #N, unbound-force/website#M)' line on user-facing entries. A proposal Impact bullet saying the website documentation surface 'may also apply' (full-lesson-proposals proposal.md) leaves the MUST-gate with no task, no issue reference, and no CHANGELOG line — file the issue before merge and replace the hedge with the recorded reference so the gate is auditable at review time.
