---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-10-31T21:32:23Z
identity: openspec-20261031T213223-jay-flowers
tier: draft
---

A change that authors a new convention pack must be audited against that pack's own MUST rules — self-application gaps are systematically missed because the rules and the artifacts they govern land in the same diff and no baseline comparison exists yet. On the model-matrix-delegation spec review (2026-10-31), the change shipped meta-prompt.md (MP-001 through MP-024) while its own artifacts missed three of its MUSTs: the unleash delta's numbering header omitted MP-014's reserved-range element, tasks.md lacked MP-018's per-phase run-the-suite checkpoints, and — the severe cousin — the MODIFIED delta restated predecessor prose while dropping 14 of 16 scenarios, which was caught only because a reviewer diffed the scenario set rather than the prose. Review procedure: when a diff both adds a rules artifact and adds or edits artifacts those rules govern, explicitly run the new rules against the sibling artifacts in the same diff before approving; for OpenSpec MODIFIED deltas specifically, always diff the scenario set against the requirement being replaced, not just the normative prose.
