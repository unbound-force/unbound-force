# Feature Specification: Unleash OpenSpec Support

**Feature Branch**: `031-unleash-openspec`  
**Created**: 2026-04-15  
**Status**: Draft  
**Input**: User description: "issue #105 — Extend /unleash to support OpenSpec (opsx/*) branches"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Single Command for Both Workflows (Priority: P1)

When an engineer runs `/unleash` on an `opsx/*` branch,
the autonomous pipeline runs the OpenSpec variant —
spec review, implement, code review, retrospective, and
demo — using artifacts from `/opsx-propose`. Today,
`/unleash` stops with an error on `opsx/*` branches,
forcing engineers through a manual multi-step workflow
(`/opsx-apply` → `/review-council` → `/finale`) where
spec review and code review are optional and frequently
skipped.

**Why this priority**: `/unleash` is the primary
autonomous pipeline. Excluding OpenSpec changes from it
means tactical changes get less quality scrutiny than
strategic specs — the opposite of what you'd expect
(smaller changes should be faster to validate, not
exempt from validation).

**Independent Test**: Create an OpenSpec change via
`/opsx-propose`. Switch to the `opsx/<name>` branch.
Run `/unleash`. Verify it detects the OpenSpec workflow,
finds the change artifacts, and runs spec review →
implement → code review → retrospective → demo.

**Acceptance Scenarios**:

1. **Given** the engineer is on branch `opsx/<name>`
   with completed artifacts from `/opsx-propose`,
   **When** they run `/unleash`, **Then** the OpenSpec
   pipeline runs: spec review → implement → code review
   → retrospective → demo.
2. **Given** the engineer is on branch `opsx/<name>`
   but no `tasks.md` exists in the change directory,
   **When** they run `/unleash`, **Then** it stops with:
   "No tasks.md found for change `<name>`. Run
   `/opsx-propose` first."
3. **Given** the engineer is on branch `NNN-*`,
   **When** they run `/unleash`, **Then** the Speckit
   pipeline runs unchanged (backward compatible).
4. **Given** the engineer is on `main`, **When** they
   run `/unleash`, **Then** it stops with the existing
   error (unchanged).

---

### User Story 2 — Resumability for OpenSpec (Priority: P1)

When `/unleash` is interrupted during an OpenSpec
pipeline and re-run, it resumes from where it left off
— the same resumability that Speckit changes enjoy.
The same `<!-- spec-review: passed -->` and
`<!-- code-review: passed -->` markers are used.

**Why this priority**: Without resumability, a network
interruption during implementation forces re-running
spec review. The markers make the pipeline restartable.

**Independent Test**: Run `/unleash` on an `opsx/*`
branch. After spec review passes, interrupt the session.
Re-run `/unleash`. Verify it skips spec review
(marker present) and resumes at implementation.

**Acceptance Scenarios**:

1. **Given** spec review has passed (marker in tasks.md),
   **When** `/unleash` re-runs, **Then** it skips to
   implementation.
2. **Given** all tasks are `[x]` but code review marker
   is absent, **When** `/unleash` re-runs, **Then** it
   resumes at code review.
3. **Given** both markers are present and all tasks
   done, **When** `/unleash` re-runs, **Then** it skips
   to retrospective (idempotent).

---

### User Story 3 — Skip Clarify/Plan/Tasks for OpenSpec (Priority: P2)

When `/unleash` runs the OpenSpec pipeline, it skips
Steps 1-3 (clarify, plan, tasks) because `/opsx-propose`
has already created all artifacts in one step. The
pipeline starts at Step 4 (spec review).

**Why this priority**: This is the key architectural
difference from Speckit. Without skipping, `/unleash`
would try to generate a plan.md and tasks.md that
already exist, potentially overwriting the artifacts
from `/opsx-propose`.

**Independent Test**: Run `/unleash` on an `opsx/*`
branch. Verify the output announces "OpenSpec mode —
artifacts from /opsx-propose, skipping clarify/plan/
tasks" and proceeds directly to spec review.

**Acceptance Scenarios**:

1. **Given** an `opsx/*` branch with proposal.md,
   design.md, specs/, and tasks.md from `/opsx-propose`,
   **When** `/unleash` runs, **Then** it announces that
   clarify/plan/tasks are skipped and starts at spec
   review.
2. **Given** an `opsx/*` branch, **When** the
   resumability detection runs, **Then** clarify, plan,
   and tasks are always reported as "done" (not checked).

---

### User Story 4 — Spec Review with OpenSpec Artifacts (Priority: P2)

