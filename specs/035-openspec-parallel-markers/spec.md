---
title: OpenSpec Parallel Markers
status: draft
created: 2026-05-03
spec: "035"
---

# Feature Specification: OpenSpec Parallel Markers

**Feature Branch**: `035-openspec-parallel-markers`
**Created**: 2026-05-03
**Status**: Draft
**Input**: GitHub issue #153 — Add [P] parallel marker
support to OpenSpec task template

## User Scenarios & Testing

### User Story 1 — Parallel Task Execution for OpenSpec Changes (Priority: P1)

As a developer using `/unleash` on an OpenSpec branch,
I want independent tasks within a task group to be
marked with `[P]` so that Replicator can execute them
in parallel via git worktrees, reducing implementation
time for multi-file changes.

**Why this priority**: This is the core value
proposition. Without `[P]` markers, every OpenSpec
task runs sequentially even when tasks touch different
files and have no dependencies. For changes with 5+
independent tasks, this can double or triple execution
time.

**Independent Test**: Run `/unleash` on an `opsx/*`
branch whose `tasks.md` contains `[P]`-marked tasks.
Verify that Replicator spawns parallel workers for
`[P]` tasks and sequential workers for non-`[P]` tasks.

**Acceptance Scenarios**:

1. **Given** an OpenSpec `tasks.md` with `[P]`-marked
   tasks, **When** `/unleash` runs the implementation
   step, **Then** `[P]` tasks within the same group
   are executed in parallel (up to 4 concurrent workers)
2. **Given** an OpenSpec `tasks.md` with a mix of `[P]`
   and non-`[P]` tasks, **When** `/unleash` runs,
   **Then** non-`[P]` tasks execute first sequentially,
   followed by `[P]` tasks in parallel
3. **Given** an OpenSpec `tasks.md` with no `[P]`
   markers, **When** `/unleash` runs, **Then** all
   tasks execute sequentially (backward compatible)

---

### User Story 2 — LLM-Generated Parallel Markers (Priority: P1)

As a developer using `/opsx:propose`, I want the LLM
to automatically add `[P]` markers to tasks that touch
different files and have no inter-task dependencies, so
that I don't have to manually mark tasks for parallel
execution.

**Why this priority**: Manual `[P]` marking defeats the
purpose of automation. The schema instructions must
guide the LLM to identify independent tasks and mark
them correctly.

**Independent Test**: Run `/opsx:propose` for a change
that produces multiple independent tasks. Verify that
the generated `tasks.md` contains `[P]` markers on
tasks that touch different files.

**Acceptance Scenarios**:

1. **Given** a proposal with multiple independent file
   changes, **When** `/opsx:propose` generates tasks,
   **Then** tasks touching different files with no
   shared dependencies are marked `[P]`
2. **Given** a proposal with sequential dependencies,
   **When** `/opsx:propose` generates tasks, **Then**
   dependent tasks are NOT marked `[P]`
3. **Given** a single-file change, **When**
   `/opsx:propose` generates tasks, **Then** no tasks
   are marked `[P]` (parallel execution of tasks
   touching the same file risks merge conflicts)

---

### Edge Cases

- What happens when two `[P]` tasks modify the same
  file? Replicator's worktree merge will detect
  conflict markers and exit with instructions. The
  schema instructions should prevent this by only
  marking tasks as `[P]` when they touch different
  files.
- What happens when a `[P]` task depends on a non-`[P]`
  task in the same group? The non-`[P]` task runs first
  (sequential), then `[P]` tasks run in parallel. This
  is the existing `/unleash` behavior for Speckit tasks.
- What about existing OpenSpec `tasks.md` files without
  `[P]` markers? They continue to work — all tasks run
  sequentially. This is fully backward compatible.

## Requirements

### Functional Requirements

- **FR-001**: The OpenSpec task template MUST demonstrate
  the `[P]` marker format:
  `- [ ] N.M [P] Task description`
- **FR-002**: The OpenSpec schema instructions MUST
  instruct the LLM to add `[P]` markers on tasks that
  meet ALL of: (a) touch different files from other
  `[P]` tasks in the same group, (b) have no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints
- **FR-003**: The OpenSpec schema instructions MUST
  instruct the LLM to NOT mark tasks as `[P]` when
  they modify the same file as another task in the group
- **FR-004**: The template MUST include an example
  showing both `[P]` (parallel) and non-`[P]`
  (sequential) tasks in the same group to demonstrate
  the mixed pattern
- **FR-005**: The change MUST be backward compatible —
  existing `tasks.md` files without `[P]` markers MUST
  continue to work unchanged

### Key Entities

- **Task template**: The Markdown template at
  `openspec/schemas/unbound-force/templates/tasks.md`
  that scaffolds the structure for generated tasks
- **Schema instructions**: The `instruction` field in
  `openspec/schemas/unbound-force/schema.yaml` that
  guides the LLM during task generation

## Success Criteria

### Measurable Outcomes

- **SC-001**: `/opsx:propose` generates `tasks.md` files
  with `[P]` markers on independent tasks when the
  change involves multiple independent file modifications
- **SC-002**: `/unleash` on an `opsx/*` branch correctly
  splits `[P]` and non-`[P]` tasks into parallel and
  sequential groups (this already works — no Replicator
  changes needed)
- **SC-003**: Existing OpenSpec `tasks.md` files without
  `[P]` markers continue to execute successfully in
  `/unleash` (backward compatibility)
- **SC-004**: The task template clearly demonstrates
  the `[P]` marker format and usage rules

## Dependencies

- **Spec 018 — Unleash Command**: Defines the `[P]`
  marker detection and parallel execution logic
  (already implemented)
- **Spec 008 — Swarm Orchestration**: Defines the
  worktree-based parallel execution infrastructure
  (already implemented)
- **Replicator**: Already detects `[P]` markers
  regardless of task ID format — no Replicator changes
  needed

## Assumptions

- The `[P]` marker format (`- [ ] N.M [P] description`)
  is already parsed by `/unleash` and Replicator. No
  parser changes are needed — the marker is detected via
  string matching on `[P]`.
- The OpenSpec numbered task format (`1.1`, `1.2`) is
  compatible with the `[P]` marker — the marker appears
  after the task number, same as in Speckit format.
- LLM-generated task files will follow the template
  and schema instructions for `[P]` placement.
<!-- scaffolded by uf vdev -->
