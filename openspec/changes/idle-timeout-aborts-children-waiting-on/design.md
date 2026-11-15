# idle-timeout-aborts-children-waiting-on — Design

## Context

`idle_timeout` was added in #950 to catch model stalls where no tool calls are
made for an extended window. The implementation in
`delegation-budget.ts:probeSubtreeActivity` fingerprints each session's latest
message (`messageFingerprint`) and considers the subtree idle when the
fingerprint hash is unchanged for `idle_timeout_ms`. This misclassifies a healthy
subtree that is blocked on a single long `bash` invocation — no new messages
appear until `bash` returns, so the hash is stable by definition.

Current `messageFingerprint` already collapses synthetic `opencode-compaction-*`
markers (DT-FR-005..012) and mixes `text + output + metadataLength` to catch
streaming growth. In practice that was insufficient for `gaze crap` / `go test`
bursts: streaming is buffered and the fingerprint did not move for the entire
`idle_timeout` window. See `.opencode/lib/delegation-budget.ts` and
`.opencode/lib/agent-executor.ts` `TimeoutKind`.

Related work: `compaction-storm-on-a-tight-context-mode` (#1100 context) taught
that compaction markers MUST NOT count as progress (DT-FR-002). This change is
the complementary lesson — tool `running` state MUST count as progress.

References: constitution I–VI in `proposal.md`; affected capability
`delegation-timeouts`.

## Goals / Non-Goals

### Goals

- A `bash` (or any tool) part with `status: running` anywhere in the subtree
  counts as activity for idle detection. The running set (`partId`, `status`,
  and derived tool name) is folded into the progress fingerprint or an
  explicit `running_tool` signal.
- While any tool is running, the idle abort is deferred at least until that
  tool's own timeout plus a grace period. Tool timeout remains model-handled;
  idle remains the model-stall signal.
- The abort envelope distinguishes tool-running aborts from true stalls
  (`running_tool: "bash" | null`) and stays machine-parseable (Constitution III).
- Verification covers: 15-minute `bash` under 10-minute `idle_timeout` does not
  abort; true no-progress stall still aborts with `kind: idle`; envelope
  carries `running_tool`.

### Non-Goals

- Changing wall-clock budgets (`DT-FR-001` hierarchy, `BUDGET_RESERVE_MS`).
- Introducing a new runtime dependency or background supervisor process.
- Reworking Opencode's tool execution protocol; only the observer (probe +
  watcher) changes.
- Docs/blog/tutorial or OpenCode v2 migration (separate tracks).

## Decisions

### Decision 1 — Fold running tool parts into the progress signal

Rationale: the current fingerprint is message-centric. Running tools already
appear as parts with `state.status === "running"` on the latest message of the
active session (and, for nested delegations, on the descendant session's latest
message). Including the full `(partId, status, tool)` triple in the
fingerprint makes "still running the same bash" distinguishable from
"nothing changed", and a tool swap on the same `partId` observable.

Rejected alternative: hashing streamed output length alone. That depends on
buffering behavior and would still miss quiet-but-healthy long commands.
Hashing the sorted triple set is deterministic and independent of output
volume.

Implementation sketch: extend `ProgressSample` / `probeContribution` to emit
`runningTools: Array<{partId, status, tool}>` collected from `message.parts`
(and nested tool-call parts), sorted lexicographically by `partId`. All
three fields are hashed: the part id, its `running` status, and the tool
name (empty string when the part carries none). A new tool invocation
changes the set (progress). Continuous running of the same tool keeps
`runningTools` non-empty — see Decision 2 for how that defers abort.

### Decision 2 — Defer idle abort while any tool is running

Rationale: the expected behavior in #1117 is that a running tool defers idle
until `toolTimeout + grace`. Idle should only fire when the model is idle
between tool calls (the #1117 fifth evidence: `read.filePath` on a missing tool,
no running part, `dispatch_review_runs` never returned).

