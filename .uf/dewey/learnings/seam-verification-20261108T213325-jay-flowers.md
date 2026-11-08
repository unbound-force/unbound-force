---
tag: seam-verification
author: jay-flowers
category: pattern
created_at: 2026-11-08T21:33:25Z
identity: seam-verification-20261108T213325-jay-flowers
tier: draft
---

Extraction-based discoverability tests that pin a production regex/parser against a real repo file must assert the known-good set as a subset, not just non-emptiness plus existence. A `size > 0` + exists-on-disk check locks only the empty-set failure mode: partial extraction drift (one entry of a four-entry list reshaped while the rest stay extractable) stays green while the tool silently under-credits, and any design.md claim that "test-layout drift flips credit loudly" becomes false for that drift class. Verified shape (PR #959 round 3): tool-inventory.test.ts's registration-credit pin asserts non-emptiness against the live registration-surface.test.ts, whose PLUGINS array holds exactly four `{ id: ... }` entries; rewriting one entry to a tuple would pass the test while seamReport drops that plugin's registration credit. Pair every "drift flips loudly" design claim with a subset assertion naming the known entries (expect.arrayContaining or per-id has checks), which tolerates additive growth but fails on any loss.
