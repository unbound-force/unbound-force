---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-05T18:34:49Z
identity: openspec-20261105T183449-jay-flowers
tier: draft
---

In an OpenSpec delta-of-delta stack, a MODIFIED requirement's Previously paragraph quotes the requirement's current stack text — which may have been last amended by a sibling unarchived change rather than the original requirement author (e.g., a pin bump or rollback-SHA refresh). When writing the design's archive-ordering statement, enumerate every unarchived predecessor carrying the replaced text, not just the siblings whose files you edited: the archive step is a directory move with no automated canonical-text verification, so an unlisted sibling's delta can conflict silently.
