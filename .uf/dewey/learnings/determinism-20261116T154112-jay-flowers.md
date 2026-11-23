---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-16T15:41:12Z
identity: determinism-20261116T154112-jay-flowers
tier: draft
as_of: "unbound-force@c2dfd4b8f15f9ae106c599aef9a0418bf075dd93"
---

In the /uf.address-feedback flow, the completion gate between post_feedback_replies and write_feedback_triage_artifact lives only in command prose (Section 4.6: a `complete: false` result ends the run; Section 4.7 runs only after `complete: true` or the skip-posting path). post_feedback_replies computes `complete` and records the reply-comments checkpoint only when complete, but write_feedback_triage_artifact's joinFiledIssues renders failed issue filings as nulls-as-data rather than failing closed — so no tool enforces the gate. General pattern: when a consuming tool deliberately renders failure states as data ("recording reality rather than an error"), the completion gate for the downstream call must move into that tool's fail-closed validation (e.g. invalid_input when the replies projection is incomplete or a FILE item carries `reply: "issue_failed"`), or the gate remains prose routing — DE-005 residue that cannot be seam-tested and drifts after context compression.
