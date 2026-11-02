---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:57:09Z
identity: issue-triage-20261102T025709-jay-flowers
tier: draft
---

Triage verification for documentation-gap issues requires checking both halves of the claim: the change shipped AND the doc gap persists. For unbound-force issue #690 (document resolve_base_ref from change #687), repo grep confirmed the tool (.opencode/plugins/uf-workflow/index.ts), the review-council integration (uf.review-council.md:132), and the Key Files Changed demo removal (CHANGELOG fix-stale-base-ref entries), while a docs/-scoped grep for "resolve_base_ref|base ref" returned zero matches — the gap is real and persists weeks after the change merged. Survey-style doc items ("update any docs referencing X") proved to be potential no-ops: docs/ never described the old hardcoded-main base-ref behavior, the demo output shape, or Key Files Changed, so only the "document the new fallback behavior for fork users" item carries substantive work. Method worth repeating: classify each doc-issue item by whether an existing doc surface mentions the old behavior (update path) versus nothing (add path), verified with a scoped grep before validating the issue's scope.
