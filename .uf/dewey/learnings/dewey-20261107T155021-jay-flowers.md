---
tag: dewey
author: jay-flowers
category: gotcha
created_at: 2026-11-07T15:50:21Z
identity: dewey-20261107T155021-jay-flowers
tier: draft
---

Before committing a batch of .uf/dewey/learnings/*.md files, run a body-termination check over every file: the write path can truncate learning bodies at ~512 content characters with no truncation marker, and the uniform +10/-0 diffstat of same-shaped learning files hides the cut entirely. Observed twice in unbound-force: PR #949's 55-learning batch (3 of 7 sampled files cut mid-sentence) and PR #968's 50-learning batch (15 of 50; 14 bodies exactly 512 content chars and 1 at 511 — a mechanical cap, not editorial shortening). Detection is mechanical and cheap: flag any learning body whose length sits within a few characters of 512 or that ends mid-word without terminal punctuation. Recovery is only viable pre-merge — probe the Dewey store (learning/<identity> page) for a stored full text first; in PR #968's batch no store pages existed for the affected identities, leaving the ended review sessions as the only full-text source, so once merged the truncated file becomes the canonical record with no recovery path. Per Constitution Principle VI, the check belongs in an executable guard at the learning write/commit path, not in review-time prose.
