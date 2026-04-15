# Implementation Plan: AGENTS.md Behavioral Guidance Injection

**Branch**: `030-agents-md-guidance` | **Date**: 2026-04-15 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/030-agents-md-guidance/spec.md`

## Summary

Add a new step to the `/uf-init` command that injects 8
standardized behavioral guidance sections into a repo's
`AGENTS.md`. The guidance blocks codify critical agent
behavioral rules — gatekeeping protection, workflow phase
boundaries, CI parity, review council prerequisites,
website documentation sync, spec-first development,
knowledge retrieval, and core mission — ensuring every
repo in the Unbound Force ecosystem enforces the same
quality gates and process discipline.

The implementation modifies exactly 2 files: the live
`/uf-init` command file and its scaffold asset copy.
No Go code, no tests, no agent files, no convention
packs. The 8 guidance block texts are defined inline
in the command instructions.

## Technical Context

**Language/Version**: Markdown (OpenCode command file)
**Primary Dependencies**: OpenCode agent runtime (renders
  Markdown instructions), existing `/uf-init` command
  (Steps 1-9)
**Storage**: N/A (modifies `AGENTS.md` in target repo at
  runtime; no persistent state)
**Testing**: Manual verification — run `/uf-init` in a
  repo and check `AGENTS.md` content. Existing scaffold
  drift detection tests verify the asset copy stays in
  sync.
**Target Platform**: Any Unbound Force repository with an
  `AGENTS.md` file
**Project Type**: Command file (Markdown instructions for
  AI agent execution)
**Performance Goals**: N/A
**Constraints**: Must be idempotent. Must not modify
  `AGENTS.md` files that already have the sections. Must
  skip gracefully when `AGENTS.md` does not exist.
**Scale/Scope**: 6 repos in the Unbound Force org (meta,
  gaze, website, dewey, homebrew-tap, replicator)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after
Phase 1 design.*

### I. Autonomous Collaboration — PASS

The guidance blocks are self-contained text injected into
`AGENTS.md`. No runtime coupling between repos. Each repo
gets its own copy of the guidance — no shared mutable
state, no cross-repo dependencies at injection time. The
`/uf-init` command reads the target repo's `AGENTS.md`
and writes to it locally.

### II. Composability First — PASS

The guidance injection step is additive — it does not
require any other hero to be installed. If `AGENTS.md`
does not exist, the step is skipped gracefully (FR-004).
The 8 blocks are independently useful: a repo can have
some but not all, and each block provides value on its
own. No mandatory dependencies introduced.

### III. Observable Quality — PASS

The `/uf-init` report summary (Step 9, extended by
FR-006) includes an "AGENTS.md Guidance" section showing
which blocks were injected, which were already present,
and which were skipped. This provides machine-readable
(structured summary) and human-readable (status
indicators) output for every injection action.

### IV. Testability — PASS

The implementation is Markdown-only — no Go code changes,
so no new functions to unit test. Testability is verified
through:
1. Existing scaffold drift detection tests ensure the
   asset copy matches the live file.
2. Manual acceptance testing: run `/uf-init` in a repo,
   verify sections appear, run again, verify idempotent.
3. The idempotency check (section heading detection) is
   deterministic and reproducible.
4. Content correctness test: add assertions to an
   existing scaffold test that verify the `/uf-init`
   command file contains all 8 detection phrases (one
   per guidance block). This catches typos, missing
   sections, or malformed content without introducing
   new test infrastructure.

**Coverage strategy**: No new Go code → no coverage
impact. Existing `TestScaffoldAssetDrift` continues to
verify the scaffold asset copy stays synchronized. A
lightweight content-presence test (string assertions
for 8 detection phrases in the scaffold asset) provides
automated verification that all guidance blocks are
defined in the command file.

## Project Structure

### Documentation (this feature)

```text
specs/030-agents-md-guidance/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0: canonical text sources
├── data-model.md        # Phase 1: guidance block schema
├── quickstart.md        # Phase 1: implementation guide
├── checklists/
│   └── requirements.md  # Pre-existing checklist
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
.opencode/command/
└── uf-init.md                              # Live command (add Step 10)

