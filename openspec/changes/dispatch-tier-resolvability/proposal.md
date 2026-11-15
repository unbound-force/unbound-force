## Why

Issue #1009 (epic #1060, Release plan: v0.22.0, Tier 1). Two
closely-related observability gaps remain in the review-dispatch
plugin after #946/#949 (`review-dispatch-consumer-bounds`):

1. **The planner does not validate tier resolvability.** When an
   advisor-destination profile has no `model` configured — the
   shipped v3 matrix scaffold's default state, where every
   profile in `profiles:` is empty — `baseRuns()` →
   `profilePair()` resolves `model: null` and
   `plan_review_dispatch` emits the `advisor-run` include-entry
   anyway: `status: "ready"`, no warning, no advisory. The
   execution contract is that such runs silently inherit the
   parent session's model (advisor host-model fallback, #693,
   RD-FR-022), but the plan gives the caller no signal that this
   is happening. The symmetric defect is worse for
   explicit-source runs: a `runs.<mode>.<agent>` entry using the
   `profile` form against a model-less profile is emitted as a
   `ready` include-entry even though execution deterministically
   rejects it with non-retryable `tier_model_unavailable`
   (source `explicit` never reaches host-model inheritance), so
   the dispatch-advisor skill's fail-closed contract ("malformed
   matrix ... returns INCONCLUSIVE") is not honoured for the
   unresolvable-profile case.

2. **The all-runs-failed summary still hides causes.**
   `dispatchSummaryLine` (#946/#949) appends `result.message`,
   but on the `runs_failed > 0 && runs_completed === 0` path the
   aggregate result sets no `message`, so the visible line ends
   with `— error: no failure message reported`. The actual
   causes — `tier_model_unavailable: profile "heavy" has no
   model configured in the review matrix`, `host_model_unavailable`,
   `host_model_structural` — never reach the caller at all:
   pre-execution early exits return fulfilled-but-failed results
   whose `error` the batch loop counts and discards (only
   rejected batch promises produce today's `run rejected:`
   warnings), and rejected-promise causes live in metadata-only
   `warnings`. The visible summary is the only channel a retry
   loop sees, so a blind identical-input retry is what happens
   in practice.

Fixes #1009 (lineage: summary contract #946/#949; execution
contracts advisor-host-model-fallback #693, optional-profile-models
RD-FR-019..RD-FR-024, host-fallback-provenance RD-FR-037,
split-host-resolution-failure-classes RD-FR-038).

## What Changes

Split the planner's treatment of model-less profiles by execution
semantics, and make the aggregate's all-failed result carry the
first failure cause:

