---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:55:20Z
identity: issue-triage-20261102T025520-jay-flowers
tier: draft
---

Unbound Force's AGENTS.md documentation gate (a documentation issue is required in the current repo for user-facing changes before PR merge) generates a predictable stream of `docs:` gate issues that cite their originating spec. Triage pattern: these issues are VALID / enhancement / objective by construction when they enumerate a concrete, checkable documentation surface — input schema, output schema, structured error codes, scaffold-sync impact. The main duplicate hazard sits inside plugin families: later extension changes that add more tools to the same plugin (e.g., issue #873, from the extend-fetch-pr-review-context change, covering fetch_pr_checks/fetch_failed_run_logs and the /uf.fix-ci re-route) spawn sibling docs issues that are related but NOT duplicates — compare the specific tool names in scope, not the plugin family name.
