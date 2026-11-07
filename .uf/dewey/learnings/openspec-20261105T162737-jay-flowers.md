---
tag: openspec
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:27:37Z
identity: openspec-20261105T162737-jay-flowers
tier: draft
---

When a change shipping no spec artifacts (e.g., a quick-tier PR) supersedes plugin-tool contracts pinned by live, unarchived OpenSpec deltas, a MODIFIED-only delta in a new retroactive change is not a durable reconciliation: the superseded MUST text lives in sibling deltas, openspec validation cannot flag cross-capability contradictions at archive time, and the outcome depends on which delta archives first. The durable pattern, per unbound-force precedents (unleash-deterministic-tools tasks 9.3/9.4; fix-tui-git-output-leak annotations inside finale-token-reduction/specs/plugin-tools.md), is threefold: (1) a retroactive OpenSpec change carrying MODIFIED/ADDED deltas for every touched capability, with MP-013 Previously paragraphs and MP-014 numbering that surveys unarchived siblings; (2) inline superseded-by annotations on the affected requirement text in each superseded live delta, so the annotation travels with the text whenever it archives; (3) a cross-change archive-order constraint recorded in both the superseding and superseded changes.
