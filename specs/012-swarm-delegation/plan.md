# Implementation Plan: Swarm Delegation Workflow

**Branch**: `012-swarm-delegation` | **Date**: 2026-03-22 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/012-swarm-delegation/spec.md`

## Summary

Add execution mode awareness to the hero lifecycle
workflow so the swarm can run autonomously from
implementation through review, pausing at a human
checkpoint before acceptance. Rename the final "measure" stage to
"reflect" -- a richer retrospective that includes
metrics collection, cross-hero learning analysis, and
empirical data from all hero artifacts.

Two interconnected changes to `internal/orchestration/`:

1. **Execution modes**: Each `WorkflowStage` gains an
   `ExecutionMode` field (`"human"` or `"swarm"`).
   `Advance()` detects swarm-to-human transitions and
   pauses the workflow at `StatusAwaitingHuman`. Calling
   `Advance()` again resumes from the checkpoint.

2. **Reflect stage**: Rename `StageMeasure` to
   `StageReflect`. The stage is still owned by Mx F but
   the SKILL.md documentation instructs the swarm to
   consume quality report and review verdict artifacts
   and run `AnalyzeWorkflows` for learning feedback.

## Technical Context

**Language/Version**: Go 1.24
**Primary Dependencies**: `github.com/spf13/cobra`
(CLI), `github.com/charmbracelet/log` (logging),
`github.com/charmbracelet/lipgloss` (terminal styling)
**Storage**: JSON files at
`.unbound-force/workflows/{id}.json` (existing)
**Testing**: Standard library `testing` package,
`go test -race -count=1`
**Target Platform**: macOS, Linux (CLI binary)
**Project Type**: CLI
**Performance Goals**: N/A (state machine transitions,
not throughput-sensitive)
**Constraints**: Backward compatible with existing
workflow JSON files
**Scale/Scope**: 7 Go files + 5 Markdown files changed,
~200 lines of new Go code, ~250 lines of new tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check
after Phase 1 design.*

### I. Autonomous Collaboration — PASS

The execution mode and checkpoint mechanism communicate
through artifacts (JSON workflow files) not runtime
coupling. The `StatusAwaitingHuman` state is persisted
to disk -- the swarm and human operator never interact
synchronously. Each hero still completes its stage
independently and produces self-describing artifacts.

### II. Composability First — PASS

Execution modes default to `"human"` when absent
(FR-010), so the feature is opt-in. A workflow without
swarm infrastructure advances exactly as before. No hero
requires another hero to be present for its stage to
work. The reflect stage enriches output when Gaze and
Divisor artifacts are available, but functions without
them (produces metrics snapshot alone).

### III. Observable Quality — PASS

The `ExecutionMode` field is persisted in the JSON
workflow file, making the delegation state
machine-parseable. The reflect stage produces structured
artifacts (metrics snapshot, learning feedback) in the
standard envelope format with provenance metadata.

### IV. Testability — PASS

All changes are in `internal/orchestration/` which uses
injected dependencies (temp dirs, stubbed clocks,
mock `LookPath`). New tests cover:
- Checkpoint pause at swarm-to-human boundaries
- Resume from `awaiting_human` status
- Swarm-to-swarm transitions (no pause)
- Execution mode population on new workflows
- Backward compatibility for legacy JSON
- `Latest()` discovery of `awaiting_human` workflows

**Coverage strategy**: Unit tests using standard library
`testing` package. All filesystem operations use
`t.TempDir()`. Clock and binary lookup are injected.
Target: maintain existing ~87% coverage for the
orchestration package. No integration tests needed --
the orchestration engine is a pure state machine with
no external service dependencies.

## Project Structure

### Documentation (this feature)

```text
specs/012-swarm-delegation/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── workflow-commands.md
└── checklists/
    └── requirements.md  # Already created
```

### Source Code (repository root)

```text
internal/orchestration/
├── models.go            # Stage/status constants, WorkflowStage struct
├── heroes.go            # StageHeroMap, StageExecutionModeMap (new)
├── engine.go            # Advance checkpoint logic, NewWorkflow mode population
├── record.go            # No changes expected
├── store.go             # Latest() updated for awaiting_human
├── learning.go          # No changes expected
├── engine_test.go       # New checkpoint tests, updated assertions
├── heroes_test.go       # Updated stage assertions
└── store_test.go        # New Latest() test for awaiting_human

.opencode/
├── skill/
│   └── unbound-force-heroes/
│       └── SKILL.md     # Updated stages, delegation docs
└── command/
    ├── workflow-start.md    # Updated output format
    ├── workflow-advance.md  # Checkpoint behavior docs
    └── workflow-status.md   # Awaiting_human indicator

AGENTS.md                # Updated Recent Changes entry
```

**Structure Decision**: All changes are within the existing
`internal/orchestration/` package and `.opencode/` Markdown
files. No new packages, directories, or binaries. The
scaffold asset directory is not affected (the heroes
SKILL.md is a local-only file, not a scaffolded asset).

## Complexity Tracking

No constitution violations. No complexity justifications needed.
