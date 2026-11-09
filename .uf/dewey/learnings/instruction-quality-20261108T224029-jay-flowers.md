---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-08T22:40:29Z
identity: instruction-quality-20261108T224029-jay-flowers
tier: draft
---

When a review-council child runs against a PR-base checkout with a manifest-delivered diff, the Linguist's scoped instruction-lint run (`--only <changed instruction files>`) exits 2 on any file that exists only at head (e.g., a newly added agent file), because the base tree lacks both the file and the head-side linter flags. The persona's degradation row correctly turns this into one LOW finding, but reviewers should also state what was consequently NOT verified (frontmatter, references, IL-042/IL-043, fence parity, token budgets) and ground those properties in the CI-declared instruction-lint check that runs at head, or in a Step 1 re-run from a head checkout — instead of silently omitting lint evidence.
