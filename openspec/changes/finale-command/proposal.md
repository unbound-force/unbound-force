## Why

Finalizing a branch requires a repetitive multi-step
manual workflow: stage changes, generate commit message,
commit, push, create PR, watch CI checks, merge, switch
to main, pull. Every feature branch and OpenSpec change
ends with these same steps, and forgetting one (e.g.,
not pushing before archiving) causes problems.

The `/finale` command automates this entire end-of-branch
workflow into a single invocation.

## What Changes

A new OpenCode slash command `/finale` that performs the
full branch-finalization workflow:

1. Validates not on `main` (branch safety)
2. Stages and commits all changes with an auto-generated
   conventional commit message
3. Pushes to remote
4. Creates a PR (or uses existing one)
5. Watches CI checks until pass/fail
6. Merges the PR via rebase
7. Switches to `main` and pulls

## Capabilities

### New Capabilities
- `finale-command`: Single slash command to finalize a
  branch -- commit, push, PR, merge, and return to main

### Modified Capabilities
- None

### Removed Capabilities
- None

## Impact

- **New file**: `.opencode/command/finale.md` (the
  command definition, tool-owned)
- **New scaffold asset**:
  `internal/scaffold/assets/opencode/command/finale.md`
- **Modified test**: `cmd/unbound-force/main_test.go`
  (expected file count 50 → 51)
- No Go library code changes
- No behavioral changes to existing commands
- Scaffolded into all repos on next `uf init`

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This is an OpenCode slash command (agent instruction
file). It does not affect inter-hero artifact
communication or introduce runtime coupling.

### II. Composability First

**Assessment**: PASS

The command is independently usable -- it works on any
branch in any repo with a git remote and GitHub CLI.
It does not require any other hero to be installed.

### III. Observable Quality

**Assessment**: N/A

This is an agent workflow command, not a data-producing
hero. It does not generate machine-parseable artifacts.

### IV. Testability

**Assessment**: PASS

The scaffold file count is verified by an existing test
(`cmd/unbound-force/main_test.go`). The command itself
is a Markdown instruction file with no code to unit
test. The workflow it describes uses standard git and
gh CLI commands that can be verified by their output.
