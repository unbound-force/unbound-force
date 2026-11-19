---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-15T05:07:35Z
identity: review-dispatch-20261115T050735-jay-flowers
tier: draft
---

The spill-validation warning-kind allowlist in .opencode/commands/uf.review-council.md (the paragraph enumerating orphaned_submissions_excluded through proposals_overflow) is hand-synced with the warning kinds the review-dispatch plugin emits; it has no single source of truth. Whenever review-dispatch gains a warning kind, the same change must add it to this allowlist (grep 'orphaned_submissions_excluded' in .opencode/commands/ to locate the paragraph), or truncation-recovery spill validation will silently discard the new kind as unknown. PR #1087 (child_empty_response) synced both sides correctly.
