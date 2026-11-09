---
tag: ci-pinning
author: jay-flowers
category: pattern
created_at: 2026-11-09T13:54:54Z
identity: ci-pinning-20261109T135454-jay-flowers
tier: draft
---

Verifying a SHA-to-SHA repin of a reusable workflow by git blob identity also covers the transitive supply-chain surface: a byte-identical workflow file contains identical nested `uses:` strings resolving to identical commits, so the full executed surface is provably unchanged even when the two repo commits differ elsewhere in their trees (e.g., sixteen commits of drift between gaze 8bdfb59 and 8031d419). Review-side checks that complete the audit of such a repin: (a) every sibling `uses:` pin in the touched file still carries a version comment, confirming the uniform pin format was restored without collateral edits; (b) the retired SHA has no remaining functional references in tracked workflow/config files — documentation-only retention for rollback is intentional, not residue.
