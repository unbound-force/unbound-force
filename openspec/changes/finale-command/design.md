## Context

Finalizing a branch is the most repeated manual
workflow in this project. Every feature branch and
OpenSpec change ends with the same 7-step sequence.
We have already added branch-safety guardrails to the
OpenSpec and Speckit commands to prevent uncommitted
changes from being lost -- `/finale` is the positive
counterpart that makes the "right path" easy.

## Goals / Non-Goals

### Goals
- Single command to finalize any feature branch
- Auto-generate conventional commit messages from diff
- Auto-create PR with summary from branch commits
- Watch CI checks and stop on failure
- Rebase-merge and return to main
- Works with both Speckit (`NNN-*`) and OpenSpec
  (`opsx/*`) branches

### Non-Goals
- Running tests locally (that is `/review-council`)
- Cutting a release (separate operation)
- Archiving OpenSpec changes (that is `/opsx-archive`)
- Supporting squash-merge or merge-commit strategies
  (rebase only, per project convention)
- Interactive rebase or conflict resolution

## Decisions

### D1: Auto-stage all changes

`/finale` runs `git add .` to stage all changes. This
is intentional -- the command is the "wrap it all up"
action. Users who want selective staging should commit
manually before running `/finale`.

Exception: files that likely contain secrets (`.env`,
`credentials.json`, etc.) trigger a warning and
confirmation prompt before staging.

### D2: Commit message from diff analysis

The command analyzes `git diff --cached` to generate a
conventional commit message (`feat:`, `fix:`, `docs:`,
`chore:`). The user is shown the proposed message and
can approve or edit it. If `$ARGUMENTS` are provided,
they are used as a hint for the commit message.

### D3: PR title from commit history

The PR title is derived from the commit messages on the
branch (via `git log main..HEAD`). The PR body is a
generated summary of all commits, not just the latest.
If a PR already exists for the branch, creation is
skipped and the existing PR is used.

### D4: Rebase-merge only

The project uses rebase-merge exclusively (per Git &
Workflow conventions in AGENTS.md). The command runs
`gh pr merge --rebase --delete-branch`. This is not
configurable.

### D5: Stop on failure, don't guess

If CI checks fail or merge fails, the command stops
and reports the error with context. It does not
attempt to fix CI failures or resolve merge conflicts
-- those require human judgment.

### D6: Markdown command file only

This is a slash command definition (`.md` file under
`.opencode/command/`), not Go code. It follows the
same pattern as `/cobalt-crush`, `/review-council`,
and `/uf-init`. The file is tool-owned and
auto-updated by `uf init`.

## Risks / Trade-offs

### Risk: `git add .` may stage unintended files

If `.gitignore` is incomplete, `git add .` could stage
files that should not be committed (build artifacts,
editor configs, etc.).

**Mitigation**: The command checks for likely secret
files and warns. For other unintended files, the user
should maintain their `.gitignore`. This is the same
risk as running `git add .` manually.

### Risk: CI check timeout

`gh pr checks --watch` blocks until checks complete.
For long-running CI pipelines, this could take minutes.

**Mitigation**: The `--watch` flag has a built-in
polling interval and can be interrupted. The command
reports progress while waiting.

### Trade-off: No squash-merge option

The command always rebase-merges. Teams that prefer
squash-merge cannot use this command as-is.

**Acceptance**: This project exclusively uses rebase.
The trade-off is acceptable for the target audience.
