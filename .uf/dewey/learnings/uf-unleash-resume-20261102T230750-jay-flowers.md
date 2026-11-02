---
tag: uf-unleash-resume
author: jay-flowers
category: gotcha
created_at: 2026-11-02T23:07:50Z
identity: uf-unleash-resume-20261102T230750-jay-flowers
tier: draft
---

When a spec-review gate closes with a human-adjudicated approving verdict plus consented post-dispatch spec fixes (the per-finding "Apply suggested fix" choice), record the adjudication in the checkpoint data (a gate_disposition field on the spec-review checkpoint) — because uf_unleash_detect_resume's artifact-freshness probe will report the spec-review step incomplete on the next resume: the consented fixes modified spec.md/tasks.md after the approving dispatch artifact, and the probe reconciles review completion by "most recent current-branch artifact with an approving verdict AND clean spec artifacts since that artifact". The recorded human adjudication in durable get_progress state governs over the mechanical probe result, so a resumed run honors "no further spec-review dispatch per adjudication" instead of re-dispatching a review the human already closed. Same applies after a delegation timeout kills the orchestrator late in the pipeline: re-anchor via get_progress (authoritative completed_steps and gate_disposition) rather than re-deriving gate state from artifact freshness alone. Learned resuming opsx/fix-uf-finale-ci-poll-gh-schema on 2026-11-02 after a delegation timeout killed the prior orchestrator between code-review round-2 consolidation and the demo step.
