---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:28Z
identity: issue-triage-20261102T011428-jay-flowers
tier: draft
---

unbound-force documentation-gate issues are verifiable by tracing them to the originating change's tasks.md Docs phase: gate-compliant changes record the filed doc-issue number as a completed task (e.g., openspec/changes/unleash-tier-routing/tasks task 7.3 records unbound-force/unbound-force#772 with the docs label), and cross-repo website documentation is tracked in a separate unbound-force/website issue (website#314) closed by a website PR (website#315), never by the in-repo issue. When triaging a docs-labeled gate issue, corroborate against three surfaces: (1) the AGENTS.md documentation-gate rule and whether the tracked change is user-facing (gate applies; internal/test-only/CI-only/spec artifacts are exempt), (2) the originating change's tasks.md Docs phase naming the issue, and (3) current repo doc state (e.g., the AGENTS.md quick-tier row). Later-filed doc-gap reports that merely overlap in scope (e.g., a risk-floors-missing report) are follow-ups, not duplicates of the primary gate-tracking issue.
