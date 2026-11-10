---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-09T20:27:29Z
identity: spec-authoring-20261109T202729-jay-flowers
tier: draft
---

New-capability OpenSpec deltas need the full MP-014 allocation header (allocated range, reserved future range, sibling-capability locations), but repo precedent is uneven — several merged changes carry no header at all, and the newest precedent omits the reserved-range line — so authors copying the nearest neighbor under-comply. When allocating a first-of-capability FR range, state the reserved continuation explicitly in the header comment (e.g. "AFC-FR-005+ reserved for future changes of address-feedback-checkout").
