# Implementation Plan: Sub-tool Error Reporting

**Branch**: `036-subtool-error-reporting` | **Date**: 2026-06-15 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/036-subtool-error-reporting/spec.md`

## Summary

`uf init` and `uf setup` swallow sub-tool error output,
showing only hardcoded summaries like "specify init failed"
instead of the actual error. The fix captures
`CombinedOutput()` bytes at all `ExecCmd()` call sites and
includes them in failure messages. The `subToolResult` struct
(scaffold) gains `err` and `output` fields; the `stepResult`
struct (setup) gains an `output` field. A truncation helper
limits long output to the last 10 lines. No new flags or
dependencies are required.

## Technical Context

**Language/Version**: Go 1.25+
**Primary Dependencies**: cobra, charmbracelet/log, lipgloss (no new deps)
**Storage**: N/A
**Testing**: stdlib `testing`, table-driven tests, `cmdRecorder`/`scaffoldCmdRecorder` stubs
**Target Platform**: Linux, macOS (CLI)
**Project Type**: CLI tool
**Performance Goals**: N/A (error path only)
**Constraints**: No additional output for successful steps (FR-007)
**Scale/Scope**: 2 packages (~30 ExecCmd call sites total)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after
Phase 1 design.*

### Pre-research check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Autonomous Collaboration | N/A | Internal bug fix, no inter-hero artifacts |
| II. Composability First | PASS | No new dependencies, no coupling |
| III. Observable Quality | PASS | Error messages become more diagnostic |
| IV. Testability | PASS | Coverage strategy defined in research.md |
| V. Security by Default | PASS | No new dependencies, no external input changes |

**Gate result**: PASS -- no violations.

### Post-design re-check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Autonomous Collaboration | N/A | No change from pre-research |
| II. Composability First | PASS | No new packages or dependencies introduced |
| III. Observable Quality | PASS | Error output is plain text, consistent with existing output format |
| IV. Testability | PASS | All changes testable via existing cmdRecorder stubs; coverage strategy covers 5 test categories |
| V. Security by Default | PASS | Error output shown to the same user who invoked the command; no secrets in sub-tool stderr; no new attack surface |

**Gate result**: PASS -- no violations.

## Project Structure

### Documentation (this feature)

```text
specs/036-subtool-error-reporting/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0: research decisions
├── data-model.md        # Phase 1: entity changes
├── quickstart.md        # Phase 1: before/after examples
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── scaffold/
│   ├── scaffold.go          # subToolResult struct, initDewey,
│   │                        # initSimpleTool, printSummary
│   └── scaffold_test.go     # Failure scenario tests
└── setup/
    ├── setup.go             # stepResult struct, ~25 ExecCmd
    │                        # call sites, printStepResult
    └── setup_test.go        # Failure scenario tests
```

**Structure Decision**: No new files or directories. All
changes modify existing files in existing packages. The
truncation helper is defined locally in each package (small
function, not worth a shared package).

## Complexity Tracking

No constitution violations to justify. No new projects,
patterns, or dependencies.
