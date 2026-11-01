---
tag: openspec
author: jay-flowers
category: pattern
created_at: 2026-10-31T21:32:16Z
identity: openspec-20261031T213216-jay-flowers
tier: draft
---

When a change introduces a new convention pack (or any new governance convention), audit the change's own artifacts against the conventions it introduces before completion — self-compliance gaps are the most common structural defects and are mechanically findable. On the model-matrix-delegation spec review (three OpenSpec deltas + tasks.md shipping the new meta-prompt pack, MP-001..MP-024): the three deltas complied with the pack's allocation-header and id-line rules to three different degrees (only the model-matrix delta fully exemplified MP-013/MP-014; the unleash delta had no reserved range and no id line on its MODIFIED requirement; the scaffold delta omitted sibling locations); tasks 2.5 and 4.3 carried [P] markers contradicting the strict reading of the pack's own MP-017 rule (b); task 2.5 lacked the exact file path MP-016 mandates; and the new pack was never registered in AGENTS.md's Convention Packs inventory because Phase 7 had no task slot for it. A fixed self-audit step — diff each shipped artifact against every MUST rule in the pack being introduced — catches all four defect classes before review.
