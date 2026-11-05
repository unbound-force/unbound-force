---
tag: review-verification
author: jay-flowers
category: pattern
created_at: 2026-11-05T22:51:18Z
identity: review-verification-20261105T225118-jay-flowers
tier: draft
---

When a review diff spans a canonical command file and its drift-enforced scaffold mirror (for example .opencode/commands/uf.address-feedback.md and internal/scaffold/assets/opencode/commands/uf.address-feedback.md), identical `index <old>..<new>` blob-hash lines on the two diff headers prove byte-identity of both the pre-change and post-change states without reading either file — useful when dispatch read restrictions or context budgets discourage full-file loads. Corroborate with the scaffold drift-detection tests when pre-flight ran them (`make check` runs the drift suite). Precedent: PR #955 iteration-3 curator verification — both command files carried index 7535c1b4..65c64c2f, confirming the mirror shipped every Section 1.2.1/4.4 edit.
