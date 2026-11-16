---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T15:41:03Z
identity: determinism-20261116T154103-jay-flowers
tier: draft
as_of: "unbound-force@c2dfd4b8f15f9ae106c599aef9a0418bf075dd93"
---

Pinning the expected findings profile of `openspec validate --strict` in tasks.md prose is a recurring mechanical guard across OpenSpec changes: unleash-interactive-clarify task 5.4 recorded a full exit-code + WARNING/ERROR/INFO profile in prose, and unify-checknotonmain-with-the-shared-branch-resolution task 3.3 requires "the established output shape ... no other findings" (the accurate prose-profile precedents are unleash-interactive-clarify and clarification-marker-superset, not compaction-storm-detector). The comparison of validator output against a fixed expected set passes the twice-test (same result every run), and no repo tool performs it (tool-inventory owner lookup for "openspec validate" returns nothing), so every change re-reasons the comparison from prose. Extraction candidate: a repo script taking the change name and expected findings and asserting the profile, making the guard tool-rendered per DE-001; until extracted, the prose guard is a pack-sanctioned interim mitigation reportable at LOW when newly added.
