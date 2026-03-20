# Implementation Plan: Cobalt-Crush Architecture (Developer)

**Branch**: `006-cobalt-crush-architecture` | **Date**: 2026-03-20 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/006-cobalt-crush-architecture/spec.md`

## Summary

Cobalt-Crush is the Developer hero — an AI agent persona that
defines coding behavior, convention adherence, and feedback loop
integration for the `/speckit.implement` workflow. The entire
implementation is a single Markdown agent file
(`cobalt-crush-dev.md`) deployed via `unbound init`.

A prerequisite refactor moves convention packs from
`.opencode/divisor/packs/` to `.opencode/unbound/packs/` — a
neutral, org-level location shared by both Cobalt-Crush and
The Divisor. This refactor touches Go source code, tests,
embedded assets, agent files, and documentation.

## Technical Context

**Language/Version**: Markdown (agent file), Go 1.24+ (scaffold engine refactor)
**Primary Dependencies**: `embed.FS` (asset embedding), existing scaffold engine
**Storage**: Filesystem only (Markdown files deployed to target directory)
**Testing**: Standard library `testing` package, `-race -count=1`, drift detection
**Target Platform**: macOS, Linux (cross-compiled via GoReleaser)
**Project Type**: CLI tool (extending existing `unbound` binary)
**Performance Goals**: N/A (one-shot scaffold; agent execution is OpenCode-bound)
**Constraints**: No external network calls during scaffold
**Scale/Scope**: 1 new agent file + prerequisite refactor (~97 path references across 20 files)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Autonomous Collaboration — PASS

- Cobalt-Crush is a standalone agent file with no runtime
  coupling to other heroes.
- Feedback loops (Gaze, Divisor) are instruction-based: the
  agent reads artifact files from `.unbound-force/artifacts/`.
- Convention packs are shared files, not runtime dependencies.
- graphthulhu MCP integration is a SHOULD enhancement with
  file-based fallback.

### II. Composability First — PASS

- Cobalt-Crush functions without Gaze or The Divisor
  (SC-007). Missing heroes trigger graceful degradation
  notes, not errors.
- Convention packs at `.opencode/unbound/packs/` are shared
  extension points — same files used by Divisor.
- `unbound init` deploys Cobalt-Crush alongside everything;
  no separate install step.

### III. Observable Quality — PASS

- Cobalt-Crush produces code that is validated by Gaze
  (machine-parseable quality reports) and The Divisor
  (structured review verdicts).
- The agent instructs the LLM to document design decisions
  and cite conventions — creating traceable provenance.
- JSON artifact output deferred to Spec 009 (same as Divisor).

### IV. Testability — PASS

- **Coverage strategy**: Drift detection tests verify the
  embedded `cobalt-crush-dev.md` matches the canonical source.
  Scaffold integration tests verify deployment. No new Go
  functions beyond the prerequisite refactor (which updates
  existing tested functions).
- **Isolation**: All tests use `t.TempDir()`. No shared state.
- The prerequisite refactor changes path strings in existing
  tested functions — tests are updated in lockstep.

**Gate Result**: ALL FOUR PRINCIPLES PASS. Proceed to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/006-cobalt-crush-architecture/
├── spec.md              # Feature specification (clarified)
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
# Prerequisite refactor: convention pack relocation
.opencode/unbound/packs/                    # NEW location (moved from .opencode/divisor/packs/)
├── go.md                                   # MOVED from .opencode/divisor/packs/go.md
├── go-custom.md                            # MOVED
├── default.md                              # MOVED
├── default-custom.md                       # MOVED
├── typescript.md                           # MOVED
└── typescript-custom.md                    # MOVED

internal/scaffold/assets/opencode/unbound/packs/  # NEW embedded location
├── go.md                                   # MOVED from .../divisor/packs/
├── go-custom.md                            # MOVED
├── default.md                              # MOVED
├── default-custom.md                       # MOVED
├── typescript.md                           # MOVED
└── typescript-custom.md                    # MOVED

# New Cobalt-Crush agent file
.opencode/agents/cobalt-crush-dev.md        # NEW canonical source
internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md  # NEW embedded copy

# Modified files (prerequisite refactor)
internal/scaffold/scaffold.go               # MODIFIED: isConventionPack() path, isDivisorAsset() logic
internal/scaffold/scaffold_test.go          # MODIFIED: 35 path references
.opencode/agents/divisor-guard.md           # MODIFIED: pack path reference
.opencode/agents/divisor-architect.md       # MODIFIED: pack path reference
.opencode/agents/divisor-adversary.md       # MODIFIED: pack path references (3)
.opencode/agents/divisor-sre.md             # MODIFIED: pack path reference
.opencode/agents/divisor-testing.md         # MODIFIED: pack path reference
internal/scaffold/assets/opencode/agents/divisor-*.md  # MODIFIED: sync with canonical
AGENTS.md                                   # MODIFIED: project structure tree
```

**Structure Decision**: No new Go packages. The prerequisite
refactor changes path strings in existing code. The Cobalt-Crush
agent is a single Markdown file added to the existing
`opencode/agents/` directory alongside the Divisor agents.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
<!-- scaffolded by unbound vdev -->
