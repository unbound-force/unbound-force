---
tag: gatekeeping
author: jay-flowers
category: decision
created_at: 2026-11-06T18:48:59Z
identity: gatekeeping-20261106T184859-jay-flowers
tier: draft
---

A pinned-dependency security bump may amend the version literal inside its own unarchived OpenSpec delta in the same PR. The delta is the change's live spec text until archive; the gatekeeping rule's documented path (documented security fix, verified upstream integrity, gates preserved, documented rollback, reproducible validation, review + CI approval) is the authorization; and shipping unamended would archive a spec that mandates a known-vulnerable exact pin, forcing a zero-waste-violating follow-up change and leaving the merged manifest contradicting the archived requirement. A separate change adds ceremony while the red CI check keeps the PR unmergeable. Soundness conditions: the amendment is limited to version literals and override enumerations (not intent), sweeps all literal sites, and is documented in the CHANGELOG with GHSA ids and rollback.
