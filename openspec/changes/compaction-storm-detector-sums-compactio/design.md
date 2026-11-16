## Context

`executeAgentSession` (`.opencode/lib/agent-executor.ts`) probes
its child's session subtree on an interval through
`probeSubtreeActivity` (`.opencode/lib/delegation-budget.ts`).
Since #1100 the same probe feeds a compaction-storm detector:
`foldProbeObservations` classifies each session's latest message
(`isCompactionMarker`), identity-gates it per session
(`markerKey` / `lastMarkerKey`), and maintains ONE
`compactionsSinceProgress` counter on the root's `ProbeState`.
A new marker in *any* subtree session increments the counter; a
completed non-read tool call (`isSubstantiveProgress`) in *any*
session resets it. The executor aborts with
`code: "context_thrash"`, `timeout.kind: "storm"`, and a resume
envelope when the counter reaches `stormThreshold`
(`normalizeStormThreshold`, default 3, range 2-10). The
final-text guard re-checks the same accounting when a child's
final text structurally matches a synthetic wrapper.

The subtree is where the design breaks. During spec review,
`/uf.unleash` fans out six reviewer sessions in parallel, and
each nested `uf_delegate_command`/`invoke_agent` run is itself an
executor-created session watched by its own always-armed
detector. Muse-Spark reviewers compact 1-3 times each (normal
for a tight-context model) and are read-heavy, so nothing resets
the shared counter. Markers from *different* sessions sum past
the threshold and the parent delegation aborts with
`child_active: true`, cancelling healthy in-flight runs (issue
#1125, four false aborts on 2026-11-15).

Three structural facts shape the fix:

1. Every executor-created session has a run deadline registered
   in the process-wide deadline registry
   (`registerDeadline` at child creation, `unregisterDeadline`
   in the run's `finally`) and is probed by its own executor --
   the detector is always armed, so no executor-created session
   goes unwatched.
2. The #1079 incident (dewey#203) was a storm in the probe
   *root* itself: the delegated `/uf.unleash` session compacted
   ~150 times. Root counting must stay.
3. The per-session identity gates (`lastMarkerKey`,
   `lastProgressKey`) already exist in `ProbeSessionState`; only
   the *counter* is shared. Per-session counters are a small,
   local change to the fold.

## Goals / Non-Goals

### Goals

- Scope the storm count per session: sibling sessions each below
  the threshold never sum to an abort (issue #1125 acceptance
  criterion 1).
- Keep the single-session abort: one session with N consecutive
  compactions and no non-read progress still aborts with
  `kind: "storm"` and a resume envelope (acceptance criterion 2).
- Exclude executor-managed descendants from the parent's count:
  a reviewer or nested delegation with its own executor and
  limits trips only its own detector.
- Name the storming session in the envelope so recovery can tell
  one thrashing model from a slow, healthy dispatch (acceptance
  criterion 3).
- Preserve the #1079 protection: the probed root stays counted.
- Keep the threshold knob, its range/default/validation, the
  failure codes, and the resume envelope unchanged.

### Non-Goals

- Fixing #1117's idle-timeout false aborts (same class, idle
  path; a separate change -- though this change must not
  diverge the shared accounting it will build on).
- Changing model selection (Muse-Spark still being picked for
  reviewers is noted in #1125 as the root cause of frequent
  compactions; the headroom gate is DELEG-FR-001's concern).
- Making the parent selectively abort a storming descendant run
  (the issue's alternative formulation). The chosen design
  reaches the same outcome -- only the storming run ends -- by
  letting the descendant's own always-armed detector abort it,
  with no new parent-to-descendant abort path.
- Changing the orchestrator's single storm re-delegation
  protocol; with false aborts gone, that budget is no longer
  wasted, but the protocol itself is untouched.

## Decisions

### D1 — Per-session consecutive accounting
(delegation-budget.ts)

**Decision:** `ProbeSessionState` gains a mutable
`consecutiveCompactions` counter. A new marker in session S
increments S's counter (and the root-level `compactionCount`
observability total); substantive progress in session S resets
*S's* counter to zero (and increments `nonReadProgressCount`).
The root-level summed `compactionsSinceProgress` field is
removed. The fold computes the storm signal as the maximum
per-session `consecutiveCompactions` among *counted* sessions
(D2) and exposes it with the identity of the holding session:

```ts
interface ProbeCounters {
  readonly maxSessionCompactionsSinceProgress: number
  readonly stormSessionID: string | null
  readonly nonReadProgressCount: number
  readonly compactionCount: number
}
```

`ActivitySample` and `registeredProbeCounters` carry the same
fields. The old `compactionsSinceProgress` field is removed
outright rather than redefined: every consumer (the interval
detector, the final-text guard, the tests) is internal and is
updated in this change, and a rename forces each consumer to
confront the new semantics instead of silently misreading a
familiar name. `stormSessionID` is null exactly when the maximum
is zero. Ties (several sessions sharing the maximum) resolve
deterministically to the first such session in the fold's
observation order (the probe's newest-first walk).

The reset is strictly per session: a completed non-read tool call
in session A must not mask compactions in session B. This
preserves the identity-gating design from #1100 (a quiet
session's stale completed tool cannot reset anything twice --
`lastProgressKey` gating stays) while removing the cross-session
reset that let one busy session hide another's storm, and the
cross-session *sum* that let healthy siblings impersonate one.

**Alternatives:** (a) Keep the summed field and add a max
alongside -- rejected: two counters over the same markers invite
drift and the summed one has no remaining consumer. (b) Expose a
full per-session map (`Record<string, number>`) -- rejected: the
acceptance criteria need the maximum and its holder only; a map
bloats every sample and envelope for diagnostics nobody asked
for. (c) Count only the probe root and ignore all descendants --
rejected: unmanaged descendants (subagents spawned outside any
executor) would lose storm detection entirely; per-session
counting plus D2's exclusion keeps them covered.

**Constitution:** IV Testability -- a deterministic fold testable
through the existing `treeClient` harness; III Observable
Quality -- the sample names the storming session.

### D2 — Executor-managed descendants are excluded
(delegation-budget.ts)

**Decision:** At fold time, a descendant observation is
*executor-managed* when `registeredDeadline(sessionID) !==
undefined` -- another `executeAgentSession` run created that
session and registered its deadline at child creation, and that
run's detector is always armed. Managed descendants' markers are
recorded (identity gates and observability totals update) but do
NOT accrue toward any per-session storm count. The probe root is
counted unconditionally: it is *this* executor's direct child
(the #1079 case), even though this executor also registered its
deadline.

Marker keys are recorded for every observed session, managed or
not. This is load-bearing: a finished reviewer's deadline is
unregistered when its run ends, and its last marker remains its
latest message; because the key was already seen, the marker is
never retroactively counted after the run ends. Counting
therefore depends on management status *at first observation*,
which is exactly the run window.

Unrecognized descendants -- sessions in another host process,
where the process-wide registry is not shared -- are counted per
session under D1's rule. The failure direction is conservative
(more detection, never silence), and per-session counting still
excludes the summed false-abort class that produced #1125.

**Alternatives:** (a) Probe-state registry membership
(`probeRegistry().has(id)`) as the managed signal -- rejected:
probe state exists only after the descendant's executor completes
its first successful probe, leaving a first-interval window where
a fresh reviewer's marker would count. (b) Session title or
agent-name heuristics -- rejected: fragile, non-mechanical,
and untestable against representation drift. (c) Walk-depth
cutoff (count only depth 0) -- rejected: cannot distinguish a
managed reviewer from an unmanaged subagent at the same depth.

**Constitution:** II Composability -- reuses the existing
deadline registry (delegation-timeout-budgets D1); no new
dependency. V Security by Default -- fails toward detection, not
silence.

### D3 — Storming-session identity through the executor
surface (agent-executor.ts)

**Decision:** `TimeoutDetails` gains
`storm_session_id: string | null` (null for every non-storm
kind; non-null on both storm paths). The interval detector
checks `activity.maxSessionCompactionsSinceProgress >=
stormThreshold`, records the max and `stormSessionID` at abort,
and the `context_thrash` message names both, e.g. `invoke_agent
run recorded N consecutive compaction events with no substantive
progress in session <id> (storm threshold T); aborting so the
caller can resume from durable state`. The final-text guard's
reclassification reads the same fields from the final probe (or
the `registeredProbeCounters` fallback) and its
`context_thrash` message names the session identically;
`synthetic_final_text` (threshold not met) keeps its existing
message. `uf_delegate_command` surfaces the code-prefixed
executor message and threads `timeout` into the command result
by pass-through, so the plugin needs no code change for the
envelope; its tests assert the new message shape and
`storm_session_id`.

**Alternatives:** (a) A per-session compaction breakdown inside
`TimeoutDetails` -- rejected: envelope bloat; the id plus count
satisfies the acceptance criterion. (b) Resolving the session id
to a model/agent name in the envelope -- rejected: the executor
does not know descendant metadata reliably, and the session id is
the key recovery already has (the orchestrator can resolve it);
the message stays bounded and sanitized.

**Constitution:** III Observable Quality -- machine-parseable
`storm_session_id` with the storm kind; I Autonomous
Collaboration -- the envelope remains a self-describing artifact.

### D4 — Docs, knob semantics, and mirrors

**Decision:** `delegation.storm_threshold` keeps its range
(2-10), default (3), and fail-closed validation; its meaning
tightens from a subtree sum to a per-session run, so every
configuration aborts strictly less often -- no migration, no
gate changes. `docs/configuration.md`'s "Compaction-storm
threshold" section is rewritten for per-session semantics with a
note that operators who raised the threshold to tolerate
parallel reviewers may lower it again. `CHANGELOG.md` gains a
Fixed entry. Every code change to `.opencode/lib/
delegation-budget.ts` or `.opencode/lib/agent-executor.ts` is
mirrored byte-for-byte into `internal/scaffold/assets/opencode/
lib/` (the drift-detection tests in `scaffold_test.go` pin
parity; the mirrors exist and are currently identical).

**Constitution:** V Security by Default -- no gate is weakened;
VI Deterministic Execution -- semantics live in the tested tools,
docs describe them.

## Contracts

Changed runtime surfaces (all internal to the plugin suite; the
external failure envelope is additive only):

- `ProbeCounters` / `ActivitySample` / `registeredProbeCounters`
  (delegation-budget.ts): `compactionsSinceProgress` is removed;
  `maxSessionCompactionsSinceProgress: number` and
  `stormSessionID: string | null` are added;
  `nonReadProgressCount` and `compactionCount` keep their
  subtree-wide meanings (markers in managed descendants still
  count toward `compactionCount`).
- `TimeoutDetails` (agent-executor.ts):
  `storm_session_id: string | null`, non-null only on
  `kind: "storm"`.
- `context_thrash` error messages (interval abort and final-text
  reclassification): name the storming session id and its
  consecutive count; codes, retryability, and the resume
  envelope are unchanged.
- No changes to `InvokeAgentResult` status codes, the resume
  envelope shape, `delegation.storm_threshold` validation, or
  any uf-workflow plugin code.

These surfaces are the seams verified in tasks.md (sample shape,
exclusion behavior, envelope round-trip through
`uf_delegate_command`).

## Risks / Trade-offs

- **A storming reviewer now fails its own run, not the parent.**
  The dispatch sees a retryable failed run (existing machinery:
  `redispatch_failed_runs`, DR-FR-025 exclusion) while the
  delegation continues -- this is the issue's requested outcome.
  If a dispatch fans out many tight-context reviewers, each may
  fail individually; that is a selection problem (DELEG-FR-001,
  out of scope), not a detector problem.
- **Cross-process descendants are unrecognized as managed** and
  are counted per session. A cross-process reviewer sitting
  exactly at N consecutive compactions with no non-read progress
  would trip the parent -- but that is a genuine single-session
  storm by the #1079 definition, so the parent aborting is the
  conservative and correct direction. The #1125 false-abort
  class (summed siblings) is eliminated regardless of process
  topology.
- **Tie-breaking names one session.** When several sessions
  share the maximum, the envelope names the first in observation
  order -- deterministic, and recovery can re-probe for the rest.
- **Threshold semantics tighten silently.** Per-session counting
  is always <= the old sum, so no configuration becomes more
  abort-happy; operators who previously raised the threshold as
  a workaround may now lower it (documented in D4). Not lowering
  it costs only later aborts of genuine single-session storms.
- **Mirror drift.** The scaffold mirrors must be updated in the
  same tasks as the lib files; the drift tests fail closed, so
  drift cannot ship silently.

## References

- Issue #1125 (this change); #1079 / #1100 (storm detection
  being fixed); #1117 (same false-abort class, idle path, out of
  scope); dewey#203 (the protected root-storm incident).
- Epic unbound-force/unbound-force#1060 (release v0.22.0),
  Tier 1; sibling issue #1118 shares the evidence class.
- Files: `.opencode/lib/delegation-budget.ts`,
  `.opencode/lib/agent-executor.ts`,
  `internal/scaffold/assets/opencode/lib/delegation-budget.ts`,
  `internal/scaffold/assets/opencode/lib/agent-executor.ts`,
  `.opencode/test/delegation-budget.test.ts`,
  `.opencode/test/agent-executor.test.ts`,
  `.opencode/test/unleash-delegate.test.ts`,
  `docs/configuration.md`, `CHANGELOG.md`.
- Prior designs: compaction-storm-on-a-tight-context-mode D2
  (the accounting being refined), delegation-timeout-budgets D1
  (the deadline registry reused by D2 here).