- **Advisor-destination runs (advisory, plan stays ready).**
  When an include-entry with `source: "advisor"` — baseRuns
  advisor runs, full-panel runs, and risk-augmentation runs —
  resolves to a profile with no configured model, the planner
  emits a non-truncating plan advisory naming the profile (the
  tier when the profile is a tier name) and the host-model
  inheritance fallback, with the remediation (configure
  `profiles.<name>.model` in `.uf/review-matrix.yaml` or the
  gitignored override). One advisory per distinct profile,
  deterministically ordered. `status` stays `ready` because the
  fallback is sanctioned execution (#693); failing closed here
  would break every scaffold-matrix deployment, which ships
  empty profiles by design.
- **Explicit-source profile runs (fail closed).** When a
  `runs.<mode>.<agent>` entry using the `profile` form resolves
  to a profile with no configured model, the planner records a
  validation error naming the profile, flips the entry to
  `skip`, and returns `status: "inconclusive"` /
  `workflow_result: "INCONCLUSIVE"` before any child dispatch —
  riding the existing invalid-explicit-run mechanism. Execution
  can never succeed for such a run (non-retryable
  `tier_model_unavailable` pre-execution early exit, no run
  file), so a `ready` plan is a lie.
- **Failure-cause capture.** Every failed run's cause is
  captured into the `dispatch_review_runs` result `warnings`:
  fulfilled-but-failed results (including pre-execution early
  exits) record a warning naming the agent, sequence, error
  code, and error message; rejected batch promises keep today's
  `run rejected:` warning.
- **All-failed message fold.** When `runs_failed > 0` and
  `runs_completed === 0`, the aggregate result sets `message`
  from the first run-failure cause (deterministic: first failure
  in batch order), falling back to the first warning, so the
  one-line summary renders the cause under the existing #946
  contract (collapsed to one line, at most 512 characters
  including the ellipsis). The `no failure message reported`
  placeholder survives only when no cause and no warning exist.
  `dispatchSummaryLine` itself is unchanged — it already renders
  `result.message`.

No execution-semantics change: model resolution, host-model
inheritance, error classes, and retryability are untouched. No
schema change: `DispatchPlan.advisories` and
`DispatchReviewRunsOutputSchema.message` already exist. The plan
output gains advisory entries only; `plan_version` stays 1.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `review-dispatch`: the planner signals host-model inheritance
  (new requirement RD-FR-063), fails closed on unexecutable
  explicit profile runs (new requirement RD-FR-064), and the
  aggregate dispatch result carries failure causes in warnings
  and folds the first cause into the all-failed summary message
  (new requirement RD-FR-065).

### Removed Capabilities
- None.

## Impact

- **Live plugin**:
  `.opencode/plugins/review-dispatch/index.ts` — `explicitRuns`
  (explicit no-model validation), the internal `ResolvedRun`
  shape (profile-name provenance for advisory emission, not
  emitted into plan entries), `planReviewDispatch` Pass 2
  (advisory accumulation and emission), the `dispatchReviewRuns`
  batch-result loop (failure-cause capture), and the aggregate
  result construction (all-failed message fold). The
  `dispatch_review_runs` tool description gains a sentence
  documenting the all-failed message source.
- **Scaffold asset**:
  `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`
  — byte-identical mirror of the live plugin (Go drift gate
  `TestEmbeddedAssets_MatchSource`,
  `internal/scaffold/scaffold_test.go`). No `.opencode/commands/`
  or `openspec/` files change, so no other mirrors are affected.
- **Tests**: `.opencode/test/review-dispatch.test.ts` (planner
  advisory emission, explicit fail-closed path, determinism),
  `.opencode/test/dispatch-review-runs.test.ts` (failure-cause
  warnings, all-failed message fold through the real tool),
  `.opencode/test/planning-boundaries.test.ts` (fail-closed
  boundary shape), plus any existing assertions that pinned
  empty `warnings` on failed-run scenarios.
- **Behavior change**: plans for matrices with model-less
  profiles now carry a LOW-severity advisory per affected
  profile (previously silent); explicit profile runs against
  model-less profiles now return INCONCLUSIVE at plan time
  (previously a doomed `ready` plan); all-failed
  `dispatch_review_runs` calls now end their summary with the
  first failure cause instead of `no failure message reported`.
- **No external API changes**: tool input schemas, the plan
  entry schema, the dispatch output schema, run-file shape, and
  `plan_version` are unchanged; advisories and `message` are
  additive fields that already existed.
- **Release tracking**: Tier 1 of the v0.22.0 release plan
  (#1060); must close before tagging.
- **CI replication**: `make plugin-test` (vitest suite with the
  94 percent statements coverage floor) and the Go scaffold
  drift gate (`go test -race -count=1 ./internal/scaffold/...`).
- **Documentation impact**: CHANGELOG.md entry. The change is
  user-facing (calling agents and operators consume the advisory
  and the all-failed summary cause), so the documentation gate
  applies: an in-repo documentation issue and a
  unbound-force/website documentation issue are filed before PR
  merge (precedent: #891 ↔ website#360, #913 ↔ website#368).

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0 — six
active principles).

### I. Autonomous Collaboration

**Assessment**: PASS

Both signals become self-describing artifacts. The plan advisory
travels in the authoritative plan bytes the dispatch-advisor
skill already treats as the contract, so a calling agent learns
the host-model inheritance without consulting another hero or
re-deriving matrix state. The all-failed message travels in the
one-line summary — the only channel a retry loop sees — so the
caller can remediate (fix the matrix, or use the sanctioned
recovery tools) instead of blind-retrying. No synchronous
coupling is introduced.

### II. Composability First

**Assessment**: PASS

The change adds no dependency, no new interface, and no
cross-hero requirement. It reuses the existing plan advisory
channel, the existing invalid-explicit-run validation mechanism,
the existing `warnings` array, and the existing optional
`message` field — each hero repo stays independently installable
and the plugin remains usable standalone.

### III. Observable Quality

**Assessment**: PASS

Machine-parseable signals replace silence: the advisory is a
structured `PlanAdvisory` (severity + description) in the plan
JSON; the failure causes become warning strings with a stable
`run failed: <agent>/<sequence>: <code>: <message>` shape; the
all-failed cause lands in both the result `message` (metadata)
and the rendered summary line (prose), bounded by the existing
512-character contract. Provenance (profile name, agent,
sequence, error code) is preserved end to end.

### IV. Testability

**Assessment**: PASS

Every behavior is testable in isolation through the existing
vitest doubles: the planner tests drive `planReviewDispatch`
with fixture matrices (model-less profiles, explicit profile
runs) and assert plan output; the aggregate tests drive the real
`dispatch_review_runs` tool with fake clients whose sessions
fail at model resolution. No external services or network.

### V. Security by Default

**Assessment**: PASS

No new inputs or external surface. The captured failure causes
are error codes and messages the dispatch core already produces
through `failedResult()` (sanitized at construction); the
advisory text contains only profile names and remediation paths
from repository configuration. Nothing new crosses a trust
boundary.

### VI. Deterministic Execution

**Assessment**: PASS

The change is runtime dispatch code — planner advisory emission,
validation, and result construction — byte-identical between the
live plugin and the scaffold asset, pinned by the vitest suite
and the Go drift gate. Advisory ordering (one per distinct
profile, profile names sorted, agents in plan order) and
first-cause selection (first failure in batch order) are
deterministic functions of the inputs; no instruction prose is
added or altered.
