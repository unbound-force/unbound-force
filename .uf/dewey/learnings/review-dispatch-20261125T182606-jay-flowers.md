---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-25T18:26:06Z
identity: review-dispatch-20261125T182606-jay-flowers
tier: draft
as_of: "unbound-force@59d260231240f979d60ee4fc3d451fa5e5c39723"
---

When a review-council dispatch runs from a base checkout (working tree == merge base, PR diff delivered via manifest chunks), the Determinist's Step 1 tool evidence comes back empty by construction: instruction-lint --base main --changed reports delta.files: [] and tool-inventory --seams reports [] even though the PR changes tool code. Recognize the condition (clean tree at the base SHA, matching the pre-flight note) and apply the graceful-degradation fallback for DE-007: verify seam layers by reading whether the change adds or edits tests referencing each changed tool — registration-surface.test.ts and uf-workflow.test.ts for registration, and the tool's behavioral test file for output shape and round-trips (tool-inventory.test.ts pins unleash-delegate.test.ts as uf_delegate_command's output-shape + round-trip layer home). DE-004 is unaffected: it keys off delta.files, which correctly reports no instruction files changed.
