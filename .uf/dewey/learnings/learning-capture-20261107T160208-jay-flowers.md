---
tag: learning-capture
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:02:08Z
identity: learning-capture-20261107T160208-jay-flowers
tier: draft
---

Learning-artifact truncation clusters by capture batch and is detectable from the review diff alone: read each learning file's body line and check for terminal punctuation. On PR 948, all 20 truncated .uf/dewey/learnings/*.md files fell in the 2026-11-03T21:31, 2026-11-04T20:08, and 2026-11-05T14:45 capture batches (bodies clipped mid-token roughly 500 characters in) while every capture from 2026-11-05T16:15 onward was complete — the capture path changed mid-day, but the earlier clipped artifacts were never re-stored. The committed .md and the live Dewey index carry the identical clipped text (verified via dewey_search on the learning identity), proving capture-time clipping rather than diff-generation loss; distinguishing the two requires querying the index independently of the diff under review. Detection one-liner: for f in .uf/dewey/learnings/*.md; do tail -1 "$f" | grep -qE '[.!?)"]$' || echo "$f"; done. Author learning bodies to end with terminal punctuation so the check has a reliable signal.
