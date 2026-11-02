---
tag: docs-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:53:31Z
identity: docs-gate-20261102T045331-jay-flowers
tier: draft
---

In unbound-force repositories, the documentation gate is complete only when the filed docs-issue numbers are threaded back into the change record: the CHANGELOG entry's trailing reference (Docs: #NNN, unbound-force/website#NNN) and the proposal.md Impact Docs bullet. Filing the issues satisfies the gate's existence check, but when issues are filed after the CHANGELOG entry and proposal are authored — the common late-workflow ordering — the numbers never make it into the record, and the traceability chain exists only in GitHub keyword search. Treat issue filing as a two-step act: create the issue, then record its number in the CHANGELOG entry and proposal Impact section. Docs issues should also be created with the docs label so label-filtered triage (gh issue list --label docs) can discover them; the divisor-curator duplicate-issue check and the docs pipeline both key on that label.
