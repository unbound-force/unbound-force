---
tag: dewey
author: jay-flowers
category: pattern
created_at: 2026-11-07T15:50:18Z
identity: dewey-20261107T155018-jay-flowers
tier: draft
---

Pre-commit completeness checks for standalone Dewey-learning batches: the uniform +10/-0 diffstat of learning files hides two provenance defects, so gate the commit on two mechanical checks. (1) Body integrity: every learning body must end in a complete sentence — capture sessions that persist bodies from a 512-character-capped source (a tool-result preview rather than the dispatch artifact's payload.proposals) produce mid-sentence cuts with no truncation marker, and the corruption splits cleanly by session (PR #968: all 7 files from the 2026-11-03 21:18 session and all 8 from the 2026-11-04 22:58 session were cut while every other session's files were complete; the same defect was already self-documented for PR #949's batch). (2) Cluster duplication: files created within seconds of each other restating one insight indicate per-persona proposals stored without the near-identical merge (PR #968: seven clusters covering 22 of 50 files). Both symptoms share one root cause — persistence from a capped, unmerged source — so observing either symptom warrants checking the other.