internal/scaffold/assets/opencode/command/
└── uf-init.md                              # Scaffold asset copy (sync)
```

**Structure Decision**: This feature modifies exactly 2
files — the live `/uf-init` command and its scaffold
asset copy. Both are Markdown. No new directories, no
new packages, no new Go source files.

## Implementation Approach

### Step 10 Design

The new step is inserted after Step 8 (OpenSpec Command
Guardrails) and before Step 9 (Report Results). The
existing Step 9 is renumbered to Step 10, and the new
AGENTS.md guidance injection becomes Step 9. The report
in the new Step 10 is extended with an "AGENTS.md
Guidance" section.

**Alternative considered**: Inserting as a sub-step of
an existing step. Rejected because the AGENTS.md
guidance injection is conceptually distinct from OpenSpec
or Speckit customizations — it operates on a different
target file (`AGENTS.md` vs. skill/command files).

### Idempotency Strategy

Each guidance block has a **detection heading** — a
specific Markdown heading or key phrase that the AI agent
searches for in `AGENTS.md`. If the heading (or semantic
equivalent) is found, the block is skipped. This mirrors
the idempotency pattern already used in Steps 2-4 of
`/uf-init`.

| Block | Detection Heading |
|-------|-------------------|
| Core Mission | `## Core Mission` |
| Gatekeeping Value Protection | `### Gatekeeping Value Protection` |
| Workflow Phase Boundaries | `### Workflow Phase Boundaries` |
| CI Parity Gate | `CI Parity Gate` |
| Review Council PR Prerequisite | `Review Council` + `PR Prerequisite` |
| Website Documentation Sync Gate | `Website Documentation` + `Gate` |
| Spec-First Development | `Spec-First Development` |
| Knowledge Retrieval | `## Knowledge Retrieval` |

### Canonical Text Sources

The 8 guidance blocks are derived from two canonical
sources:

1. **Meta repo AGENTS.md** (`unbound-force/unbound-force`):
   - Core Mission (lines 12-16)
   - Gatekeeping Value Protection (lines 25-38)
   - Workflow Phase Boundaries (lines 40-54)
   - Website Documentation Sync Gate (lines 397-418)
   - Knowledge Retrieval (lines 533-592)

2. **Gaze AGENTS.md** (`unbound-force/gaze`):
   - CI Parity Gate (line 27)
   - Review Council PR Prerequisite (lines 38-54)
   - Spec-First Development (lines 56-82)
   - Core Mission (lines 12-16, identical to meta repo)

The inline text in `/uf-init` is a **generalized version**
of these canonical sources — repo-specific details
(hero names, specific constitution principles) are
replaced with generic placeholders that apply to any
repo. See `research.md` for the full canonical text
extraction and generalization analysis.

### Report Extension

The existing Step 9 report template gains a new section:

```
### AGENTS.md Guidance
  [status] Core Mission: [action]
  [status] Gatekeeping Value Protection: [action]
  [status] Workflow Phase Boundaries: [action]
  [status] CI Parity Gate: [action]
  [status] Review Council PR Prerequisite: [action]
  [status] Website Documentation Sync Gate: [action]
  [status] Spec-First Development: [action]
  [status] Knowledge Retrieval: [action]
```

Using the same status indicators as existing sections:
`✅` (inserted), `⊘` (already present), `❌` (error).

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| AI agent misidentifies section as present when it's not | Low | Medium | Use specific heading text, not vague keywords |
| AI agent inserts at wrong location in AGENTS.md | Low | Low | Provide explicit placement guidance (after which section) |
| Scaffold asset copy drifts from live file | Medium | Low | Existing `TestScaffoldAssetDrift` catches this |
| Guidance text is too meta-repo-specific | Medium | Medium | Generalize all text; remove hero names, specific principles |
| Large AGENTS.md causes context window issues | Low | Medium | Guidance blocks are self-contained; agent reads headings first |

## Complexity Tracking

> No constitution violations to justify. All four
> principles pass cleanly.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (none) | — | — |
