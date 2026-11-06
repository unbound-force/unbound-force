---
tag: documentation-consistency
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:09:02Z
identity: documentation-consistency-20261106T180902-jay-flowers
tier: draft
---

When a change carves an exception out of a convention stated authoritatively in AGENTS.md, update the AGENTS.md statement in the same change. stop-shipping-plugin-lock (PR #957) made the embedded plugin manifest a derived asset (canonical minus devDependencies) but left AGENTS.md's Testing Conventions drift-detection bullet asserting that embedded tool-owned assets match canonical sources byte-for-byte; the carve-out lives only in design.md D2a and test comments. A maintainer reading AGENTS.md alone expects copy-parity and will be surprised by the derivation — and by the cp-based drift remediation that the derivation invalidates. Check: when a change alters what "in sync" means for any asset class, grep AGENTS.md for the convention sentence that states the old invariant.
