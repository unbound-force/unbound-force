# idle-timeout-aborts-children-waiting-on — Change Proposal

## Why

`delegation-timeout-budgets` (#950) introduced an `idle_timeout` (default 600,000 ms)
that probes every `idle_poll_interval` for progress via the fingerprint of the
latest message per session in the delegation subtree. A `bash` tool call that
remains `running` for minutes — e.g., `gaze crap` or `go test -race` with `gaze
crap` — creates no new messages until it completes. The probe therefore sees an
unchanged fingerprint for the entire window and emits a false idle signal
(`kind: idle`, `child_active: false`, `last_progress_ms_ago: 600000`).

Evidence: five consecutive `/uf.unleash` delegations, four during `bash` (`gaze
crap` 10m07s, `bash` timing out after 600s then second `gaze crap` via fresh
invoke, variant that aborts) and one true stall (`read.filePath` on a
non-existent tool, `dispatch_review_runs` never returned). Long-running tool
work is being mistaken for model stalls, breaking otherwise healthy review
loops. When a running tool hits its own timeout that is a tool failure for the
model to handle — it MUST NOT become an idle abort.

## What Changes

- Progress fingerprint for each polled session MUST include running tool parts
  as `(partId, status, tool)` triples sorted lexicographically by `partId`
  (all three fields; `tool` empty string when the part carries none). A `part`
  with `state.status === "running"` anywhere in the subtree counts as
  activity.
- A running tool invocation defers the idle abort until a bounded deferral:
  the tool's own remaining timeout plus grace when observable, else a fixed
  deferral cap (5-15 minutes) measured from the last fingerprint change.
  Idle is never held later than `idle_timeout` + bound, and wall-clock
  remains the absolute backstop. Tool timeout is the model-handled
  boundary; idle is the model-stall boundary.
- The abort envelope (`TimeoutKind`, `last_progress_ms_ago`, `child_active`)
  gains a `running_tool` discriminator (e.g., `"bash"`) so the operator can
  distinguish "tool was still running" from "nothing was happening".
- Optional guidance is added for long analyses that currently shell out
  synchronously: run `gaze crap` (or similar) in the background and poll an
  output file, rather than blocking the turn. Documented in
  `.opencode/commands/gaze.md` and `.opencode/commands/gaze-fix.md` as
  applicability judgment, not copyable mechanics (design Decision 4). This
  pattern is documented, not enforced.

## Capabilities

### New Capabilities
- None. This is a reliability fix to an existing capability.

### Modified Capabilities
- `delegation-timeouts`: idle detection, wall-clock budgeting, and probe
  reporting are amended to be tool-aware. New invariants DT-FR-013 through
  DT-FR-015 cover the running-tool fingerprint, bounded deferral, and abort
  envelope. DT-FR-004's `kind` enumeration is aligned with the implemented
  `TimeoutKind` (which includes `storm` since the compaction-storm fix).

### Removed Capabilities
- None.

## Impact

- Affected: `.opencode/lib/delegation-budget.ts` (`probeSubtreeActivity`,
  `messageFingerprint`, `probeContribution`, `ProgressSample` shape), `.opencode/lib/agent-executor.ts`
  (`TimeoutKind`, idle watcher, abort envelope), and instruction files that
  document the `DELEGATE_FAILED` timeout envelope (task 2.4).

### Instruction-text deltas (DE-004)

- `.opencode/commands/gaze.md` ~+8 lines and
  `.opencode/commands/gaze-fix.md` ~+8 lines (task 3.2, optional
  background-and-poll guidance).
- `.opencode/commands/uf.run-pipeline-plan.md` ~+4 lines (task 2.4,
  `running_tool` added to the enumerated `timeout` fields).
- Other command and skill files that enumerate `timeout` object fields or
  `dispatch_review_runs` handling gain +0-6 lines each, only where they
  enumerate (task 2.4); generic references are unchanged.
- Total instruction-text growth ≤ +30 lines; the implementer records actual
  deltas in the PR.
- Risk: low. Purely expands what counts as progress; wall-clock budget remains
  the backstop against genuinely stuck tools. No new runtime dependencies.
- Out of scope: `docs/`, `blog/`, tutorial, and OpenCode v2 migration
  (tracked separately). Spec writing keeps RFC 2119 + Given/When/Then.

## Constitution Alignment

Assessed against the Unbound Force org constitution
(`.specify/memory/constitution.md`).

### I. Autonomous Collaboration

**Assessment**: PASS

Artifacts remain the collaboration boundary. The change corrects the progress
artifact (fingerprint + abort envelope) so independent review subagents are not
killed while doing tool work. No runtime coupling is introduced; the hierarchy
budget table is unchanged.

### II. Composability First

**Assessment**: PASS

Every hero stays independently installable. The fix is scoped to the delegation
budget/idle watcher; heroes that do not delegate are unaffected. No mandatory
dependency is added. The background-poll guidance is optional and composable
with any long-running analysis.

### III. Observable Quality

**Assessment**: PASS

Improves observability: the abort envelope becomes machine-parseable for the
distinction between tool-running and true stalls (`running_tool`). Probe
samples retain provenance (`sessionId`, `child_active`, `last_progress_ms_ago`)
and add the discriminator without breaking existing consumers.

### IV. Testability

**Assessment**: PASS

Observable side effects (fingerprint includes running parts, envelope includes
`running_tool`, idle abort deferred) are unit-testable in isolation via
`delegation-budget` and `agent-executor` tests without external services.
Wall-clock vs idle vs tool-timeout boundaries are separately exercised.

### V. Security by Default

**Assessment**: PASS

No secrets, authz, or trust-boundary change. Tool execution and timeout
handling remain model-mediated; the abort envelope carries only the tool
name, not payload. The name originates from model-controlled message
parts, so DT-FR-015 mandates charset/length validation
(`^[A-Za-z0-9._:-]{1,128}$`) with `null`-on-failure and no verbatim
logging of invalid values — adversarial tool names cannot inject into
envelopes or logs.

### VI. Deterministic Execution

**Assessment**: PASS

No LLM judgment is introduced. Idle vs progressing is a syntactic check over
`(partId, status, tool, streamedLength)` and the running-tool set. Outcomes are
reproducible for a given session trace.
