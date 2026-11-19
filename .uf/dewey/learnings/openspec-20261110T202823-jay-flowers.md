---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:28:23Z
identity: openspec-20261110T202823-jay-flowers
tier: draft
---

In the unbound-force OpenSpec workflow, a delta header saying "this change allocates no FR ids" is valid only for modify-only deltas that amend existing id'd requirements in place. A delta that ADDS new requirements must allocate ids from the capability's reserved range and give each requirement a "**ID: XX-FR-NNN** This requirement MUST be satisfied." body line (meta-prompt pack MP-002/MP-013/MP-014). The scaffold capability's chain reserves SC-FR-016..018 for future changes (init-override-stubs-gitignore-sync header); copying the "allocates no ids" wording from a modify-only sibling (as plugin-lock-advisories initially did) leaves new requirements unamendable through the "modifies SC-FR-0NN in place" protocol and orphans the reservation. When authoring a delta, read the target capability's most recent allocation header and allocate from — and update — the reserved range.
