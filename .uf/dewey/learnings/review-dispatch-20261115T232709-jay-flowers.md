---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-15T23:27:09Z
identity: review-dispatch-20261115T232709-jay-flowers
tier: draft
as_of: "unbound-force@505dedf886fceeb2e552b7c1cf7633e2b6c36310"
---

instruction-lint --base <sha> --changed reports an empty delta.files (clean summary, zero findings) when the review session's working tree sits on main at the PR base instead of the PR head — the linter is the mechanism of record but can only lint what is on disk. Manifest-delivered review dispatches to a base-state checkout therefore produce no head-state lint evidence, and the reviewer must not hand-derive the linter's checks to compensate (DE-001). Check delta.files before treating a clean linter run as evidence about the change: an empty delta means the tool never saw the changed prose; report the scope caveat in the review summary and rely on the implementer's recorded lint run for head-state claims.
