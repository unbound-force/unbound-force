---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:54:42Z
identity: issue-triage-20261102T035442-jay-flowers
tier: draft
---

RFC-style enhancement issues that pair a concrete problem statement, evidence-backed elimination of alternative approaches, and an implementation plan with a transparently enumerated list of open questions should triage as VALID. Open questions inside an RFC are design solicitation, not missing information: NEEDS-CLARIFICATION applies only when the issue's intent, scope, or expected behavior cannot be determined from the content. Before accepting, verify separately that the applied label matches the content (enhancement when the change adds capability to an existing workflow rather than fixing a defect), the affected artifacts live in the target repository, and the duplicate-issue search is clean. Example: unbound-force/unbound-force#645 (session directory relocation via plugin, 7 enumerated open questions) triaged VALID / enhancement / objective with no split recommended — the open questions were answered through proposal discussion, not by re-opening triage.
