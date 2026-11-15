## Why

A `/uf.unleash` child delegated by `/uf.run-pipeline-plan`
(orchestrator `ses_e258662fdffeo4EoTjIGt5NEvw`, child
`ses_e254aedb8ffeqlLkdOgSpKNj60` on 2026-11-14 15:36-17:29
UTC) was sent to
`enmaas/rits/meta-models/Muse-Spark-Open-816B-A42B-NVFP4`
via classifier `classifier-llm`/`spec-authoring`. That
model reports `context: 131072` and `output: 32768`
(placeholders, noted as unpublished) leaving ~98k usable
input.

After opencode compaction every turn starts with ~76.5k
tokens of cached prefix (system prompt + `/uf.unleash`
command body), leaving ~20k for work. One or two file
reads overflow it, so it compacts again — about 150
compactions in 110 minutes (~every 30s, 75 min loop), tool
counts `read 1099 / bash 99 / edit 17`, message finishes
`165 stop / 450 tool-calls`, 18-min stall on
`dispatch_review_runs` 15:58->16:16.

The executor's progress fingerprint treated each
compaction as progress, so the idle timeout from #950 never
fired. The child finally returned `status: success` with a
compaction summary as final text and
`checkpoint_recorded: false`, which the orchestrator could
not distinguish from a real gate pause (companion
gate-integrity track).

## What Changes

- **Model selection headroom gate:** Delegation MUST
  reject or mark down candidate models whose usable window
  (`limit.context - limit.output`) cannot fit the
  command's base footprint plus a working margin. The
  rejection/penalty reason MUST be recorded in
  `selection_fallback_reason` so the audit trail explains
  why the classifier-preferred model was not used.
- **Compaction-storm detection:** The executor MUST detect
  N compactions with no non-read tool progress in between
  and fail fast with a distinct retryable failure code
  (e.g. `context_thrash`) and a `resume` envelope carrying
  durable state, instead of treating compactions as
  progress.
- **Final-text success guard:** A child whose final text
  is a compaction summary or synthetic `Continue if you
  have next steps...` prompt MUST NOT be reported as
  `success`. It MUST be reclassified as a retryable failure
  (or `context_thrash` where applicable) with appropriate
  error code.
- **Tests and observability:** Unit/integration tests for
  each of the three behaviors; provenance fields
  (`selection_fallback_reason`, timeout details) remain
  machine-parseable for orchestrator resumption.

## Capabilities

### New Capabilities

- `delegation`: New capability owning the delegation-layer
  robustness requirements — selection headroom gate
  (DELEG-FR-001), compaction-storm detection
  (DELEG-FR-002), and final-text success guard
  (DELEG-FR-003).

### Modified Capabilities

- `delegation-timeouts`: DT-FR-002 (Idle Timeout) progress
  accounting is refined so compaction events and synthetic
  continuation prompts no longer count as probe progress;
  probe state exposes the compaction-since-progress count
  shared with the storm detector.

The `model-matrix` capability is unchanged: the headroom
gate runs after the existing precedence chain (MM-FR-003)
and records through the existing audit-trail surface
(MM-FR-007), with the footprint supplied by the caller so
selection stays pure.

### Removed Capabilities

- (none)

## Impact

- **Affected systems/files:** `.opencode/lib/model-matrix.ts`
  (headroom gate), `.opencode/lib/delegation-budget.ts`
  (fingerprint / probe), `.opencode/lib/agent-executor.ts`
  (idle watchdog + final-text guard),
  `.opencode/plugins/uf-workflow/index.ts`
  (`uf_delegate_command` `SelectionResult` /
  `InvokeAgentResult` surface, footprint caller,
  docs/configuration timeout docs).
- **Behavior:** Tight-context models will no longer be
  selected for large-prompt commands; storming children
  terminate in minutes not hours; summaries no longer
  masquerade as success.
- **Compatibility:** Failure codes are additive
  (`context_thrash`); `resume` envelope reuses existing
  `DelegateResumeState` / `InvokeAgentResult.timeout`
  patterns. No schema break; orchestrator fallback is
  retry/resume.
- **Rollout:** Threshold `N` and margin are configurable
  with validated, fail-closed defaults; existing tests for
  #930/#950 idle budgets remain green.

## Constitution Alignment

Assessed against the Unbound Force org constitution v1.4.0
(six principles).

### I. Autonomous Collaboration

**Assessment**: PASS

Change preserves artifact-based collaboration. Selection
decisions remain recorded in `selection_*` audit fields
and `selection_fallback_reason`; the new `context_thrash`
failure and `resume` envelope are artifacts the
orchestrator consumes without synchronous coupling. No new
inter-hero RPC.

### II. Composability First

**Assessment**: PASS

No new mandatory dependencies. Matrix, executor, and
workflow plugin remain independently installable; headroom
gate and storm detector are local to delegation and do not
require other heroes. Fallback to host model remains.

### III. Observable Quality

**Assessment**: PASS

Enhances observability: `selection_fallback_reason`
explains rejected models; `context_thrash` error code +
`timeout` details + `resume` envelope give
machine-parseable provenance for stalls and continuation.
Final-text guard prevents false `success` provenance.

### IV. Testability

**Assessment**: PASS

All new behavior is testable in isolation via injected
`client.session.children/messages`, `now` clock seam, and
matrix fixtures — no external services. Tasks include unit
tests for headroom math, the footprint estimator's guard
paths (path-escape rejection and the fail-closed
read-failure default), storm counting, and final-text
classification, plus integration probe and seam tests.

### V. Security by Default

**Assessment**: PASS

No new secrets, privileged containers, or dependencies.
Input validation (model limits, compaction counters, knob
ranges) stays fail-closed; error messages are sanitized
via `sanitizeInvocationError`.

### VI. Deterministic Execution

**Assessment**: PASS

Compaction counting, fingerprinting, and final-text
classification are mechanical procedures that MUST live in
executable tools (`delegation-budget.ts`,
`agent-executor.ts`) with tests, not prose. Prescribed
failure envelopes are rendered by tools, not re-reasoned
from prompts. Change includes measurable deltas
(compaction window, margin) for efficiency claims.
