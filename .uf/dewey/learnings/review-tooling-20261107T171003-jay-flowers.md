---
tag: review-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:10:03Z
identity: review-tooling-20261107T171003-jay-flowers
tier: draft
---

Probing .uf/ paths during reviews can silently produce false 'absent' signals: glob returns 'No files found' for .uf/dewey/learnings/*.md even while those files exist on disk (hidden-directory blindness), and content grep missed the .uf/artifacts/dispatch/ directory that a direct directory read revealed. When a review conclusion depends on a .uf/ path being absent, control-test the probe first (glob a known-populated path under .uf/) and confirm with the read tool on the directory itself, which lists entries regardless of ignore rules.
