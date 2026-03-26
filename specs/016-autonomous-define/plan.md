# Implementation Plan: Autonomous Define with Dewey

**Branch**: `016-autonomous-define` | **Date**: 2026-03-26 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/016-autonomous-define/spec.md`

## Summary

Make the define stage's execution mode configurable so
it can run as `[swarm]` instead of `[human]`. When in
swarm mode, Muti-Mind autonomously drafts specifications
using Dewey's cross-repo semantic context, resolves
ambiguities by querying Dewey instead of asking the
human, and references historical learning feedback.
An optional spec review checkpoint allows a lightweight
human gate between define and implement for high-stakes
features. A seed command reduces the operator's role to
one sentence of intent plus one acceptance decision.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: `github.com/spf13/cobra`
(CLI), `github.com/charmbracelet/log` (logging)
**Storage**: JSON workflow files at
`.unbound-force/workflows/{id}.json` (existing)
**Testing**: Standard library `testing` package,
`go test -race -count=1`
**Target Platform**: macOS, Linux (CLI binary)
**Project Type**: CLI
**Performance Goals**: N/A (orchestration state machine)
**Constraints**: Backward compatible -- default
execution mode map must produce identical behavior to
pre-change workflows. No new workflow status constants
(reuses `StatusAwaitingHuman` from Spec 012).
**Scale/Scope**: ~5 Go files changed in
`internal/orchestration/`, ~3 Markdown command/skill
files updated, ~200 lines of new Go code, ~150 lines
of new tests.

## Constitution Check

### I. Autonomous Collaboration -- PASS

The configurable execution mode map is persisted in the
workflow JSON file. The seed command creates a backlog
item (artifact) and starts a workflow (artifact). No
runtime coupling introduced. Muti-Mind's autonomous
specification is documented in its agent file (Markdown
artifact).

### II. Composability First -- PASS

The define stage defaults to `human` mode (FR-002).
Changing it to `swarm` is opt-in. The spec review
checkpoint is optional and disabled by default (FR-013).
No hero requires the define stage to be in swarm mode
for its core function. Dewey degradation (Tier 1
fallback) is already in Muti-Mind's agent file.

### III. Observable Quality -- PASS

The execution mode configuration is persisted in the
workflow JSON (machine-parseable). The `/workflow status`
command shows the execution mode per stage (already
implemented in Spec 012). The seed command produces the
same workflow JSON as `/workflow start`.

### IV. Testability -- PASS

The `StageExecutionModeMap()` function currently returns
hardcoded values. Making it configurable means the
function accepts parameters or reads from a config
source -- both are injectable for testing. The existing
`TestOrchestrator_NewWorkflow_SetsExecutionModes` test
validates mode population. New tests will cover
configurable modes, the spec review checkpoint, and
the seed command.

**Coverage strategy**: Unit tests using the existing
DI pattern (temp dirs, stubbed clocks, mock LookPath).
Target: maintain existing ~87% coverage for the
orchestration package.

## Project Structure

### Documentation (this feature)

```text
specs/016-autonomous-define/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── contracts/           # Phase 1 output
│   └── workflow-seed.md
├── quickstart.md        # Phase 1 output
└── checklists/
    └── requirements.md  # Already created
```

### Source Code (repository root)

```text
internal/orchestration/
├── models.go            # ExecutionModeConfig type
├── heroes.go            # StageExecutionModeMap() accepts config
├── engine.go            # Configurable mode in NewWorkflow(),
│                        # spec review checkpoint in Advance()
└── engine_test.go       # New tests for configurable modes

.opencode/
├── agents/
│   └── muti-mind-po.md  # Autonomous spec workflow instructions
├── skill/
│   └── unbound-force-heroes/
│       └── SKILL.md     # Seed workflow, configurable define
└── command/
    ├── workflow-start.md # Updated for seed parameter
    └── workflow-seed.md  # NEW: seed command
```

## Complexity Tracking

No constitution violations. No complexity justifications
needed.
