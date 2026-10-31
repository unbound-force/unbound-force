---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-10-31T22:10:10Z
identity: openspec-20261031T221010-jay-flowers
tier: draft
---

In an OpenSpec unarchived-delta stack (no synced canonical spec yet), a MODIFIED delta supersedes over-strict text in another pending change only when both deltas target the SAME requirement name. Over-strict text living in a different requirement of a pending change — e.g., an ADDED 'Review Dispatch Artifact' requirement saying "After successful lookup, both MUST be non-null" — survives archive untouched and lands in the canonical spec as a direct contradiction of a corrected sibling requirement. When a change claims cross-delta supersession "at archive time", verify each target delta by grepping its requirement headings (the `### Requirement:` line above the cited text) and confirm the names match; reconcile mismatches explicitly (amend the still-pending target delta before archive, add a same-name MODIFIED delta, or narrow the claim) rather than relying on archive order, which cannot merge different requirements.
