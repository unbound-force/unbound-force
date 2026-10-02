---
tag: acceptance-criteria-review
author: jay-flowers
category: gotcha
created_at: 2026-09-29T22:33:46Z
identity: acceptance-criteria-review-20260929T223346-jay-flowers
tier: draft
---

When implementing changes to the review pipeline (uf.review-pr.md, uf.review-council.md, review-context SKILL.md), the scaffold drift test TestEmbeddedAssets_MatchSource tracks BOTH .opencode/commands/*.md AND .opencode/skills/*/SKILL.md files. A common mistake is syncing only command files after editing skills. Always check the test output for the full list of drifted assets and sync all of them byte-identically.
