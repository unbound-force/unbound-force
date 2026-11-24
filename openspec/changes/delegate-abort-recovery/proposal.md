## Why

`uf_delegate_command` in command-path mode is the
orchestration primitive for `/uf.run-pipeline-plan`
and `/uf.unleash`. Four consecutive delegations on
2026-11-24 (replicator#182, unattended) failed at
report time with `MessageAbortedError` / `model_mismatch`
while the child sessions had already produced durable
work: checkpoints, branches, commits (~750 insertions),
dispatch artifacts, and progress state. Each failure was
classified `retryable:false` with no `resume` payload,
forcing full re-delegation and discarding the child's
final report text. The checkpoint design allowed forward
progress, but every cycle paid the cost of re-discovery
and lost observability.

## What Changes

This change makes report-time failures recoverable and
diagnosable. It does not remove any command or pipeline
behavior.

- **Final text survives.** A `child_message_error` or
  `model_mismatch` failure envelope carries the child's
  final text (bounded, sanitized). When the prompt
  response holds no text, the wrapper reads it from the
  child session's persisted transcript.
- **Aborts become resumable.** A command-path delegation
  that fails with `MessageAbortedError` carries the same
  `resume` payload the timeout path emits and is
  retryable when that state is readable.
- **Report diagnostics.** Those envelopes name the error,
  the elapsed time, the last tool call, the context
  usage, and the turn that ended the run (its agent, its
  model, and whether it was injected into the session
  rather than sent by the wrapper).

Root cause (investigation on issue 1199): the
opencode-agent-supervisor plugin's critique delivery
aborts and re-steers child sessions, and its injected
turn runs as agent `build` on the default model. The
supervisor fix is jflowers/opencode-supervisor#86; gh-hud
dispatch panes run the supervisor in `repair-only` until
it ships (jflowers/gh-hud#182). The issue's suggested
model threading is not needed: the executor already
passes the resolved model in the prompt body, and the
child honors it until an injected turn replaces it. The
diagnostics above make that cause visible instead.

## Capabilities

### New Capabilities
- `delegation-recovery`: abort recovery, resumable
  classification, and abort instrumentation for
  `uf_delegate_command` delegated runs.

### Modified Capabilities
- `uf-workflow`: `uf_delegate_command` report-time
  error handling and result recovery.

### Removed Capabilities
- None.

## Impact

Touches the delegation wrapper used by `/uf.unleash`
and `/uf.run-pipeline-plan`: the `.opencode/lib`
delegation executor, a new `.opencode/lib` transcript
reader, and the `uf-workflow` plugin wiring. Model
selection is unchanged. `invoke_agent` and review
dispatch keep their current classification. No schema,
branch, or CLI contract changes. No
new dependencies. Existing checkpoint and dispatch
artifact layouts are consumed as-is.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

Change strengthens artifact-based collaboration: child
outputs remain durable on disk and become usable even
when the parent channel aborts. No new synchronous
coupling is introduced.

### II. Composability First

**Assessment**: PASS

`uf_delegate_command` remains standalone. Recovery and
resume are additive behaviors inside the wrapper; no new
mandatory peer or install dependency.

### III. Observable Quality

**Assessment**: PASS

Adds machine-parseable resume payloads and structured
abort instrumentation. Existing JSON envelope gains
fields rather than breaking format; provenance is
preserved.

### IV. Testability

**Assessment**: PASS

New behavior is exercised via executor unit tests and
pipeline resume fixtures without external services.
Side effects are verified through envelope shape,
resume payload, and logged diagnostics.

### V. Security by Default

**Assessment**: PASS

No secret or privilege changes. Recovered final text
passes through the existing invocation-error sanitizer
(secrets and home paths redacted) before it reaches the
envelope.

### VI. Deterministic Execution

**Assessment**: PASS

Report-time recovery and resume payload are mechanical
procedures extracted to tested tools; no new
prose-only choreography.
