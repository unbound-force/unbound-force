## Context

Speckit enforces `NNN-<name>` feature branches at every
pipeline step via `check_feature_branch()` in
`common.sh`. OpenSpec has no branch awareness. This
creates inconsistency in the development workflow.

## Goals / Non-Goals

### Goals
- OpenSpec changes use `opsx/<name>` branches
- Hard gate: error if not on correct branch
- Propose creates branch, apply validates it,
  archive returns to main
- Cobalt-crush validates branch for OpenSpec path

### Non-Goals
- Modifying the openspec CLI (third-party npm package)
- Adding bash scripts (keep it Markdown instruction-only)
- Changing speckit branch enforcement
- Making explore mode require a branch

## Decisions

**Branch naming**: `opsx/<change-name>`. Uses git's
namespace convention (like `feature/`, `fix/`). Does
not collide with speckit's `NNN-*` pattern. The `opsx/`
prefix makes OpenSpec branches instantly recognizable
in `git branch` output.

**Enforcement via instructions**: Since we cannot modify
the openspec CLI, enforcement is done in the Markdown
command/skill files that AI agents follow. The agent
runs `git rev-parse --abbrev-ref HEAD` and checks the
result before proceeding.

**Hard gate pattern**: The agent checks the branch and
stops with a clear error message if wrong. No `--force`
override (keep it simple).

**Branch lifecycle**:
1. `/opsx-propose` → `git checkout -b opsx/<name>`
2. `/opsx-apply` → validate `HEAD == opsx/<name>`
3. `/opsx-archive` → `git checkout main` after archive

## Risks / Trade-offs

**Risk**: AI agents may not perfectly follow branch
validation instructions. Mitigated by making the
instructions explicit and unambiguous.

**Trade-off**: No automated enforcement (unlike
speckit's bash scripts that `exit 1`). Acceptable
because OpenSpec is tactical -- the blast radius of
a branch mistake is smaller.
