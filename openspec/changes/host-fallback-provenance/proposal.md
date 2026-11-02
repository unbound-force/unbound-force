## Why

`consolidate_dispatch` with provenance — the finalization step —
deterministically fails validation whenever a dispatch panel
includes a host-fallback run: an agent planned with
`source: "host"` because it appears in neither `runs[<mode>]` nor
`advisor[<mode>]` in `.uf/review-matrix.yaml`. The dispatcher
records advisor-shaped model provenance for host runs while the
finalizer enforces host-specific provenance invariants, so every
such run is rejected with four validation defects. The dispatch
itself succeeds — only finalization fails — forcing fail-closed
INCONCLUSIVE outcomes despite successful native verdicts.

Verified against this repo's
`.opencode/plugins/review-dispatch/index.ts` (line numbers below
are this repo's copy, not the gaze scaffold copy from the issue):

1. `sourceFor()` (L1426) classifies an agent in neither
   `matrix.runs[mode]` nor `matrix.advisor[mode]` as
   `source: "host"`.
2. `baseRuns()` (L1536-1547) emits host-fallback entries with
   `tier: null, model: null, variant: null` and reason "no
   explicit or advisor run; use the active host model".
3. `resolvePlanEntryRuntime()` (L3063) normalizes those nulls to
   `undefined`, so `dispatch_review_runs` (L5762) and
   `redispatch_failed_runs` (L7350) pass the shared core neither
   model nor tier.
4. `executeDispatchRun` (L3226) silently defaults
   `tier = options.tier ?? "standard"` and resolves that profile's
   model via `profilePair` — a per-developer
   `.uf/review-matrix.override.yaml` MAY set one. When the
   resolved model is null, ONLY `source === "advisor"` calls
   `resolveHostModel()` recording `requested_model: null` plus
   `resolved_parent_model`/`variant` (#693); there is NO branch
   for `source === "host"`, which instead fails early with
   `tier_model_unavailable`.
5. `validateRunModel()` (L2316-2360) rejects the persisted host
   run with four defects: "host run must retain null requested
   model provenance"; "model provenance does not match its plan
   entry" (plan model null vs run requested_model set);
   model_mismatch expectation (expected model comes from a null
   `resolved_parent_model`); "successful host run requires
   resolved parent model".

Real-world impact: bulk `/uf.triage-issue` over 12 issues in
unbound-force/gaze (2026-11-01). gaze registered
`divisor-gaze-testing` in `advisor.code/spec/feedback` but not
`advisor.triage`, so 9 of 12 triage dispatches planned it as
include + host-fallback. All native verdicts were VALID, but each
panel failed finalization three times and reported fail-closed
INCONCLUSIVE. The 3 dispatches whose planners scope-skipped that
agent finalized cleanly. Existing recovery does not apply:
`redispatch_failed_runs` (DR-FR-002) only re-runs failed runs,
and host-fallback runs succeed — the rejection happens only at
finalization. The workaround (listing every capability-relevant
persona in each mode's advisor list) is fragile: swapping an
agent out of an advisor list merely relocates the exposure.

The finalizer is not the defect. RD-FR-008 as modified by the
`accept-variantless-parent-model` change requires host-source
runs to carry null requested provenance and a non-null
`resolved_parent_model` after successful host lookup. The
dispatcher violates the artifact spec; the fix brings it into
compliance.

Fixes #901.

## What Changes

Handle `source === "host"` symmetrically with the advisor branch
(#693) in the dispatch core's model resolution: a host-source run
with no direct model bypass MUST resolve the parent session's
active model via the shared `resolveHostModel()` and record
host-run provenance — `requested_model`, `provider`, `model_id`,
and `variant` null, `resolved_parent_model` non-null after
successful lookup, `resolved_parent_variant` variant-or-null —
populating the existing `hostResolvedParent` mechanism that
already produces exactly this shape. The silent
`tier ?? "standard"` default becomes unreachable for host-source
entries (it remains in force for direct `dispatch_agent_run`
invocations that supply neither model nor tier and default to
`source: "explicit"`).

No finalizer change: `validateRunModel`'s host invariants already
match RD-FR-008. No planner change: `baseRuns` already records
the correct host-fallback intent. Host resolution failure fails
the run with retryable `host_model_unavailable` (mirroring the
advisor branch) instead of the current `tier_model_unavailable`
early exit. All changes mirror byte-identical into the scaffold
asset copy.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `review-dispatch`: The shared dispatch core
  (`executeDispatchRun`) gains a host-source model-resolution
  branch: host-source entries resolve the active host model via
  `resolveHostModel()` regardless of profile configuration, and
  persist host-run provenance per RD-FR-008. New requirement
  RD-FR-037 governs the branch.

### Removed Capabilities
- None.

## Impact

- **Live plugin**: `.opencode/plugins/review-dispatch/index.ts`
  — `executeDispatchRun` model-resolution section only; no
  schema, planner, prompt-builder, or finalizer changes.
- **Scaffold asset**:
  `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`
  — byte-identical mirror of the live plugin (Go drift gate
  `TestEmbeddedAssets_MatchSource`,
  `internal/scaffold/scaffold_test.go`).
- **Tests**: `.opencode/test/dispatch-agent-run.test.ts`
  (host-source resolution block mirroring the advisor
  host-model-fallback block),
  `.opencode/test/dispatch-review-runs.test.ts`
  (host-fallback panel dispatch finalizes cleanly),
  `.opencode/test/dispatch-recovery.test.ts` (re-dispatch of a
  failed host run reuses host resolution),
  `.opencode/test/consolidate-dispatch.test.ts` /
  `.opencode/test/finalize-review-dispatch.test.ts`
  (provenance-carrying finalization over a panel including a host
  run).
- **Behavior change**: host-fallback runs execute on the parent
  session model even when a profile model is configured, and
  host-source runs without a profile model no longer fail with
  `tier_model_unavailable` (they resolve the host model or fail
  retryable `host_model_unavailable`). Both outcomes align the
  dispatcher with the planner's recorded intent and RD-FR-008.
- **No external API changes**: tool input schemas are unchanged.
- **CI replication**: `make plugin-test` (vitest suite with the
  94 percent statements coverage floor) and the Go scaffold
  drift gate (`go test -race -count=1 ./internal/scaffold/...`).
- **Documentation impact**: CHANGELOG.md entry; no AGENTS.md or
  README structural changes. docs/configuration.md's Review
  Council Configuration section gains a host-fallback semantics
  note (parent-session model execution for host-source runs,
  profile/override bypass, advisor lists as the sanctioned
  pinning mechanism) — tasked as tasks.md 3.4 with the
  documentation issue filed against this repo per the
  documentation-gate precedent (#778, #891, #676). The org
  constitution's Cross-Repo Documentation rule applies (the
  change is user-facing): a unbound-force/website documentation
  issue is filed before PR merge (tasks.md 3.5), pairing with
  the in-repo issue per precedent (#778 ↔ website#316, #891 ↔
  website#360, model-matrix-scaffold-deployment's website#362).

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0 —
six active principles).

### I. Autonomous Collaboration

**Assessment**: PASS

The change restores artifact-based collaboration: dispatch panels
that include host-fallback runs can again finalize the
review-dispatch artifact that downstream consumers
(`/uf.review-council`, `/uf.triage-issue`, `/uf.finale`) read.
Provenance remains self-describing — the artifact distinguishes
"ran on the active host model" (null requested, resolved parent
recorded) from "ran on a configured profile model" (requested
recorded). No synchronous coupling is introduced; host-model
resolution already exists in the shared agent-executor library.

### II. Composability First

**Assessment**: PASS

The fix reuses the existing shared `resolveHostModel()` from
`.opencode/lib/agent-executor.ts` — the same function
`invoke_agent` uses for host-source entries (RD-FR-006) and the
advisor branch uses (#693). No new dependency, no new interface;
one branch in one function converges three execution paths on one
resolution mechanism. Each hero repo remains independently
installable.

### III. Observable Quality

**Assessment**: PASS

Host-run provenance becomes accurate and machine-parseable:
today's host-fallback run either carries advisor-shaped
provenance (four validation defects at finalization) or fails
`tier_model_unavailable`. After the fix, the persisted run
carries the RD-FR-008 host shape, `model_mismatch` detection
compares the child's reported model against the resolved parent
model, and finalization succeeds with a verdict instead of
INCONCLUSIVE. The bug's observable signature (successful native
verdicts bricked by finalization) disappears.

### IV. Testability

**Assessment**: PASS

The host branch is testable in isolation through the existing
vitest doubles: `makeDeps({ readText })` controls matrix
profiles, and the fake client's `session.message` controls host
resolution (success, failure, abort) — the exact pattern the
advisor host-model-fallback tests use. The persisted run-file
provenance and the finalization round-trip are assertable without
external services; `make plugin-test` enforces the coverage
floor.

### V. Security by Default

**Assessment**: PASS

The change adds no dependency and no new external surface: host
resolution reuses the shared `resolveHostModel()` client lookup
the advisor branch already exercises, and its failure mode is
the sanitized, retryable `host_model_unavailable` error — the
message passes through the shared `sanitizeInvocationError()`
redaction (Bearer tokens, credential-shaped assignments, URL
query strings, home paths) rather than surfacing raw provider
or runtime text. No credentials, tokens, or filesystem paths
enter the persisted run file.

### VI. Deterministic Execution

**Assessment**: PASS

The change is runtime dispatch code: one branch in
`executeDispatchRun`'s model-resolution section, byte-identical
between the live plugin and the scaffold asset, pinned by the
vitest suite and the Go drift gate. No instruction prose,
command markdown, agent definition, or skill text is added or
altered — the model-resolution procedure moves out of
accidental profile inheritance into tested tool code, which is
the principle's exact intent.
