---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:42:32Z
identity: issue-triage-20261102T034232-jay-flowers
tier: draft
---

Issues that hard-code a file or agent inventory in their problem statement or acceptance criteria must be verified against the canonical registry and live files at triage time before marking VALID. Issue #660 claimed "All 5 Divisor agents" share review boilerplate, but the repo registers six review-capability Divisor agents (.uf/reviewer-capabilities.yaml: adversary, architect, curator, guard, sre, testing), and the omitted divisor-architect carries the identical Step 0 Prior Learnings block, dual-mode sentence, git-diff Review Scope text, [SEVERITY] Output Format skeleton, and APPROVE/REQUEST CHANGES Decision Criteria. Executing the acceptance criteria as written ("All 5 Divisor agents updated") would leave one reviewer agent in a drifted half-migrated state — inline duplicated boilerplate while the other reviewers reference the new skill. The content agents (scribe, herald, envoy) were correctly excluded because they lack the reviewer boilerplate, so the stale enumeration was a reviewer omission, not over-inclusion. Inventory enumerations in issues go silently stale as agents are added; checking them against the registry at triage time catches the drift cheaply, before an implementer codifies it in acceptance-criteria checkboxes.
