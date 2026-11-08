## Context

AGENTS.md commit-scope rules require Dewey learnings and
compiled articles to be included in feature PRs. The
`/uf.finale` command template currently stages all changes
with `git add .` but has no explicit instruction to verify
Dewey artifacts are included. Because `.uf/dewey/` files
may be created by Dewey MCP tools outside the normal
working-tree flow, agents could miss them.

The proposal (constitution alignment: PASS on Autonomous
Collaboration and Observable Quality) calls for an
explicit staging sub-step in `/uf.finale` Step 2.

## Goals / Non-Goals

### Goals
- Ensure Dewey knowledge artifacts are always staged
  during `/uf.finale`
- Make the instruction idempotent (no-op when no Dewey
  files exist)
- Keep the change minimal -- a single sub-step addition

### Non-Goals
- Validating Dewey file content or format
- Adding Dewey-specific commit message conventions
- Changing any user-facing gates or approval flows
- Modifying the scaffold drift-detection test (file
  count does not change since no files are added or
  removed from the scaffold)

## Decisions

**D1: Place the instruction after `git add .`**

The `git add .` in Step 2 already stages the working tree.
The Dewey step follows as a belt-and-suspenders measure:
an explicit `git add .uf/dewey/learnings/ .uf/dewey/compiled/`
ensures these directories are staged even if a prior
step or agent behavior excluded them. Placing it after
`git add .` means it acts as a verification/reinforcement
rather than the primary staging mechanism.

**D2: Use `2>/dev/null` for idempotency**

The `git add` targets directories that may not exist in
every repository. Suppressing stderr avoids noisy warnings
when the directories are absent.

**D3: Include rationale in the instruction text**

The instruction explains why Dewey artifacts must be
committed so agents understand the step's purpose and
do not skip it as redundant with `git add .`.

## Risks / Trade-offs

- **Risk**: Agents may treat the explicit `git add` as
  redundant after `git add .`. Mitigated by the "Do NOT
  skip this step" directive in the instruction text.
- **Trade-off**: Adds ~12 lines to the template. Accepted
  because the cost of lost knowledge artifacts outweighs
  the minor template growth.
