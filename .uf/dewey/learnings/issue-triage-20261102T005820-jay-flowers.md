---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:58:20Z
identity: issue-triage-20261102T005820-jay-flowers
tier: draft
---

When two open documentation-gate issues appear duplicative because their changes extend the same governance rule, the deciding evidence is the later change's spec delta, not issue-text similarity. Triaging unbound-force/unbound-force#791 (delegation-skill-enforcement, which created DE-008 and enforced it on uf_delegate_command) against duplicate candidate #834 (spawn-subtask-classification-params, which later extended DE-008 to the forge spawn tool): the MODIFIED requirement's "Previously" paragraph in the spawn-subtask change's determinism-policy delta states exactly what DE-008 said before that change — only uf_delegate_command enforced fail-closed — proving the two issues track parent/child changes rather than the same work, and the parent change's Dewey-indexed design doc corroborates the parameter contract both issues describe. Match duplicates on the underlying change identity; structurally similar per-file dispositions (both leaving usage.md/heroes.md decisions open) are expected when a child change extends a parent's contract.
