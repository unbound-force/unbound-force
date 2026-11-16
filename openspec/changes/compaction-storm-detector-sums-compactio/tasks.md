<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Probe accounting (delegation-budget.ts)

- [x] 1.1 Scope the compaction accounting per session (DT-FR-002
  modification, DELEG-FR-002) — classification: judgment
  (authorship). Add `consecutiveCompactions` to
  `ProbeSessionState`; a new marker in a session increments that
  session's counter, substantive progress in a session resets
  only that session's counter; remove the root-level summed
  `compactionsSinceProgress`; replace the `ProbeCounters` surface
  with `maxSessionCompactionsSinceProgress` plus
  `stormSessionID` (null when the max is zero; ties resolve to
  the first session in observation order), keeping
  `nonReadProgressCount` and `compactionCount` subtree-wide;
  thread the same fields through `ActivitySample` and
  `registeredProbeCounters`. Update the module's design-D2
  doc comments to the per-session contract. Files:
  `.opencode/lib/delegation-budget.ts`,
  `internal/scaffold/assets/opencode/lib/delegation-budget.ts`
  (byte-identical mirror; drift tests pin parity).

- [x] 1.2 Exclude executor-managed descendants from the parent's
  storm count (DELEG-FR-002) — classification: judgment
  (authorship). In the fold, a descendant session with
  `registeredDeadline(id) !== undefined` is executor-managed:
  record its marker keys and observability totals but accrue
  nothing toward any per-session storm count; count the probe
  root unconditionally. Marker keys are recorded for every
  observed session so a marker first seen while managed is never
  counted retroactively after that run's deadline is
  unregistered. Unrecognized descendants (no visible deadline)
  count per session under 1.1's rule. Runs after 1.1 (same
  files). Files: `.opencode/lib/delegation-budget.ts`,
  `internal/scaffold/assets/opencode/lib/delegation-budget.ts`.

## 2. Executor surface (agent-executor.ts)

- [x] 2.1 Consume the per-session storm signal and name the
  storming session (DELEG-FR-002) — classification: judgment
  (authorship). The interval detector aborts on
  `maxSessionCompactionsSinceProgress >= stormThreshold`,
  recording the max and `stormSessionID`; add
  `storm_session_id: string | null` to `TimeoutDetails` (null
  for non-storm kinds); rewrite both `context_thrash` messages
  (interval abort and final-text reclassification) to name the
  storming session and its consecutive count; the final-text
  guard reads the same fields from the final probe or the
  `registeredProbeCounters` fallback, keeping
  `synthetic_final_text` (threshold not met) unchanged. No
  plugin code change: `uf_delegate_command` surfaces the
  code-prefixed message and `timeout` details by pass-through.
  Files: `.opencode/lib/agent-executor.ts`,
  `internal/scaffold/assets/opencode/lib/agent-executor.ts`
  (byte-identical mirror).

## 3. Tests (parallel-eligible)

- [x] 3.1 [P] Probe-accounting unit tests — classification:
  judgment (test authorship). Migrate the existing
  `probeSubtreeActivity` accounting rows to the new surface and
  add regressions for issue #1125: three sibling sessions with
  one compaction each and only read calls in between keep the
  max at one (no abort signal); one session with N consecutive
  compactions reaches max N and `stormSessionID` names it;
  substantive progress in one session resets only that session's
  counter; a descendant with a registered deadline is excluded
  from the max while `compactionCount` still totals its markers;
  no retroactive counting after `unregisterDeadline` drops the
  run's deadline; the probe root is counted even though its own
  executor registered its deadline; failed probes still change
  no accounting. Files:
  `.opencode/test/delegation-budget.test.ts`.

- [x] 3.2 [P] Executor storm-detector tests — classification:
  judgment (test authorship). Regressions for the #1125
  acceptance criteria: a subtree of three reviewer-like sibling
  sessions at one compaction each with reads only completes
  without `context_thrash`; one session with N consecutive
  compactions and no non-read progress aborts with
  `code: "context_thrash"`, `timeout.kind: "storm"`, a resume
  envelope, and a message plus `storm_session_id` naming the
  session; existing threshold/reset/read/final-text rows
  migrated to `maxSessionCompactionsSinceProgress`; the
  final-text guard's `context_thrash` reclassification names the
  session and `synthetic_final_text` stays threshold-gated.
  Files: `.opencode/test/agent-executor.test.ts`.

- [x] 3.3 [P] Delegation surfacing tests — classification:
  judgment (test authorship). The `uf_delegate_command` storm
  abort envelope carries `storm_session_id` in its threaded
  `timeout` details and the code-prefixed message names the
  storming session; resume state threading unchanged. Files:
  `.opencode/test/unleash-delegate.test.ts`.

## 4. Docs and verification

- [x] 4.1 Documentation — classification: judgment (authorship).
  Rewrite the `docs/configuration.md`
  "Compaction-storm threshold" section for per-session
  semantics (sibling compactions never sum; executor-managed
  descendants are covered by their own detectors; the probed
  root stays counted) and note that operators who raised the
  threshold to tolerate parallel reviewers may lower it again;
  add the `storm_session_id` field to the documented timeout
  envelope shape. Add the `CHANGELOG.md` Fixed entry under
  Unreleased referencing #1125 and this change. Files:
  `docs/configuration.md`, `CHANGELOG.md`.

- [x] 4.2 Verification: constitution alignment and repo checks —
  classification: deterministic (exact-command verification;
  skill/script reference: repo-root `Makefile`, targets `check`
  and `lint-instructions`). Confirm I Autonomous Collaboration
  (envelope-only, additive provenance) / II Composability (no
  new dependencies; deadline registry reused) / III Observable
  Quality (`storm_session_id` machine-parseable) / IV
  Testability (fake-client harness, injected clock) / V Security
  by Default (exclusion fails toward detection; no gate
  weakened) / VI Deterministic Execution (counting and
  exclusion live in tested tools); run the checks via the
  `Makefile` targets, including the scaffold-mirror drift tests
  and the plugin suite.
