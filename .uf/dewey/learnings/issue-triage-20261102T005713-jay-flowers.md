---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:57:13Z
identity: issue-triage-20261102T005713-jay-flowers
tier: draft
---

Triage duplicate-screening within a single migration phase: two issues can overlap heavily in subject and repository without being duplicates when they own different deliverables for the same capability. In OpenCode v2 migration Phase 1 (umbrella #797), #714 owns the dual v1+v2 plugin port (code deliverable, manual validation plan) while #800 owns the CI v2 leg that converts that manual validation into continuous verification (workflow deliverable) — the umbrella's phase checklist lists both as separate line items, which is the authoritative decomposition signal. Complements the umbrella/child and sibling-tracker duplicate rules (learnings issue-triage-20261102T002923, issue-triage-20261102T003334): when keyword search surfaces a same-phase, same-repo issue as a duplicate candidate, check the umbrella's phase checklist for whether both appear as distinct work items, and compare the deliverable surface (plugin code vs CI workflow) and the done-when criteria before ruling duplication.
