## Why

The `/muti-mind.sync-push` command immediately invokes
`go run cmd/mutimind/main.go sync-push` without
presenting what will be created or updated and without
asking for user confirmation. This command creates new
GitHub Issues or updates existing ones -- irreversible
external actions that modify state in a third-party
system under the user's authenticated account.

This is a T1 weakness (irreversible external action
without mandatory confirmation gate), identified in the
parent audit (issue #346, which tracks all T1
weaknesses). The `/review-pr` confirmation gate was the
first fix under that audit. The fix pattern is
established: show the user what will happen, then gate
on explicit confirmation via AskUserQuestion before
executing.

Fixes #349.

## What Changes

Add a confirmation gate to the `/muti-mind.sync-push`
command that:

1. Runs the Go backend in a preview/dry-run mode to
   determine what will be synced (items to create,
   items to update)
2. Presents a summary of the pending actions to the
   user
3. Uses the AskUserQuestion tool to require explicit
   confirmation before proceeding
4. Only invokes the actual sync-push if the user
   confirms

This requires changes in two layers:
- **Command layer**
  (`.opencode/commands/muti-mind.sync-push.md`): Add
  preview, summary, and confirmation steps before the
  bash invocation
- **Go backend** (`cmd/mutimind/main.go` +
  `internal/sync/sync.go`): Add a `--dry-run` flag
  that lists pending actions without executing them

## Capabilities

### New Capabilities
- `sync-push --dry-run`: Preview mode that reports
  what would be created/updated without executing any
  GitHub API calls
- `sync-push confirmation gate`: AskUserQuestion-based
  confirmation step in the command file before any
  external actions

### Modified Capabilities
- `/muti-mind.sync-push`: Now includes a mandatory
  preview + confirmation flow before executing sync
  operations

### Removed Capabilities
- None

## Impact

- **Files modified**:
  `.opencode/commands/muti-mind.sync-push.md`,
  `cmd/mutimind/main.go`, `internal/sync/sync.go`
- **Tests affected**: `cmd/mutimind/main_test.go`,
  `internal/sync/sync_test.go`
- **Scaffold registry**:
  `internal/scaffold/scaffold_test.go` (if the command
  file checksum changes)
- **User experience**: Users will now see a preview of
  sync actions and must confirm before GitHub Issues
  are created or updated. This adds one interaction
  step but prevents accidental or unintended issue
  creation.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change affects the human-agent interaction layer
(command confirmation flow), not inter-hero artifact
communication. The sync-push backend continues to
produce the same artifacts and outputs.

### II. Composability First

**Assessment**: PASS

The `--dry-run` flag is additive -- existing CLI usage
without the flag is unchanged. The confirmation gate
lives entirely in the command file (agent instruction
layer) and does not introduce dependencies between
heroes.

### III. Observable Quality

**Assessment**: PASS (with documented trade-off)

The dry-run mode produces structured, human-readable
output showing exactly what actions will be taken
(create vs. update, item IDs, issue numbers). This
improves observability by making the sync-push
operation transparent before execution. The dry-run
preview is an internal mechanism consumed by the
command file, not a hero artifact -- JSON format
support is deferred per the Constitution III trade-off
documented in design.md (D2).

### IV. Testability

**Assessment**: PASS

The dry-run mode is testable in isolation using the
existing `GHRunner` test stub (enhanced with call
tracking) -- no real GitHub API calls are needed. The
confirmation gate in the command file follows the
established AskUserQuestion pattern used in other
commands.

### V. Security by Default

**Assessment**: PASS

This change directly improves security posture by
preventing irreversible external actions without
explicit user consent. It closes the sync-push T1
weakness identified in the security audit. The
bidirectional `/muti-mind.sync` command's push path
remains ungated and is tracked separately.
