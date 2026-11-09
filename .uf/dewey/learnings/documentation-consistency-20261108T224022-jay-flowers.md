---
tag: documentation-consistency
author: jay-flowers
category: pattern
created_at: 2026-11-08T22:40:22Z
identity: documentation-consistency-20261108T224022-jay-flowers
tier: draft
---

Council persona-count changes ripple across six doc surfaces carrying two distinct semantics: the meta-repo instantiation count (canonical + repo-local personas, e.g. 11) and the `uf init` consumer count (canonical only, e.g. 9). PR #959 bumped QUICKSTART.md:6, docs/usage.md:35, and docs/usage.md:292 to the unqualified instantiation count while docs/architecture.md:368 and docs/heroes.md:113 kept the "(9 canonical, 2 repo-local)" qualifier — leaving consumer-framed claims overstating what `uf init` ships and docs/usage.md self-contradictory (line 50 still said 9). Pick one semantics per doc audience, sweep every grammatical position (table cells, command tables, running prose), and update the repo-local enumeration callout that reconciles the two counts.
