## Context

A documentation audit (2026-03-29) identified
inaccuracies across README.md, AGENTS.md, and spec
frontmatter. These are living documents consumed by
both human contributors and AI agents. AGENTS.md in
particular is injected into every OpenCode agent's
context window, so stale information directly degrades
agent decision quality.

The audit was performed using Dewey MCP semantic search
and full-text search, cross-referenced against the
actual codebase state.

## Goals / Non-Goals

### Goals
- Fix all HIGH severity documentation inaccuracies
- Fix all MEDIUM severity inaccuracies
- Update spec frontmatter to reflect actual completion
  status
- Ensure AGENTS.md accurately represents the current
  project state for AI agent consumption

### Non-Goals
- Restructuring AGENTS.md (consolidating duplicated
  Active Technologies entries is desirable but out of
  scope for this change)
- Adding new documentation pages (website gaps are
  tracked separately in
  `../website-documentation-gaps.md`)
- Modifying completed spec content (only frontmatter
  `status` field changes)
- Code changes of any kind

## Decisions

### D1: Edit completed spec frontmatter only

Completed specs (001-011) are treated as historical
records and MUST NOT be modified. Specs 012-016 had
their implementation completed but the frontmatter
`status` field was never updated. Changing `status:
draft` to `status: complete` is a metadata correction,
not a content modification. The spec body text is
unchanged.

### D2: Preserve historical descriptions in Recent Changes

The "Recent Changes" section in AGENTS.md contains
historical summaries written at the time each spec was
completed. These are not updated retroactively even if
they reference outdated counts (e.g., "three principles"
when the constitution later added a fourth). The fix
applies only to current-state descriptions in the main
sections.

### D3: Update counts to current values

- Spec count: 10 → 16 (README.md)
- Scaffold file count: 47 → 50 (README.md)
- These numbers may change again with future work.
  The fix captures the current state.

### D4: Dewey replaces graphthulhu in README

README.md's Knowledge Graph section is rewritten to
reference Dewey. The graphthulhu reference is removed
entirely from README (unlike the website, which keeps
intentional historical context about the fork). README
should describe current tooling, not history.

## Risks / Trade-offs

### Risk: Scaffold file count drift

The README states a specific scaffold file count (50).
This number changes whenever scaffold assets are added
or removed. The test suite enforces the count
(`TestRun_CreatesFiles` asserts the expected number),
so any future drift will be caught by tests -- but the
README number may again go stale.

**Mitigation**: Accept this risk. The alternative
(removing the count from README) loses useful
information for users.

### Risk: Phase descriptions become stale again

The AGENTS.md spec phase descriptions reference
specific details (e.g., "Three-persona review
protocol" for Divisor). These may become stale as the
project evolves.

**Mitigation**: The Documentation Validation Gate in
AGENTS.md requires doc impact assessment before
marking any task complete. This is a process control,
not a technical one.
