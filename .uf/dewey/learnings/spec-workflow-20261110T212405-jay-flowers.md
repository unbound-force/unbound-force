---
tag: spec-workflow
author: jay-flowers
category: pattern
created_at: 2026-11-10T21:24:05Z
identity: spec-workflow-20261110T212405-jay-flowers
tier: draft
---

OpenSpec requirement ids are the anchor surface for the unbound-force in-place amendment pattern: unarchived siblings modify a capability's requirements by id ("**ID: SC-FR-NNN**" opening each requirement body), and numbering-allocation headers record allocated and reserved ranges (init-override-stubs-gitignore-sync reserves SC-FR-016..018 "for future scaffold changes"). A delta that ADDS requirements without allocating ids — even "test-mandate" requirements — leaves them unamendable-by-id after archive and strands the reserved range the sibling recorded. Every house delta carrying requirements has the ID line and a numbering header (init-override-stubs-gitignore-sync is the complete-header example; stop-shipping-plugin-lock and merge-plugin-runtime-deps carry both even when allocating no new ids); MP-013/MP-014 codify both as MUSTs. Rule: ADDED requirements always get newly allocated ids from the capability's numbering inventory (checking the reserved range and any sibling takings first), and the delta opens with a complete allocation header naming the allocated range, the reserved range, and where sibling capabilities' requirements live.
