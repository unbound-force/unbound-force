---
tag: openspec-conventions
author: jay-flowers
category: gotcha
created_at: 2026-11-08T23:16:29Z
identity: openspec-conventions-20261108T231629-jay-flowers
tier: draft
---

The meta-prompt pack's OpenSpec rules diverge from living repo practice, and reviewers must calibrate against precedent, not pack text alone (verified 2026-11-08 while reviewing PR #990): MP-014's requirement-numbering allocation header appears in zero changes under openspec/changes/ (grep "Requirement Numbering": no matches), and MP-013's FR-id body line is absent everywhere (ids appear only inconsistently as heading prefixes — FR-001 in review-pr-confirmation-gate, MR-001 in fix-review-pr-step7-gate, plain names elsewhere). The living conventions worth enforcing: all three delta sections with explicit "None." closers, a proposal "## Constitution Alignment" section (100+ proposals), "### New Capabilities" for fresh capabilities (parent review-council-github-posting proposal), and numbered design decisions (D*/R*). The pack-vs-practice drift itself needs a governance revision of meta-prompt.md rather than per-change findings.
