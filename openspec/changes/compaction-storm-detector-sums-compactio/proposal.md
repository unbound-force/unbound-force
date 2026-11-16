## Why

The compaction-storm detector added in #1100 (fixing #1079) keeps
one counter for the whole delegation subtree. In
`.opencode/lib/delegation-budget.ts`, `foldProbeObservations`
accumulates a single `compactionsSinceProgress` on the root's
`ProbeState`: a compaction marker in *any* session of the subtree
increments it, and only substantive progress (a completed non-read
tool call) in *any* session resets it.

During spec review, `/uf.unleash` dispatches six reviewers in
parallel. The Muse-Spark reviewers each compact 1-3 times --
normal for them -- and they are read-heavy, so nothing resets the
counter. Three reviewer compactions in *different* sessions reach
the default threshold of 3, and the parent `/uf.unleash` delegation
is aborted with `context_thrash` / `kind: storm` while
`child_active: true`. Every reviewer run still in flight is
cancelled. No single session is in a storm.

Four such false storm aborts occurred on 2026-11-15, all on
builds including #1100, every envelope reporting
`child_active: true` (issue #1125):

- `ses_e1f783391ffe...` (#1117 pipeline, local): opsx-propose
  grandchild on Muse-Spark, 4 compactions.
- `ses_e1f6e01abffe...` (#1117 pipeline, local): spec-review
  dispatch -- architect, curator, and adversary at 1 compaction
  each, plus 4 in the propose grandchild.
- `ses_e1f5eed24ffe...` (#1117 pipeline, local): spec-review
  dispatch -- architect 2, curator 1, adversary 3; the curator had
  already succeeded and 5 runs were cancelled mid-dispatch.
- `ses_e1f89ad63ffe...` (#1118 pipeline, Linux worker):
  spec-review dispatch -- architect 1, curator 1, adversary 1,
  guard 2.

In the #1117 run the third abort exhausted the single storm
re-delegation the protocol allows; the orchestrator had to recover
the dispatch from its own session with `redispatch_failed_runs`.
This is the same false-abort class #1117 describes for
`idle_timeout`, arriving through the storm path.

## What Changes

- **Per-session storm accounting (DELEG-FR-002 refinement):**
  consecutive compaction events MUST accrue against the session in
  which they occur, and a session's run resets only on substantive
  progress *in that session*. The storm signal is the highest
  per-session run of consecutive compactions, never the subtree
  sum. Parallel sibling sessions each below the threshold -- for
  example three reviewers with one compaction each and only `read`
  calls in between -- MUST NOT abort the parent.
- **Executor-managed descendants do not trip the parent:** a
  descendant session that is itself an executor-managed run
  (identified by a run deadline registered for it in the
  process-wide deadline registry) carries its own always-armed
  storm detector and wall clock. Its compaction markers MUST NOT
  count toward the parent's storm signal. A genuinely storming
  reviewer therefore ends *its own* run through *its own*
  detector, and the dispatch handles the failed run while the
  delegation keeps running.
- **The probe root stays counted:** the delegation's direct child
  session is always subject to storm counting, preserving the
  #1079 protection -- the dewey#203 case (about 150 compactions in
  one `/uf.unleash` session) still aborts as `context_thrash` with
  a resume envelope.
- **The storm envelope names the storming session:** the probe
  sample exposes the session holding the highest per-session run;
  the `context_thrash` failure's limit details carry a
  machine-parseable storm session id, and the error message names
  the session and its count, so recovery can tell one thrashing
  model from a slow, healthy dispatch.
- **Regression tests** for both acceptance criteria in #1125
  (sibling no-abort; single-session abort with `kind: storm` and a
  resume envelope) plus the session-named envelope.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `delegation`: DELEG-FR-002 (Compaction-storm detection) --
  counting is scoped per session, executor-managed descendants are
  excluded from the parent's count, and the storm failure names
  the storming session. The threshold knob, its range, default,
  and fail-closed validation are unchanged.
- `delegation-timeouts`: DT-FR-002 (Idle Timeout) -- the shared
  probe-accounting clause is refined to per-session compaction
  accounting. The idle clock itself (fingerprint-based) is
  unchanged.

### Removed Capabilities

- (none)

## Impact

- **Affected systems/files:** `.opencode/lib/delegation-budget.ts`
  (per-session fold, managed-descendant exclusion, sample surface)
  and its tool-owned scaffold mirror
  `internal/scaffold/assets/opencode/lib/delegation-budget.ts`;
  `.opencode/lib/agent-executor.ts` (interval detector, final-text
  guard, `TimeoutDetails.storm_session_id`, error messages) and
  its mirror
  `internal/scaffold/assets/opencode/lib/agent-executor.ts`;
  tests `.opencode/test/delegation-budget.test.ts`,
  `.opencode/test/agent-executor.test.ts`,
  `.opencode/test/unleash-delegate.test.ts`;
  `docs/configuration.md` (storm-threshold semantics);
  `CHANGELOG.md`.
- **Behavior:** healthy spec-review dispatches with parallel
  read-heavy reviewers no longer abort; a genuine single-session
  compaction loop still aborts within minutes; a storming reviewer
  fails its own run as a retryable `context_thrash` and the
  delegation continues.
- **Compatibility:** the `ProbeCounters` surface change is
  internal -- the executor and tests are updated in the same
  change. `TimeoutDetails` gains a nullable `storm_session_id`
  (additive). Failure codes, the resume envelope, and the
  `delegation.storm_threshold` knob (range 2-10, default 3,
  fail-closed validation) are unchanged; the knob's semantics
  tighten from a subtree sum to a per-session run, so every
  configuration aborts strictly less often. No config migration;
  operators who raised the threshold to tolerate parallel
  reviewers may lower it again (noted in docs).
- **Rollout:** part of release v0.22.0 (epic #1060, Tier 1).
  Related #1117 describes the same false-abort class from the
  idle-timeout path; this change fixes ONLY the storm detector
  (#1125) and does not scope in #1117's idle-timeout work.

## Constitution Alignment

Assessed against the Unbound Force org constitution v1.4.0 (six
principles).

### I. Autonomous Collaboration

**Assessment**: PASS

The fix changes no collaboration surface. The `context_thrash`
failure, `resume` envelope, and `timeout` details remain
self-describing artifacts the orchestrator consumes
asynchronously; the new `storm_session_id` is additive provenance
inside the existing envelope, so recovery distinguishes a
thrashing session from a healthy dispatch without any new
inter-hero RPC.

### II. Composability First

**Assessment**: PASS

No new dependencies. The accounting change is local to the
delegation-budget/agent-executor seam; the managed-descendant
exclusion reuses the existing process-wide deadline registry
already required by delegation-timeout-budgets D1. The plugin
suite, executor, and matrix remain independently testable and
installable.

### III. Observable Quality

**Assessment**: PASS

The storm envelope gains a machine-parseable
`storm_session_id`, and the error message names the storming
session and its consecutive count -- exactly the evidence the
four #1125 incident envelopes had to reconstruct by hand. The
per-session accounting is observable through the existing probe
counters, and regression tests pin both acceptance criteria.

### IV. Testability

**Assessment**: PASS

All new behavior is testable in isolation through the existing
fake-client harnesses (`treeClient`, `fakeClient` with
`compactionProbe`) and the injected clock seam: sibling sessions
that must not abort, a single session that must, per-session
reset isolation, managed-descendant exclusion via
`registerDeadline`/`unregisterDeadline`, and the
no-retroactive-count property after a managed run ends. No
external services.

### V. Security by Default

**Assessment**: PASS

No new secrets, dependencies, or privileged operations. Session
ids entering error messages and limit details are opaque
identifiers already carried in envelopes; messages stay bounded
and flow through the existing sanitization path. The exclusion
signal (deadline-registry membership) fails toward *more*
detection, never toward silence: an unrecognized descendant is
still counted per session.

### VI. Deterministic Execution

**Assessment**: PASS

Per-session counting, the managed-descendant test, max
computation, and tie-breaking are mechanical procedures in
`delegation-budget.ts` / `agent-executor.ts` with tests, not
prose re-reasoned at run time. The storming-session identity is
computed by the tool and rendered into the prescribed envelope,
never reconstructed from instruction text.
