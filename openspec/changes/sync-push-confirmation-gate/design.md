## Context

The `/muti-mind.sync-push` command currently invokes
the Go backend directly without any preview or
confirmation step. The backend's `Syncer.Push()`
method iterates over backlog items and either creates
or updates GitHub Issues via the `gh` CLI. Once
executed, these actions are irreversible -- issues
cannot be "uncreated" and updates overwrite previous
state.

The parent audit (issue #346, which tracks all T1
weaknesses) identified this as a T1 weakness. The
`/review-pr` confirmation gate was the first fix under
that audit. The established pattern: preview what will
happen, then gate on explicit user confirmation via
AskUserQuestion before executing.

## Goals / Non-Goals

### Goals
- Add a `--dry-run` flag to the Go backend's
  sync-push command that reports pending actions
  without executing them
- Add a confirmation gate to the command file that
  uses AskUserQuestion before any GitHub API calls
- Follow the same confirmation pattern used in
  `/review-pr` and `/triage-issue` for consistency
- Maintain backward compatibility -- the Go CLI's
  default behavior (without `--dry-run`) remains
  unchanged

### Non-Goals
- Refactoring the entire sync subsystem or changing
  the sync-pull/sync-status commands
- Adding undo/rollback capabilities for sync-push
  operations
- Changing the Syncer interface or GHRunner abstraction
  beyond the dry-run addition
- Adding confirmation gates to other commands (those
  are tracked in separate issues)

**Residual risk**: The `/muti-mind.sync` command
(bidirectional sync) calls `Syncer.Sync()`, which
internally calls `Push("")`. This push path is not
gated by the confirmation flow added here, because the
gate lives in the command file layer. This is a known
residual T1 weakness tracked separately. This change
intentionally scopes to `sync-push` only to avoid
scope creep.

## Decisions

### D1: Dry-run via flag on the Go backend, not a separate command

**Decision**: Add a `--dry-run` boolean flag to the
existing `sync-push` command rather than creating a
separate `sync-push-preview` subcommand.

**Rationale**: A flag is simpler, follows Go CLI
conventions, and keeps the command surface small. The
dry-run logic reuses the same item resolution path
(single item or all items) and only diverges at the
point of `gh` invocation. This aligns with
Composability First -- no new commands to discover
or maintain.

### D2: Dry-run output as structured text, not JSON

**Decision**: The dry-run output uses structured text
with clear action labels (CREATE / UPDATE) and item
details, rather than JSON.

**Rationale**: The primary consumer is the agent command
file, which presents the output to the user in a
conversational context. Structured text is immediately
readable. If machine-parseable output is needed later,
a `--format json` flag can be added as a separate
enhancement. This satisfies Observable Quality -- the
output is clear and informative.

**Constitution III trade-off**: The dry-run preview is
an internal mechanism consumed by the command file, not
a hero artifact output. Constitution Principle III's
JSON format MUST rule applies to hero outputs consumed
by other heroes or external tooling. JSON format support
is deferred to a future enhancement if machine-parseable
dry-run output is needed. This trade-off is documented
per the Governance section.

### D3: Confirmation gate in the command file, not the Go backend

**Decision**: The AskUserQuestion confirmation gate
lives in the command markdown file
(`.opencode/commands/muti-mind.sync-push.md`), not in
the Go binary.

**Rationale**: The Go backend is a non-interactive CLI
tool invoked via `go run`. The AskUserQuestion tool is
an agent-level capability available only in the command
file context. Placing the gate in the command file
follows the same architectural pattern as `/review-pr`
and keeps the Go backend testable without UI
interaction concerns. This aligns with Testability --
the backend dry-run is testable with the existing
`GHRunner` stub (enhanced with call tracking).

### D4: Dry-run implemented via a DryRun field on Syncer

**Decision**: Add a `DryRun bool` field to the `Syncer`
struct. When set, `Push()` collects and reports pending
actions without calling `s.runner.Run()`.

**Rationale**: This avoids duplicating the item
resolution and classification logic. The same `Push()`
method handles both modes, branching only at the
execution point. The `DryRun` field is set by the
cobra command's flag binding, keeping the wiring
clean.

## Risks / Trade-offs

### R1: Dry-run accuracy depends on current state

The dry-run preview reflects the state at the time it
runs. If the backlog or GitHub Issues change between
preview and execution, the actual actions may differ.
This is acceptable because the time window is typically
seconds (user reviews and confirms immediately), and
the alternative (locking) would add significant
complexity for minimal benefit.

### R2: Additional user interaction step

The confirmation gate adds one interaction round-trip
to every sync-push invocation. This is an intentional
trade-off: the small friction cost is justified by
preventing accidental issue creation under the user's
authenticated GitHub account. The dry-run path involves
only local disk I/O with no network calls, making the
overhead minimal.

### R3: Command file depends on dry-run flag existing

The updated command file assumes the `--dry-run` flag
is available in the Go backend. If the command file is
updated before the Go backend, the preview step will
fail. The tasks are ordered to implement the Go backend
changes first to avoid this.

## Coverage Strategy

- Unit tests (task 2.1): Cover all dry-run code paths
  in `internal/sync/sync.go` including error paths
- Integration tests (task 2.2): Verify flag wiring in
  `cmd/mutimind/main.go` with output format assertions
- Target: 100% branch coverage for the new `DryRun`
  conditional in `Push()`
- Coverage ratchet: New code MUST NOT decrease existing
  coverage percentage
- Test infrastructure: `StubGHRunner` MUST be enhanced
  with call-tracking to verify zero `gh` invocations
  during dry-run