When `/unleash` runs spec review for an OpenSpec change,
the review council receives the OpenSpec artifacts
(proposal.md, design.md, specs/, tasks.md) instead of
the Speckit artifacts (spec.md, plan.md, tasks.md). The
council is informed this is an OpenSpec tactical change
so it can calibrate expectations (smaller scope, fewer
user stories).

**Why this priority**: The review council must know
which artifacts to review. Sending it Speckit-style
instructions when the artifacts are OpenSpec-style
would produce confused or irrelevant findings.

**Independent Test**: Run `/unleash` on an `opsx/*`
branch. Verify the spec review step passes the correct
artifact paths and workflow tier to the review council.

**Acceptance Scenarios**:

1. **Given** an `opsx/*` branch, **When** the spec
   review step runs, **Then** the review council
   receives `openspec/changes/<name>/` as the review
   scope with OpenSpec workflow tier context.
2. **Given** a Speckit branch, **When** the spec review
   step runs, **Then** the review council receives
   `specs/NNN-*/` as the review scope (unchanged).

---

### Edge Cases

- What happens when multiple OpenSpec changes have
  tasks.md files? `/unleash` uses the branch name
  `opsx/<name>` to determine which change directory
  to use. If the branch name doesn't match any change
  directory, it stops with an error.
- What happens when an `opsx/*` branch exists but the
  change directory was archived? `/unleash` checks
  `openspec/changes/<name>/tasks.md` — if it doesn't
  exist, it stops with "Run `/opsx-propose` first."
- What happens when the OpenSpec tasks.md has different
  formatting than Speckit tasks.md? Both use the same
  checkbox format (`- [ ]` / `- [x]`) and the same
  marker comments. The task execution loop is
  format-agnostic.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `/unleash` MUST detect `opsx/*` branches
  and run the OpenSpec pipeline instead of stopping
  with an error.
- **FR-002**: `/unleash` MUST extract the change name
  from the branch name (`opsx/<name>` → `<name>`) and
  use `openspec/changes/<name>/` as the feature
  directory.
- **FR-003**: `/unleash` MUST check for
  `openspec/changes/<name>/tasks.md` as the entry gate
  for OpenSpec changes.
- **FR-004**: `/unleash` MUST skip Steps 1-3 (clarify,
  plan, tasks) for OpenSpec changes, announcing that
  artifacts are from `/opsx-propose`.
- **FR-005**: `/unleash` MUST run the review council
  in Spec Review Mode with the OpenSpec workflow tier
  and the change directory as the review scope.
- **FR-006**: `/unleash` MUST use the same
  `<!-- spec-review: passed -->` and
  `<!-- code-review: passed -->` markers for OpenSpec
  resumability.
- **FR-007**: Steps 4-8 (spec review, implement, code
  review, retrospective, demo) MUST work identically
  for both Speckit and OpenSpec modes, using the
  detected feature directory.
- **FR-008**: The Speckit pipeline MUST remain unchanged
  for `NNN-*` branches (backward compatible).
- **FR-009**: The scaffold asset copy of `unleash.md`
  MUST be synchronized after modification.
- **FR-010**: All existing tests MUST continue to pass.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running `/unleash` on an `opsx/*` branch
  with completed artifacts executes the full pipeline
  (spec review → implement → code review →
  retrospective → demo) without errors — verified by
  running on a test OpenSpec change.
- **SC-002**: Running `/unleash` on an `opsx/*` branch
  produces the same quality gates (spec review + code
  review) as a Speckit branch — verified by checking
  for both marker comments in tasks.md after completion.
- **SC-003**: Running `/unleash` on a `NNN-*` branch
  produces identical behavior to the current
  implementation — verified by running on an existing
  Speckit spec.
- **SC-004**: Resumability works for OpenSpec — verified
  by interrupting and re-running `/unleash` on an
  `opsx/*` branch, confirming it skips completed steps.
- **SC-005**: All existing tests pass — verified by
  running the full test suite.

## Dependencies & Assumptions

### Dependencies

- **`/opsx-propose`** (existing): Creates the OpenSpec
  artifacts that `/unleash` consumes.
- **Review council** (existing): Already supports
  OpenSpec spec review mode with workflow tier context.
- **Task execution loop** (existing): Format-agnostic —
  works with any tasks.md that uses checkbox format.

### Assumptions

- The branch name `opsx/<name>` always matches a change
  directory at `openspec/changes/<name>/`. If not, the
  command stops with an actionable error.
- OpenSpec `tasks.md` uses the same checkbox format
  (`- [ ]` / `- [x]`) as Speckit `tasks.md`.
- The review council can determine review scope from
  the feature directory path — `openspec/changes/` vs
  `specs/` is sufficient to identify the workflow tier.
- The demo step works identically for both workflows —
  it reads the spec/proposal for "What Was Built" and
  the acceptance scenarios for "How to Verify."