Rule: `agent-executor`'s idle watcher, on each `idle_poll_interval`, evaluates
`probeSubtreeActivity`. If `runningTools.length > 0`, the subtree is considered
active for idle purposes regardless of fingerprint stasis — but the deferral
is bounded. Session parts expose no tool deadline in the session store, so
the common case is the fixed deferral cap `IDLE_DEFERRAL_CAP_MS` (pinned at
900,000 ms, inside the spec's 5-15 minute range) measured from the last
fingerprint change; where a running part exposes an observable tool
deadline, the bound is that deadline's remaining time plus grace (multiple
running tools: the earliest remaining deadline, since the first completion
changes the fingerprint and resets both clocks). Any fingerprint change
resets the deferral accumulation; with no change, idle fires no later than
`idle_timeout + IDLE_DEFERRAL_CAP_MS`. Wall-clock (`kind: wall_clock`,
including budget-clamped runs) remains the absolute backstop for stuck
tools that never time out at the tool layer.

Rejected alternative: disabling idle entirely when a tool is running. Same
effect, but deferral wording preserves the envelope fields (`last_progress_ms_ago`
still measured; `running_tool` signals why abort was deferred). Also rejected:
shortening `idle_timeout` — that would worsen false positives.

Grace: a fixed constant pinned at `30_000 ms` (spec cap: 60,000 ms) after an
observable tool timeout to let the model observe the tool error before idle
is evaluated. Grace applies only when a running part exposes an observable
deadline; the cap path needs no grace because it is already measured from
the last fingerprint change.

### Decision 3 — Abort envelope carries `running_tool`

Rationale: operators need to distinguish "killed while bash was running" from
"model truly stalled". The `timeout` object for every kind gains
`running_tool: string | null`. Selection is deterministic: the tool name of
the running part with the lexicographically smallest `partId` (null when
none). The name is model-controlled input, so it is validated against
`^[A-Za-z0-9._:-]{1,128}$` before threading; failures report `null` and are
never logged verbatim (Constitution V). This is the machine-parseable
signal required by Constitution III and by the acceptance criteria.

Rejected alternative: logging only. Logs are not machine-parseable for the
aggregated delegation-status UI; envelope is the contract.

### Decision 4 — Optional guidance for long analyses

Rationale: even with the fix, blocking a turn on a 10-minute synchronous `bash`
is wasteful. The guidance "run `gaze crap` in background and poll output
file" is documented for agents, not enforced by code. It reduces the window
during which Decision 2 must hold the idle timer.

Determinism classification (DE-001/DE-005): the note is applicability
judgment — when to background a long analysis — not copyable mechanics.
No registered launch-and-poll tool exists (closest: `uf_finale_ci_poll`,
which polls PR CI, not local processes), so the mechanics stay prose only
as conditional guidance. If a launch-and-poll tool is later registered,
this note MUST migrate to invoke it. Targets: `.opencode/commands/gaze.md`
and `.opencode/commands/gaze-fix.md` (the commands that shell out to
`gaze crap`); no skill mentions it.

Rejected alternative: mandating background execution — that would break
composability and is unenforceable.

## Risks / Trade-offs

- **A truly wedged tool could hold idle forever.** Bounded three ways: the
  fixed deferral cap `IDLE_DEFERRAL_CAP_MS` from the last fingerprint
  change, the tool's own timeout where observable, and wall-clock budgets
  (`DT-FR-001`). Wall-clock remains the ultimate backstop; spec explicitly
  states tool timeout is model-handled, not idle-handled.
- **Multiple concurrent running tools.** `runningTools` is a sorted set; any
  non-empty set defers idle up to the earliest remaining deadline (or the
  cap). The envelope reports the tool name of the running part with the
  lexicographically smallest `partId` — a deterministic rule, not "first
  or primary"; the full set is available in the probe sample for
  debugging.
- **Backwards compatibility of the envelope.** Adding optional `running_tool`
  is non-breaking; consumers that ignore it continue to work. No schema version
  bump required.
- **Test cost of 15-minute `bash` scenario.** Cannot run a literal 15-minute
  tool in CI. Verification uses a fake timer / stubbed `client.session.messages`
  and a synthetic long-running part to assert the idle watcher does not fire at
  `idle_timeout` while `runningTools` is non-empty, and does fire when empty
  after the window.
