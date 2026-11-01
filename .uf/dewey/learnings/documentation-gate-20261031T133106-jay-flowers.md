---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-10-31T13:31:06Z
identity: documentation-gate-20261031T133106-jay-flowers
tier: draft
---

Documentation-gate issue numbers have two recording surfaces in this repo: the CHANGELOG entry and the proposal's Documentation Impact section (see learnings documentation-gates-20261022T225014 and the port-uf-rebase-pr spec scenario: 'both issue numbers appear in the CHANGELOG entry and proposal.md's Impact section'). Changes whose tasks defer docs-gate classification to PR time (e.g., a task step saying 'record the decision in the PR description') reliably land the issue numbers only in the CHANGELOG entry — the proposal is never amended when the issues are filed, breaking the both-places pattern. Review heuristic: for any change citing 'Docs: #N' in its CHANGELOG entry, grep the proposal's Impact section for the same issue numbers; flag divergence at LOW severity before merge/archive. Verified in forward-aggregate-correlation-id code review (PR #883, 2026-10-31): the CHANGELOG entry cited #879 + unbound-force/website#348, but proposal.md's Impact section carried neither number.
