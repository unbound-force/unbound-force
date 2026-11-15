## Context

Issue #1009, Tier 1 of the v0.22.0 release plan (#1060). All
line numbers refer to `.opencode/plugins/review-dispatch/index.ts`
at this change's base.

**Gap 1 — planner.** `profilePair()` (L1471) resolves a profile
to `{model, variant, tier}`, returning `model: null` when the
profile omits `model` (legal since `optional-profile-models`,
matrix version 3; the shipped v3 scaffold at
`internal/scaffold/assets/uf/review-matrix.yaml` ships
`lightweight/standard/heavy` all empty). The planner's run
resolution produces include-entries with `model: null` from
three advisor-destination sites — `baseRuns()`'s advisor branch
(`matrix.always?.[agent] ?? matrix.defaults[mode]`, L1584-1595),
`fullPanelRuns()` (standard profile, L1680), and
`addAugmentation()` (adjacent tier, L1638) — and from
`explicitRuns()`'s profile form (`runs.<mode>.<agent>` entries
with `{profile}`, L1555-1563). None of these sites signals the
condition: the plan emits `status: "ready"` with no advisory.

Execution treats the two families differently
(`executeDispatchRun`, L3346-3410): the run input carries
`source: entry.source` (L7066) and tier-when-model-null
(`resolvePlanEntryRuntime`, L3184), so an `advisor`-source run
with a model-less profile inherits the parent session's model
(#693) or fails with `host_model_structural` /
`host_model_unavailable` (RD-FR-038), while an `explicit`-source
run with a model-less profile falls to the non-retryable
`tier_model_unavailable` early exit (L3397-3404) — no run file,
no child session, a dispatch that can never succeed. A
model-less non-tier profile name is the sharper variant: the
entry's `tier` is null, so execution's `options.tier ??
"standard"` default resolves the STANDARD profile's model
silently.

**Gap 2 — aggregate summary.** `dispatchReviewRuns`'s batch
loop (L7113-7128) counts a fulfilled-but-failed result in
`runsFailed` and discards its `error` — pre-execution early
exits (`tier_model_unavailable`, `host_model_structural`,
`host_model_unavailable`, `unknown_agent`) never reach
`warnings`; only rejected batch promises produce the existing
`run rejected:` warning (L7127). The aggregate return
(L7169-7181) sets `status: "error"` when `runs_failed > 0 &&
runs_completed === 0` but sets no `message`, so
`dispatchSummaryLine` (L7233) renders its `no failure message
reported` fallback even though `DispatchReviewRunsOutputSchema`
already has an optional `message` (L5745) and the #946 render
contract (one line, 512 chars, retryable omitted when unknown)
already handles it. The visible summary is the only channel a
retry loop sees.

Constitution alignment carries over from the proposal (all six
active principles of the org constitution v1.4.0): the signals
become self-describing plan and result artifacts (I), reuse the
existing advisory, validation, warnings, and message channels
with no new dependency (II), are machine-parseable with
provenance — profile, agent, sequence, code (III), testable
with the existing vitest doubles (IV), carry only
already-sanitized error text and repository configuration names
(V), and are deterministic runtime dispatch code with no
instruction-prose changes (VI).

## Goals / Non-Goals

### Goals

- Planner signals advisor host-model inheritance: one
  non-truncating, deterministic, LOW-severity plan advisory per
  distinct model-less profile behind advisor-destination
  include-entries; plan stays `ready` (RD-FR-063).
- Planner fails closed on unexecutable explicit profile runs:
  validation error naming the profile, entry `skip`, plan
  `INCONCLUSIVE` before any child dispatch (RD-FR-064).
- Aggregate captures every failed run's cause into `warnings`
  and folds the first cause into the all-failed `message` so
  the summary line carries it under the #946 bounds
  (RD-FR-065).

### Non-Goals

- No execution-semantics change: model resolution, host-model
  inheritance, error classes, and retryability are untouched.
- No plan schema or `plan_version` change: plan entries keep
  their strict shape; advisories already exist on `DispatchPlan`.
- No aggregate `retryable` on the all-failed path: a mixed
  failure set has no honest single value, and the sanctioned
  recovery paths (caller re-invocation after remediation,
  DR-FR-002 redispatch where run files exist) do not need it.
  The #946 marker stays "omitted when unknown".
- No skill or command prose changes: the dispatch-advisor
  skill's "invalid explicit run ... returns INCONCLUSIVE"
  already covers the D2 rejection class, and its plan contract
  does not enumerate advisories.
- No fix for the model-configured non-tier explicit profile
  quirk (entry `tier: null` → execution's `tier ?? "standard"`
  default): D2 rejects only the model-less combination; the
  model-configured mismatch is a separate latent issue to file
  if observed.

## Decisions

### D1. Advisory for advisor-destination runs (RD-FR-063)

`ResolvedRun` gains an internal `profileName: string | null`
field, threaded from the four resolution sites (advisor branch:
the selected profile; full panel: `"standard"`; augmentation:
the adjacent tier; explicit runs: the named profile, null for
the direct-model form). The field is internal provenance for
advisory emission — the `PlanEntry` mapping (L2030-2042) does
not emit it, so the strict plan entry schema and `plan_version`
are unchanged. In `planReviewDispatch` Pass 2, each
`decision: "include"` run with `source: "advisor"` and
`model: null` accumulates its agent into a profile-keyed map;
after Pass 2 the planner emits one advisory per distinct
profile: severity `LOW`, naming the profile, the affected
agents (plan order), the host-model inheritance fallback, and
the remediation. Advisories are appended to the existing
advisory block (L2067+) and sort by profile name.

Profile-name threading beats re-deriving the profile from
entries: `profilePair` sets `tier` only when the profile name
parses as a tier, so a non-tier profile selected via
`always`/`defaults` is unrecoverable from the emitted entry.
Threading keeps the name exact by construction.

Rejected: failing closed here — it would regress the #693
sanctioned fallback and mark every scaffold-matrix deployment
(the shipped default state) un-dispatchable. Rejected:
per-entry advisories — with the scaffold matrix every advisor
entry triggers the condition; one advisory per profile keeps
the signal without the noise.

### D2. Fail closed for explicit profile runs (RD-FR-064)

In `explicitRuns()`'s profile branch, after `profilePair`
succeeds, `model: null` records a validation error: the profile
name, that explicit-source runs cannot inherit the host model,
and the remediation (`profiles.<name>.model` in the matrix or
the gitignored override). The existing machinery does the rest:
validation errors flip the entry's `decision` to `skip`
(L2032), feed the plan `errors` array (L2027-2029), and
non-empty errors force `status: "inconclusive"` /
`workflow_result: "INCONCLUSIVE"` (L2095-2100) — identical to
today's unknown-profile rejection. The direct-model form and
advisor/host destinations are untouched; the check also closes
the model-less non-tier variant where execution would
silently resolve the standard profile.

Rejected: an advisory for this case — the host-model
inheritance fallback does not apply to explicit sources, so the
advisory text would misdescribe execution; the run is
deterministically unexecutable, which is exactly the
dispatch-advisor contract's fail-closed class.

### D3. Failure-cause capture (RD-FR-065)

The batch-result loop's fulfilled-but-failed branch
(`runResult.error !== null`) pushes
`run failed: <agent>/<sequence>: <code>: <message>` into
`warnings`. `Promise.allSettled` preserves input order, so the
settled results zip with the batch array for the agent and
sequence. Rejected promises keep the existing
`run rejected: <reason>` warning. Warning volume is bounded by
the run count (limits cap total runs at 24, personas at 16).
Capture applies on every completion, including partial
failure — the metadata channel carries causes uniformly.

### D4. All-failed message fold (RD-FR-065)

The loop maintains a first-cause variable, set once from the
first fulfilled failure (`<code>: <message>`) or rejection
reason in batch order. The aggregate return sets `message` to
that cause when `runs_failed > 0 && runs_completed === 0`,
falling back to `warnings[0]` (earlier-phase warnings such as
sibling-evidence or limits degradation) when no run cause
exists; the `no failure message reported` placeholder inside
`dispatchSummaryLine` survives only when both are absent.
`dispatchSummaryLine` is unchanged — it already renders
`result.message` under the #946 contract, so the 512-character
bound, line collapsing, and the retryable-omitted rule apply
for free, and the cause lands in both the metadata `message`
and the visible summary. First-cause precedence over
`warnings[0]` avoids surfacing an unrelated early warning when
run failures exist.

### D5. Tool description sentence

The `dispatch_review_runs` tool description's error-summary
sentence gains a clause noting that an all-failed dispatch
reports the first run-failure cause as its message. The
description is part of the mirrored plugin file; no other
prose changes.

## Testing Strategy

- **Planner (RD-FR-063/064)** —
  `.opencode/test/review-dispatch.test.ts`, driving
  `planReviewDispatch` with fixture matrices: delete a
  profile's `model` and assert the advisory (profile named,
  agents listed, `status` stays `ready`, exactly one advisory
  per profile, determinism across two runs); full-panel and
  augmentation variants; the no-advisory and host-exemption
  negatives; and the D2 matrix (`runs.code.<agent>` with
  `{profile: heavy}`, `profiles.heavy` model-less) asserting
  `inconclusive`/`INCONCLUSIVE`, the skip decision, and the
  error naming agent and profile. The existing
  `optional profile models` test (model-null advisor entry)
  stays green — the advisory is additive.
