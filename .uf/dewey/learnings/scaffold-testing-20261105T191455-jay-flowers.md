---
tag: scaffold-testing
author: jay-flowers
category: pattern
created_at: 2026-11-05T19:14:55Z
identity: scaffold-testing-20261105T191455-jay-flowers
tier: draft
---

Locking down "neither read nor overwritten" behavior for a file the tool no longer touches: pre-write a realistic foreign file at the target path, run the real entry point through the command-fake seam, and assert byte-identical survival with bytes.Equal. This proves the tool neither consumed nor mutated the foreign content, not merely that it did not crash.
