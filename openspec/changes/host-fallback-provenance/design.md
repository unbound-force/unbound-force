## Context

Issue #901: every dispatch panel containing a host-fallback run
(an agent planned with `source: "host"` because it appears in
neither `runs[<mode>]` nor `advisor[<mode>]` of
`.uf/review-matrix.yaml`) fails `consolidate_dispatch`
finalization with four provenance defects on that run, despite
the run itself succeeding. The root cause, verified against this
repo's `.opencode/plugins/review-dispatch/index.ts`:

- `sourceFor()` (L1426) classifies such agents as `"host"`;
  `baseRuns()` (L1536) plans them with `tier: null, model: null`
  and reason "no explicit or advisor run; use the active host
  model".
- `resolvePlanEntryRuntime()` (L3063) maps those nulls to
  `undefined`, so both `dispatch_review_runs` (L5762) and
  `redispatch_failed_runs` (L7350) hand the shared core
  `executeDispatchRun` neither model nor tier.
- `executeDispatchRun` (L3226) silently defaults
  `tier = options.tier ?? "standard"` and resolves that profile's
  model. Two divergent defects follow:
  - Profile model configured (per-developer
    `.uf/review-matrix.override.yaml` — the tracked matrix ships
    empty profiles): the host run executes on the profile model
    and persists advisor-shaped provenance (`requested_model`
    set, `provider`/`model_id`/`variant` set,
    `resolved_parent_model` null).
  - No profile model: only `source === "advisor"` reaches
    `resolveHostModel()` (#693); host falls to a
    `tier_model_unavailable` early exit.
- `validateRunModel()` (L2316) enforces RD-FR-008's host
  contract — null requested provenance, non-null
  `resolved_parent_model` on success — and rejects the
  advisor-shaped host run with four defects.

The finalizer matches the spec (RD-FR-008 as modified by
`accept-variantless-parent-model`, plus its ADDED "Finalization
Model Validation" requirement). The dispatcher is the defect: it
never implements host resolution for host-source runs.

Constitution alignment carries over from the proposal (all six
active principles of the org constitution v1.4.0): the fix
restores artifact finalization (I), reuses the shared
`resolveHostModel()` (II), makes host-run provenance accurate and
finalizable (III), is testable with the existing vitest doubles
(IV), fails closed with a sanitized retryable error instead of
surfacing raw runtime text (V), and is runtime dispatch code
with no instruction-prose changes (VI).

## Goals / Non-Goals

### Goals

- Host-source runs execute on the parent session's active model
  and persist RD-FR-008 host-run provenance, so panels with
  host-fallback entries finalize with their native verdicts.
- Host resolution for host entries on both the dispatch and
  re-dispatch paths (shared core, single branch).
- Host resolution failure surfaces as retryable
  `host_model_unavailable`, symmetric with the advisor branch.
- Byte-identical scaffold mirror; test coverage per the 94
  percent floor; CI replication via `make plugin-test` and the Go
  scaffold drift gate.

### Non-Goals

- No finalizer (`validateRunModel`) change — it already encodes
  the spec.
- No planner (`sourceFor`/`baseRuns`) change — the plan already
  records the correct intent.
- No review-matrix schema or advisor-list changes — the
  fragile-workaround problem is solved at the execution seam, not
  by mandating advisor membership.
- No change to the `tier ?? "standard"` default for non-host
  sources: direct `dispatch_agent_run` invocations without model
  or tier (source defaulting to `explicit`) keep today's
  behavior; the `tier_model_unavailable` regression test
  (dispatch-agent-run.test.ts, "tier with no model configured")
  pins it.
- No timeout-derivation change: `deriveRunTimeoutMs` already
  reads the plan entry's tier as null for host entries (zero
  per-tier profile term), independent of the model-resolution
  default.

## Decisions

### D1. Fix direction: symmetric host branch in the dispatch
core (chosen)

Add a `source === "host"` branch to `executeDispatchRun`'s
model-resolution section, placed after the direct-model bypass
and before tier-based resolution:

```
if (options.model !== undefined) {
  // direct bypass (unchanged)
} else if (source === "host") {
  // resolveHostModel(); requestedModelLabel = null;
  // hostResolvedParent = { model, variant }
  // errors: cancelled / host_model_unavailable (early exits)
} else {
  // tier-based resolution (unchanged; advisor fallback intact)
}
```

The branch populates the existing `hostResolvedParent` mechanism
(L3217, L3239-3242, L3275-3278, L3344-3349), which is
source-agnostic and already produces exactly the host provenance
shape the finalizer requires: null requested/provider/model_id/
variant, `resolved_parent_model`/`resolved_parent_variant` from
the host resolution. The advisor path (#693) proves this
mechanism end to end, including the `model_mismatch` comparison
of the child's reported model against the resolved parent model.

Rationale: the planner's host-fallback entry carries no model
intent ("use the active host model"); any profile consultation
fabricates intent the plan never recorded. Host-source resolution
must therefore bypass matrix profiles entirely, making the
silent `tier ?? "standard"` default unreachable for host entries
— the issue's second fix direction, resolved structurally rather
than by weakening the default for other callers.

Alternatives rejected:

- **B: extend the `resolved.model === null` branch to host**
  (`source === "advisor" || source === "host"`). Keeps the
  silent tier default: with a profile model configured, host runs
  would still execute on it with advisor-shaped provenance — the
  exact four-defect failure this change exists to remove.
- **C: relax `validateRunModel`'s host invariants to accept the
  dispatcher's shape.** Contradicts RD-FR-008's current text
  ("For host source, requested model, provider, model id, and
  variant MUST be null"), erases the artifact's distinction
  between host-model and profile-model execution, and weakens a
  provenance gate — the gatekeeping rule forbids weakening gates
  to unblock work.

### D2. Failure semantics: retryable `host_model_unavailable`

Host resolution failure for a host-source run fails with
`host_model_unavailable` and `retryable: true` — the advisor
branch's code — rather than today's non-retryable
`tier_model_unavailable`. The error is more accurate (the
unavailable resource is the parent session's model, not a matrix
profile), and retryability matches the advisor precedent and the
transient nature of session-message lookups. Both paths remain
pre-execution early exits: no run file is written; finalization
fails closed with "included plan entry has no run" → INCONCLUSIVE,
unchanged from today's early-exit behavior.

Advisor-symmetry conflation (recorded for a follow-up):
`resolveHostModel()` failures also cover structural causes
(`StructuralHostModelError` — for example a session message
carrying no model identity) that retry cannot fix, and the
persisted `errorText` drops the remediation string the shared
failure carries. This change deliberately mirrors the advisor
branch's single `host_model_unavailable` / `retryable: true`
class rather than diverging from it; splitting structural vs
transient host-resolution failure classes across both branches
is deferred to a follow-up change (issue filed per tasks.md
3.6, scoped to also state that `retryable` on a pre-execution
early exit means caller re-invocation only — no run file
exists for DR-FR-002 redispatch to target).

### D3. Direct bypass precedence preserved

`source: "host"` combined with an explicit `model` remains a
direct bypass (the `options.model` branch precedes source-based
resolution). The persisted run then carries non-host provenance
and finalization applies its existing host invariants —
fail-closed handling of contradictory provenance input, unchanged.
The planned dispatch paths cannot produce this combination (host
entries carry `model: null`), so no additional input validation is
added; the spec records the precedence.

### D4. Scaffold mirror and drift gate

The live plugin edit is mirrored byte-identical into
`internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`.
`TestEmbeddedAssets_MatchSource` (internal/scaffold/scaffold_test.go)
is the enforcement; no scaffold-side code changes beyond the copy.

### D5. Test placement and pattern reuse

Tests extend the existing suites, mirroring the advisor
host-model-fallback describe block (dispatch-agent-run.test.ts
L801-880) which already demonstrates the required doubles:
`makeDeps({ readText })` for matrix shaping (including override
ENOENT), `failingClient`/`hangingClient` for host-resolution
failure/abort. New coverage: host-source success + provenance
shape (with a profile model configured, proving D1's bypass),
host failure/cancel, host-fallback panel finalization
(dispatch-review-runs or consolidate-dispatch), re-dispatch
of a failed host run (dispatch-recovery), a direct-bypass
precedence case (RD-FR-037 scenario "Direct bypass overrides
host resolution"), and a variant-less host-result sub-case
(`resolved_parent_variant: null` with clean finalization —
RD-FR-037 scenario "Host run without an active variant").

### D6. Cross-delta supersession (MODIFIED entries)

RD-FR-037 falsifies two live requirement texts in earlier
unarchived deltas: the advisor-host-model-fallback delta's
"Non-advisor with empty profile rejected" scenario (a source
"host" entry with no profile model now resolves the host model
or fails retryable `host_model_unavailable` — never
`tier_model_unavailable`) and the shared-agent-executor
dispatch-agent-run delta's unconditional tier-default clause (a
source "host" entry bypasses it). The change therefore carries
two MODIFIED entries with Previously paragraphs (MP-013): one
in the review-dispatch delta (Advisor Host-Model Fallback,
narrowed to exclude source "host") and one in a new
dispatch-agent-run capability delta (dispatch_agent_run Tool,
tier default scoped to non-host sources). Both delta headers
declare archive ordering after their predecessors per the
accept-variantless-parent-model precedent; neither older change
is edited in place.

## Risks / Trade-offs

- **Behavior change for host runs with a profile model
  configured**: they now execute on the host model instead of the
  standard-profile model. This is the intended spec alignment —
  the plan never requested the profile — but operators who
  relied on the accidental override-inheritance for host-fallback
  reviewers will see a different model. Mitigation: the advisor
  list remains the sanctioned way to pin a persona to a profile;
  the CHANGELOG entry documents the change.
- **Host-model availability becomes a new dependency for host
  runs without a profile model** (previously instant
  `tier_model_unavailable`): each host run now performs a
  session-message lookup before child execution. Cost is one
  client call, identical to the advisor fallback; failure is
  retryable and sanitized.
- **Provenance-shape compatibility**: no schema fields change;
  host runs already validate against DispatchRunSchema with the
  host shape (the finalizer's host branch predates this change).
  Consumers reading `requested_model` for host runs already
  expect null per RD-FR-008.
- **Coverage floor risk**: the 94 percent statements floor is
  enforced by `make plugin-test`; the new branch is small and
  covered by the mirrored advisor-pattern tests, so the floor is
  expected to hold. If a gate fails, the gatekeeping rule
  applies: report and stop, never weaken the floor.
