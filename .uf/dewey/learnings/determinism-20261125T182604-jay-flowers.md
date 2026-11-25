---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-25T18:26:04Z
identity: determinism-20261125T182604-jay-flowers
tier: draft
as_of: "unbound-force@59d260231240f979d60ee4fc3d451fa5e5c39723"
---

Failure-envelope classification is a DE-006 seam best computed inside the delegation tool from structured state: uf_delegate_command's report-time recovery (openspec/changes/delegate-abort-recovery, PR #1206) marks a command-path MessageAbortedError retryable exactly when the delegate resume state is readable, routes error codes through a REPORT_RECOVERY_CODES constant with pure branching, and exposes structured payload fields (final_text, final_text_source, report_diagnostics, resume) so the caller's only remaining work is the judgment of whether to re-delegate. Verify such envelopes with output-shape tests pinning both field values and negative shape (not.toHaveProperty on paths that must stay clean, e.g. timeout failures carry no recovery fields), plus a round-trip test that writes progress state and asserts the full resume payload.