- **Aggregate (RD-FR-065)** —
  `.opencode/test/dispatch-review-runs.test.ts`, driving
  the real tool: a plan whose single entry is an
  explicit run against a model-less profile (D2 rejects at plan
  time, so drive the failure with an advisor entry whose host
  resolution fails, or a direct entry the fake client fails)
  asserting `status: "error"`, `message` carrying the code and
  profile, the summary line ending with the cause, and the
  `run failed:` warning naming agent/sequence/code/message;
  a partial-failure case asserting `ok` with the cause still
  in `warnings`; the 512-character bound via
  `dispatchSummaryLine` on a hand-built result; determinism of
  first-cause selection.
- **Suite sweep** — update any existing assertion that pinned
  empty `warnings` on failed-run scenarios or a missing
  `message`; the `no failure message reported` pin at
  `dispatch-review-runs.test.ts` (retryable-true case with
  empty warnings) stays valid by construction.
- **Gates** — `make plugin-test` (94 percent statements floor;
  gatekeeping: report and stop on failure, never adjust
  `.opencode/vitest.config.ts`) and
  `go test -race -count=1 ./internal/scaffold/...` after the
  mirror sync.

## Risks

- **Advisory volume on scaffold matrices**: every plan against
  a default scaffold matrix gains one LOW advisory per
  model-less profile in use. Accepted — that invisibility is
  the defect being fixed; the advisory names the remediation.
- **Existing tests pinning silence**: assertions on empty
  `warnings` or absent `message` for failed runs will break by
  design; the sweep task addresses them, and new assertions
  pin the new shapes.
- **Mirror drift**: the live plugin and scaffold asset must
  stay byte-identical; the drift gate is intentionally red
  from the first live-plugin edit until the sync task, so the
  phase checkpoints gate the vitest suite only.
