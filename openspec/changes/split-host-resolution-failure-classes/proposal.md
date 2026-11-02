## Why

Follow-up to `host-fallback-provenance` (Fixes #901), recorded in its
design D2 ("Advisor-symmetry conflation") and filed as #914:
`resolveHostModel()` failures cover two causes with opposite
remedies, but both dispatch branches that call the function classify
them identically.

`StructuralHostModelError` (`.opencode/lib/agent-executor.ts`) marks
causes that retrying cannot fix — the parent message is not an
assistant message, carries no model identity, or carries an invalid
variant — and carries a `remediation` string naming how to fix the
cause. Transient failures (API errors reading the parent message)
remain plain `Error`s by design. The shared dispatch core
`executeDispatchRun` (`.opencode/plugins/review-dispatch/index.ts`)
catches both kinds from either caller — the `source === "host"`
branch (RD-FR-037) and the advisor fallback (#693) — and maps every
failure to one class: error code `host_model_unavailable`,
`retryable: true`, message `sanitizeInvocationError(errorText(error))`.

Three defects follow:

1. A structural failure is reported retryable, inviting callers and
   recovery operators to re-invoke a dispatch that cannot succeed
   until the underlying cause (for example a session message carrying
   no model identity) is fixed.
2. `errorText()` returns only `err.message`, so the remediation string
   the shared failure carries is dropped from every recorded form of
   the error — the `dispatch_agent_run` result, the
   `redispatch_failed_runs` early-exit result entries that repeat
   `outcome.result.error`, and any run error a failure path records.
3. No consumer can distinguish the two classes without parsing
   message prose.

A fourth gap is specification: the spec record never states what
`retryable` means for a pre-execution early exit. No run file exists,
so `redispatch_failed_runs` (DR-FR-002) has nothing to target, and
`consolidate_dispatch` fails closed on both branches with "included
plan entry `<agent>/<sequence>` has no run" → INCONCLUSIVE.
`retryable: true` therefore denotes caller re-invocation only — a
fact operators must currently infer.

The classification precedent already exists in this repo:
`uf_delegate_command` (`.opencode/plugins/uf-workflow/index.ts`)
checks `err instanceof StructuralHostModelError` and surfaces a
non-retryable failure carrying
`<message>. Remediation: <remediation>; retrying will not fix this.`
This change brings the dispatch core to the same semantics.

Fixes #914 (lineage: parent #901, advisor precedent #693).

## What Changes

Split the host-resolution failure classification in the shared
dispatch core (`executeDispatchRun`,
`.opencode/plugins/review-dispatch/index.ts`) across BOTH branches
that call `resolveHostModel()` — the host-source branch (RD-FR-037)
and the advisor fallback (#693):

- Structural failures (`error instanceof StructuralHostModelError`)
  fail the run with a new error code `host_model_structural`,
  `retryable: false`, and a message that preserves the shared
  failure's remediation string, rendered in the established
  `uf_delegate_command` shape (`<message>. Remediation:
  <remediation>; retrying will not fix this.`) and passed through
  `sanitizeInvocationError()`.
- Transient failures keep today's class exactly:
  `host_model_unavailable`, `retryable: true`, sanitized message.
- Abort-signal cancellation keeps precedence over classification and
  keeps the `cancelled` code.
- Both branches classify through one shared helper, so the
  advisor/host symmetry the parent change established holds by
  construction rather than by mirrored edits.

Both failure paths remain pre-execution early exits: no run file is
written. The spec record states explicitly that `retryable` on these
early exits means caller re-invocation only — no run file exists for
`redispatch_failed_runs` (DR-FR-002) to target, and a session whose
included plan entry early-exited consolidates fail-closed ("included
plan entry has no run" → INCONCLUSIVE).

No planner, finalizer, input-schema, or run-file shape changes. All
plugin edits mirror byte-identical into the scaffold asset copy.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `review-dispatch`: host-resolution failure classification splits
  into structural (new non-retryable, remediation-carrying
  `host_model_structural`) versus transient (unchanged retryable
  `host_model_unavailable`) across both the host-source branch
  (RD-FR-037, modified) and the advisor fallback (modified). New
  requirement RD-FR-038 governs the classification, the remediation
  preservation, and the caller-re-invocation-only meaning of
  `retryable` on pre-execution early exits.

### Removed Capabilities
- None.

## Impact

- **Live plugin**: `.opencode/plugins/review-dispatch/index.ts` —
  the two `resolveHostModel()` catch blocks in `executeDispatchRun`
  (host branch and advisor fallback), one module-level
  classification helper, the `StructuralHostModelError` import, and
  the function's doc comment. No schema, planner, prompt-builder, or
  finalizer changes.
- **Scaffold asset**:
  `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`
  — byte-identical mirror of the live plugin (Go drift gate
  `TestEmbeddedAssets_MatchSource`,
  `internal/scaffold/scaffold_test.go`).
- **Tests**: `.opencode/test/dispatch-agent-run.test.ts`
  (structural/transient classification for both branches —
  non-assistant message, missing model identity, invalid variant
  versus a transient lookup failure), `.opencode/test/dispatch-recovery.test.ts`
  (redispatch early-exit result entries carry the classified error).
- **Behavior change**: structural host-resolution failures surface
  `host_model_structural` / `retryable: false` with the remediation
  string, instead of `host_model_unavailable` / `retryable: true`.
  Verdict projection is unaffected in practice (pre-execution exits
  write no run file); if a structural code ever appears in a run
  error, `availabilityError()` correctly does not read it as an
  availability blocker — a structural cause is a configuration
  problem, not an availability problem.
- **Out of scope** (recorded for follow-up, not changed here):
  `invoke_agent`'s identical conflation
  (`.opencode/plugins/invoke-agent/index.ts`) and
  `uf_delegate_command`'s uppercase-code rendering keep their current
  shapes.
- **No external API changes**: tool input schemas are unchanged; the
  new error code appears only inside the existing
  `error: { code, message, retryable }` result shape.
- **CI replication**: `make plugin-test` (vitest suite with the 94
  percent statements coverage floor) and the Go scaffold drift gate
  (`go test -race -count=1 ./internal/scaffold/...`).
- **Documentation impact**: CHANGELOG.md entry; the
  docs/configuration.md "Host-fallback model resolution" note (which
  currently documents only the retryable `host_model_unavailable`
  failure class) gains the structural class and its non-retryable
  semantics. The change is user-facing (operators and calling agents
  consume the error class), so the org constitution's Cross-Repo
  Documentation rule applies: an in-repo documentation issue and a
  unbound-force/website documentation issue are filed before PR merge
  (parent-change precedent: #778 ↔ website#316, #891 ↔ website#360).

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0 — six
active principles).

### I. Autonomous Collaboration

**Assessment**: PASS

The failure artifact becomes self-describing: the error class
(`host_model_structural` versus `host_model_unavailable`), the
`retryable` flag, and the remediation string travel together in the
tool result, so a calling agent or recovery operator can decide
without consulting another hero or re-deriving the cause from
message prose. No synchronous coupling is introduced.

### II. Composability First

**Assessment**: PASS

The change consumes the classification primitive the shared library
already exports (`StructuralHostModelError` with its `remediation`
field) — the same primitive `uf_delegate_command` consumes — and adds
no dependency, no new interface, and no cross-hero requirement. Each
hero repo remains independently installable.

### III. Observable Quality

**Assessment**: PASS

The class distinction becomes machine-parseable: consumers branch on
`error.code` and `error.retryable` instead of parsing message text,
and the remediation string is preserved in the recorded error text so
the failure's cause and fix remain traceable. The retryable-semantics
statement moves from operator folklore into the spec record.

### IV. Testability

**Assessment**: PASS

Both branches are testable in isolation through the existing vitest
doubles: the fake client's `session.message` shapes each structural
cause (non-assistant role, missing identity, invalid variant) and the
transient cause (lookup failure) — the exact parameterization the
`invoke_agent` suite already uses. No external services or network.

### V. Security by Default

**Assessment**: PASS

The remediation-carrying message passes through the shared
`sanitizeInvocationError()` redaction (Bearer tokens,
credential-shaped assignments, URL query strings, home paths) with
its bounded length; no new inputs, no new external surface, and no
credentials enter tool results.

### VI. Deterministic Execution

**Assessment**: PASS

The change is runtime dispatch code — one classification helper and
two catch blocks, byte-identical between the live plugin and the
scaffold asset, pinned by the vitest suite and the Go drift gate. The
meaning of `retryable` on a pre-execution early exit moves from
inference into tested tool code and spec text; no instruction prose
is added or altered.
