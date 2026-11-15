---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-15T06:08:55Z
identity: instruction-quality-20261115T060855-jay-flowers
tier: draft
as_of: "unbound-force@e5e077b35f0c7984517024431e696737e330b080"
---

When a command's control flow gains a new lynchpin tool call, the command prose must state the call's failure outcome even when the tool is internally fail-safe. The executing agent only sees the file-wide convention ("on failure, emit the envelope's error and STOP"); a branch specified solely on the success shape ("Branch on the tool's gate_paused output only") leaves a failure envelope to improvised interpretation — and improvisation on a gate check is the bypass the tool was added to prevent. Observed in harden-spec-review-gate: relay_spec_review_gate calls added to uf.unleash.md and uf.run-pipeline-plan.md with true/false branches but no failure clause, unlike every other tool call in both files.
