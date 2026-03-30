# Implementation Plan: Unleash Command

**Branch**: `018-unleash-command` | **Date**: 2026-03-29 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/018-unleash-command/spec.md`

## Summary

Create the `/unleash` slash command -- a single
autonomous pipeline that takes a Speckit spec from
draft to demo-ready code. The command orchestrates
8 steps (clarify, plan, tasks, spec review, implement,
code review, retrospective, demo) with Dewey-powered
clarification, parallel Swarm workers for `[P]` tasks,
and filesystem-based resumability. Exits to the human
only when it genuinely needs human judgment.

The deliverable is a Markdown command file
(`.opencode/command/unleash.md`) plus its scaffold
asset copy. No new Go production code -- test assertion
updates only (file count and asset list).

## Technical Context

**Language/Version**: Markdown (OpenCode command file)
**Primary Dependencies**: Existing slash commands
(`/speckit.plan`, `/speckit.tasks`, `/review-council`,
`/cobalt-crush`), Swarm MCP tools
(`swarm_spawn_subtask`, `swarm_worktree_create`,
`swarm_worktree_merge`, `swarm_worktree_cleanup`),
Dewey MCP tools (`dewey_semantic_search`,
`dewey_search`), Hivemind tools (`hivemind_store`)
**Storage**: N/A (orchestrates existing tools)
**Testing**: Scaffold drift detection test (file count
assertion 51 → 52), manual integration testing
**Target Platform**: OpenCode agent runtime
**Project Type**: CLI (scaffold asset)
**Performance Goals**: N/A (agent instruction file)
**Constraints**: Must fit in a single Markdown file.
Must orchestrate existing commands without modifying
them. Must be resumable from filesystem state alone.
**Scale/Scope**: Single command file (~300-400 lines)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check
after Phase 1 design.*

### I. Autonomous Collaboration -- PASS

`/unleash` orchestrates existing hero agents through
their slash commands. Each hero produces artifacts that
the next step consumes (plan.md → tasks.md →
implementation → review findings). No runtime coupling
-- each step reads filesystem artifacts.

### II. Composability First -- PASS

The command gracefully degrades when tools are missing:
Dewey unavailable → fall back to human questions.
Gaze unavailable → skip quality analysis. Swarm
worktrees unavailable → fall back to sequential.
Hivemind unavailable → skip retrospective. The command
works with a minimal toolset (just OpenCode + git).

### III. Observable Quality -- PASS

Every step produces observable artifacts: plan.md,
tasks.md, review findings, test results, learnings.
The demo step summarizes all results. Exit points
present full context with actionable next steps.

### IV. Testability -- PASS

The command file is tested indirectly: scaffold drift
detection verifies it's properly embedded. The file
count test asserts the correct number of assets (52).
The command orchestrates tested sub-commands rather
than implementing untested logic itself. Each
sub-command has its own test coverage.

## Project Structure

### Documentation (this feature)

```text
specs/018-unleash-command/
├── spec.md
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
.opencode/command/
└── unleash.md                    # The command file (NEW)

internal/scaffold/assets/opencode/command/
└── unleash.md                    # Scaffold asset copy (NEW)

internal/scaffold/
└── scaffold_test.go              # File count update (51 → 52)

cmd/unbound-force/
└── main_test.go                  # File count update (51 → 52)
```

**Structure Decision**: This is a Markdown command file
deployed via the existing scaffold system. No new Go
packages, no new directories. Two new files (command +
scaffold asset) and two test assertion updates.

## Coverage Strategy

### Test Coverage

This feature produces a Markdown instruction file, not
Go code. Test coverage comes from:

1. **Scaffold drift detection**: Verifies the live
   `.opencode/command/unleash.md` matches the embedded
   `internal/scaffold/assets/opencode/command/unleash.md`
2. **File count assertion**: `cmd/unbound-force/main_test.go`
   asserts 52 files processed (up from 51)
3. **Expected asset list**: `internal/scaffold/scaffold_test.go`
   includes `opencode/command/unleash.md` in the
   `expectedAssetPaths` slice
4. **Sub-command coverage**: Each orchestrated command
   (`/speckit.plan`, `/review-council`, `/cobalt-crush`,
   `/finale`) has its own test suite. `/unleash`
   composes them without adding new untested logic.

### Manual Integration Testing

The quickstart.md provides verification steps for
testing the full pipeline on a real spec.

## Complexity Tracking

No constitution violations. No complexity justifications
needed.
