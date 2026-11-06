---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-05T17:01:34Z
identity: spec-authoring-20261105T170134-jay-flowers
tier: draft
---

OpenSpec delta structure is convention-enforced only — nothing in CI validates it, so structural deviations pass every check. The repo's established structure (MP-002/MP-012/MP-013/MP-014 in .opencode/uf/packs/meta-prompt.md) is: one specs/<capability>/spec.md file per capability (directory form, never a flat specs/<capability>.md), each requirement body opening with a '**ID: <CAP>-FR-NNN** This requirement MUST be satisfied.' line, and each delta opening with a 'Requirement numbering:' allocation header re
