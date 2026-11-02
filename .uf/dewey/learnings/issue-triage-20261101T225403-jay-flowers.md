---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:54:03Z
identity: issue-triage-20261101T225403-jay-flowers
tier: draft
---

When triaging a documentation-gate issue whose dispositions contain a conditional or "not known" entry (e.g., #844's "Website: no website tracking issue known for this change; link one here if the website documents uf doctor output"), read the parent change's tasks.md before considering NEEDS-CLARIFICATION. Completed gate tasks record their discharge inline — observed in openspec/changes/replicator-version-floor/tasks.md task 3.2: "(Filed: unbound-force/unbound-force#844; website tracking issue: unbound-force/website#337.)" — so tasks.md resolves a stale disposition in the issue body without asking the author. Gate-issue bodies are written at filing time and can lag the change's recorded state; the parent tasks.md is the authoritative cross-check, and the issue's own "confirm or update dispositions in the implementing PR" acceptance criterion is the vehicle for folding the resolved value back in.
