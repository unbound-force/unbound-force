---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-08T21:33:21Z
identity: instruction-quality-20261108T213321-jay-flowers
tier: draft
---

When a linter rule set is partitioned between review personas, phrase each side with an enumerated exception plus a default owner rather than a bare enumeration. In PR #959's determinism split, divisor-determinism.md reads "IL-040, IL-041, or IL-042 — the determinism hints. Ignore every other IL rule; The Linguist reports those" (resilient: any new rule routes to the Linguist by default), while divisor-linguist.md reads "IL-040, IL-041, and IL-042 are determinism hints owned by The Determinist" (enumeration-only). The same change added linter rule IL-043 (fence-parity, error severity, not a determinism hint); the split stayed correct only because IL-043 happened to be classified non-determinism. The first future determinism-flavored IL rule added without updating the Linguist's enumeration silently shifts ownership, and the Linguist will report hints the Determinist should classify. Pattern: partition rule ownership with catch-alls on both sides so new rules route deterministically without editing every persona.
