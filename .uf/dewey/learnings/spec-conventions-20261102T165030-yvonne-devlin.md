---
tag: spec-conventions
author: yvonne-devlin
category: pattern
created_at: 2026-11-02T16:50:30Z
identity: spec-conventions-20261102T165030-yvonne-devlin
tier: draft
---

OpenSpec spec deltas in the unbound-force meta repo use capability-prefixed requirement ids (MM-FR-001, UF-FR-001, RD-FR-019) and open with a requirement-numbering allocation header recording the allocated range and the reserved range for future changes of the same capability — codified as meta-prompt pack rules MP-002 and MP-014 and modeled by openspec/changes/model-matrix-delegation/specs/ and optional-profile-models/specs/review-matrix.md. Authoring a delta with bare Speckit-style FR-NNN ids and no allocation header deviates from both the pack and concurrent-change practice; read one sibling delta's header before allocating requirement numbers.
