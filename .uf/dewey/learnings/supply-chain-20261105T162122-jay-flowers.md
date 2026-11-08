---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:21:22Z
identity: supply-chain-20261105T162122-jay-flowers
tier: draft
---

Assessing maintainer feedback that asks for a resolved dist.integrity beside an exact-version npm pin (PR 957 precedent, `PLUGIN_TEST_NPM ?= npx -y npm@11.6.2`): the comment is audit documentation, not verification — npx still resolves name+version at run time, no tool enforces the comment, and every pin bump must manually re-resolve the hash or the record goes stale and misleads auditors. When the change's own proposal Impact section already records "there are no lock integrity hashes" as an accepted trade-off of a lockless install posture, the feedback is a documentation-completeness improvement inside an accepted-risk class, not a gatekeeping or constitution violation (npm exact versions are immutable under registry policy; the constitution's "pinning by name and version alone is insufficient" concern targets mutable tags and ranges). Classify such feedback AUTHOR-DECIDES with adoption guidance: record the sha512 via `npm view <pkg>@<version> dist.integrity` beside the pin together with the resolution command and the re-resolution-on-bump obligation — or decline citing the accepted-risk record.
