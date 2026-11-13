## Context

`provenance.Stamp` records a ref name only. Ref names move, so
records cannot identify the source revision (proposal.md, Why).
Constitution alignment is assessed in proposal.md; Observable
Quality (III) motivates the change and Testability (IV) shapes
the coverage strategy below.

## Goals / Non-Goals

### Goals
- Resolve a ref to an immutable commit object name.
- Make provenance records carry that commit.
- Handle untrusted ref names safely.

### Non-Goals
- Remote refs or fetching: Lookup reads the local repository
  only.
- Updating `Stamp` callers: there are none outside the package.

## Decisions

**D1. Shell out to git rather than read refs directly.** Reading
`.git/refs` and `packed-refs` by hand would miss worktrees,
alternates, and SHA-256 repositories. `git rev-parse --verify`
is authoritative. Composability is unaffected: git is already a
project requirement.

**D2. Argument hardening.** The command is built with
`exec.Command` (argv form, no shell). The name follows
`--end-of-options`, so `--output=x` and similar are read as a
revision, not an option, and `^{commit}` makes rev-parse peel
tags and refuse non-commits. Requires git 2.24 or later
(2019).

**D3. Environment.** `GIT_*` variables are removed from the
child environment so `dir` alone selects the repository
(RR-FR-004), and `LC_ALL=C` keeps git's diagnostics stable. The
filter is a pure function, `gitEnv`, tested directly.

**D4. Output validation.** rev-parse output is parsed by the
pure function `parseObjectName`, which accepts exactly 40 or 64
lowercase hex digits (RR-FR-002), so the malformed-output path is
unit tested without a fake git.

**D5. Stamp fails closed.** `Stamp` returns an error rather than
a Record without a commit (AP-FR-002): partial provenance would
look valid to consumers.

## Coverage Strategy

- **Unit, `internal/refs`**: `parseObjectName` and `gitEnv` are
  pure and table tested, including every rejection branch.
  `Lookup` is tested against real repositories created in
  `t.TempDir()`: each spelling of a branch, an annotated tag, a
  SHA-256 repository, hostile names (asserting no side-effect
  file appears), an ambient `GIT_DIR`, and an unknown ref
  (asserting git's message is carried). Expected values come
  from `git rev-parse` run by the test itself, so tests check the
  resolved value, not just its shape.
- **Unit, `internal/provenance`**: Stamp's success, failure and
  JSON shape, against a temporary repository.
- **Subprocess tests** skip under `testing.Short()`.
- **Integration / e2e**: none; there is no caller outside the
  packages yet.
- **Target**: both packages are in the helper coverage scope
  (80 percent) in `coverage-gate.json`; no threshold changes.

## Risks / Trade-offs

- git older than 2.24 rejects `--end-of-options`; Lookup then
  fails for every ref (fail closed, with git's message).
- Each Lookup spawns a process. Stamp runs once per artifact, so
  the cost is negligible.
