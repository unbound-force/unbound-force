## Context

The `openspec-archive-change` skill guides agents through
archiving a completed OpenSpec change. The current skill has
two structural gaps where irreversible operations lack
explicit user confirmation gates:

1. The transition from step 5 (commit/push) to step 6
   (archive) relies on a prose-only "CRITICAL" warning
   that agents with compressed session context may skip.
2. Step 7 executes `git checkout main` without any
   confirmation gate.

Issue #346 root cause analysis established that agents
bypass prose-only guards during session compression. The
fix pattern -- adding `AskUserQuestion` gates before
irreversible operations -- is proven and already applied
in issue #356 for the related `opsx-archive.md` command.

## Goals / Non-Goals

### Goals

- Add an `AskUserQuestion` gate before step 6 (archive)
  that confirms all changes are committed and pushed
- Add an `AskUserQuestion` gate before `git checkout main`
  in step 7 that confirms the user wants to switch branches
- Follow the same gate pattern used in issue #356 for
  consistency across archive workflows

### Non-Goals

- Refactoring the overall skill structure or step ordering
- Adding gates to other steps (steps 1-4 already use
  `AskUserQuestion` where appropriate)
- Modifying any other skill files
- Adding automated tests for skill file behavior

## Decisions

### D1: Gate placement -- between steps, not within

The commit confirmation gate will be placed between step 5
and step 6, not within step 5 itself. This ensures the
gate fires after commit verification is complete but before
the archive operation begins. The gate acts as a checkpoint
boundary, not an inline check.

**Rationale**: The existing step 5 already handles the
commit/push logic. The gate is a transition guard -- it
belongs at the boundary between "ensure committed" and
"perform archive."

### D2: Gate option text mirrors issue #360 specification

The `AskUserQuestion` options will use the exact text
specified in issue #360:

- Gap A: `["Changes committed and pushed -- proceed to archive", "Abort -- need to commit first"]`
- Gap B: `["Return to main", "Stay on branch"]`

**Rationale**: The issue author specified these options
after analyzing the gap pattern. Using the exact specified
text maintains traceability and avoids unnecessary
divergence.

### D3: Abort behavior -- stop and inform, do not retry

When the user selects the abort option at either gate:
- Gap A abort: Stop the archive workflow and inform the user
  to commit changes before retrying.
- Gap B stay: Skip the branch switch and inform the user
  they remain on the `opsx/<name>` branch.

**Rationale**: The agent should not attempt to fix the
situation (e.g., auto-committing). The gates exist to
give the user control over irreversible operations.

### D4: No session-resume guard needed

Unlike the `/review-pr` fix in issue #346, this change
does not need a separate session-resume guard. The
`AskUserQuestion` tool is inherently session-bound -- it
must execute fresh in every session because it requires
real-time user input. The gap was that no question was
asked at all, not that an existing question was skipped
during resume.

## Risks / Trade-offs

### R1: Additional friction in the archive workflow

Adding two confirmation gates introduces two extra user
interactions. This is acceptable because:
- Archive is a low-frequency operation (once per change)
- Both gates protect against irreversible operations
- The cost of a misarchived change or lost uncommitted
  work far exceeds the cost of two confirmation clicks

### R2: Consistency with opsx-archive.md (issue #356)

The `opsx-archive.md` command file has the same Gap B
pattern being fixed in issue #356. Both fixes should
use identical gate text and placement for consistency.
This change uses the same "Return to main" / "Stay on
branch" options specified in issue #356.
