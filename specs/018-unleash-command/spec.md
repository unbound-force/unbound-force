---
spec_id: "018"
title: "Unleash Command"
status: draft
created: 2026-03-29
branch: 018-unleash-command
phase: 2
depends_on:
  - "[[specs/003-specification-framework/spec]]"
  - "[[specs/008-swarm-orchestration/spec]]"
  - "[[specs/012-swarm-delegation/spec]]"
  - "[[specs/014-dewey-architecture/spec]]"
  - "[[specs/016-autonomous-define/spec]]"
---

# Feature Specification: Unleash Command

**Feature Branch**: `018-unleash-command`
**Created**: 2026-03-29
**Status**: Draft
**Input**: Autonomous swarm execution command that runs
the full Speckit pipeline with Dewey-powered
clarification, parallel implementation, code review,
retrospective, and demo instructions.

## User Scenarios & Testing *(mandatory)*

### User Story 1 -- Happy Path (Priority: P1)

A developer has written a specification using
`/speckit.specify` and now wants the swarm to take over.
They type `/unleash` and the command autonomously
clarifies ambiguities using Dewey, generates the
implementation plan and tasks, reviews the specs,
implements the code with parallel workers for
independent tasks, reviews the implementation, stores
learnings, and returns with demo instructions. The
developer never has to intervene.

**Why this priority**: This is the core value
proposition -- a single command that takes a spec from
draft to demo-ready code. Without this, the developer
must manually invoke 8+ commands in sequence.

**Independent Test**: Create a small spec with no
ambiguities (all questions answerable by Dewey), run
`/unleash`, verify it produces plan.md, tasks.md,
implementation code, passes code review, stores at
least one learning, and presents demo instructions.

**Acceptance Scenarios**:

1. **Given** a Speckit feature branch with a spec.md
   that has no unresolvable ambiguities,
   **When** the developer runs `/unleash`,
   **Then** the command completes autonomously through
   all 8 steps (clarify, plan, tasks, spec review,
   implement, code review, retrospective, demo) and
   presents demo instructions with verification steps.

2. **Given** a spec.md with ambiguities that Dewey can
   answer from the existing knowledge base,
   **When** `/unleash` runs the clarify step,
   **Then** Dewey's answers are auto-accepted silently,
   the spec is updated, and the pipeline continues
   without returning to the human.

3. **Given** implementation completes and code review
   passes,
   **When** the retrospective step runs,
   **Then** at least one learning is stored in
   semantic memory describing patterns, gotchas, or
   decisions from the session.

4. **Given** all steps complete successfully,
   **When** demo instructions are presented,
   **Then** the instructions include: what was built,
   how to verify, key files changed, test results,
   and the options `/finale` or `/speckit.clarify`.

---

### User Story 2 -- Exit on Unanswerable Question (Priority: P1)

During the clarify step, Dewey cannot resolve one or
more ambiguities in the spec. The command exits
gracefully, presents all unanswerable questions to the
human at once, and tells them to answer the questions
and re-run `/unleash`. When re-run, the command detects
that clarification is complete and resumes from the
plan step.

**Why this priority**: This is the primary exit point
and the most common human interaction. If the command
can't handle this gracefully, the entire workflow
breaks.

**Independent Test**: Create a spec with a
domain-specific ambiguity that Dewey cannot answer
(e.g., a business policy question). Run `/unleash`,
verify it exits with the question(s). Answer the
question manually in the spec. Re-run `/unleash`,
verify it resumes at the plan step.

**Acceptance Scenarios**:

1. **Given** a spec with an ambiguity that Dewey
   cannot resolve,
   **When** `/unleash` runs the clarify step,
   **Then** it exits with all unanswerable questions
   presented to the human, and suggests "Answer these
   questions in the spec, then re-run `/unleash`."

2. **Given** the human has resolved all clarification
   questions in the spec,
   **When** `/unleash` is re-run,
   **Then** it detects that clarification is complete
   (no `[NEEDS CLARIFICATION]` markers, Clarifications
   section exists) and skips to the plan step.

3. **Given** Dewey can answer 2 of 3 ambiguities,
   **When** `/unleash` runs the clarify step,
   **Then** the 2 Dewey-answered questions are
   auto-resolved in the spec and only the 1
   unanswerable question is presented to the human.

---

### User Story 3 -- Exit on Spec Review Failure (Priority: P2)

The spec review council finds HIGH or CRITICAL issues
in the spec artifacts. The command exits with the
findings and tells the human to fix the issues and
re-run `/unleash`. When re-run, it detects that spec
review needs to re-run (tasks exist but implementation
has not started).

**Why this priority**: Spec quality gates prevent wasted
implementation effort. This exit point ensures the swarm
doesn't implement a flawed spec.

**Independent Test**: Create a spec with a deliberate
HIGH-severity issue (e.g., missing acceptance criteria
for a user story). Run `/unleash`, verify it exits
after spec review with the findings.

**Acceptance Scenarios**:

1. **Given** the spec review council finds HIGH or
   CRITICAL issues after plan and tasks generation,
   **When** `/unleash` reports the findings,
   **Then** it presents the issues with context and
   suggests "/speckit.clarify to address the findings,
   then re-run /unleash."

2. **Given** the spec review council finds only LOW
   and MEDIUM issues,
   **When** `/unleash` processes the review,
   **Then** LOW/MEDIUM issues are auto-fixed and the
   pipeline continues to implementation.

---

### User Story 4 -- Parallel Implementation (Priority: P2)

During implementation, tasks marked with `[P]` in
tasks.md are executed in parallel using Swarm workers
with dedicated worktrees. Sequential tasks run one at
a time. Each phase has a build+test checkpoint before
proceeding.

**Why this priority**: Parallel execution is the key
differentiator from the existing sequential
`/cobalt-crush` workflow. It reduces implementation
time for specs with many independent tasks.

**Independent Test**: Create a tasks.md with a phase
containing 4 `[P]`-marked test tasks (different files).
Run `/unleash`, verify all 4 tasks execute in parallel
via separate Swarm workers, and the phase checkpoint
passes.

**Acceptance Scenarios**:

1. **Given** a tasks.md phase with 4 `[P]`-marked
   tasks touching different files,
   **When** `/unleash` reaches that phase,
   **Then** it spawns 4 parallel Swarm workers, each
   in a dedicated worktree, and merges results after
   all complete.

2. **Given** a tasks.md phase with 2 sequential tasks
   (no `[P]` marker),
   **When** `/unleash` reaches that phase,
   **Then** it executes tasks one at a time in order.

3. **Given** a parallel worker fails during execution,
   **When** the failure is detected,
   **Then** no additional workers are spawned,
   already-running workers complete or fail, all
   worktrees are cleaned up, and the command exits to
   the human with error context.

4. **Given** parallel workers complete and worktree
   merge produces a conflict,
   **When** the conflict is detected,
   **Then** the command attempts auto-resolution
   (accept both changes). If auto-resolution succeeds,
   the pipeline continues. If auto-resolution fails
   (semantic conflict), the command exits to the human
   with conflict details.

---

### User Story 5 -- Exit on Code Review Failure (Priority: P2)

After implementation, the code review council finds
issues that cannot be auto-fixed within 3 iterations.
The command exits with the persistent findings and
tells the human how to address them.

**Why this priority**: Code review is the quality gate
before presenting demo instructions. Persistent
failures need human judgment.

**Independent Test**: Create a scenario where a review
finding causes a fix that breaks another reviewer's
check (circular dependency). Verify `/unleash` stops
after 3 iterations and reports the circular findings.

**Acceptance Scenarios**:

1. **Given** the code review council finds issues,
   **When** `/unleash` attempts fixes,
   **Then** it retries up to 3 iterations, fixing
   findings and re-running the review each time.

2. **Given** 3 fix iterations are exhausted with
   remaining issues,
   **When** `/unleash` cannot resolve the findings,
   **Then** it exits with the outstanding findings
   and suggests the human fix them manually before
   re-running `/unleash`.

---

### User Story 6 -- Resumability (Priority: P3)

The command can be re-run at any point after an exit
and will resume from the correct step. State is
detected from filesystem artifacts, not from a
separate state file.

**Why this priority**: Resumability enables the
iterate-and-resume workflow. Without it, every exit
would require re-running all steps from scratch.

**Independent Test**: Run `/unleash` through plan
step, interrupt it. Re-run, verify it skips clarify
and plan (both exist) and resumes at tasks.

**Acceptance Scenarios**:

1. **Given** plan.md and tasks.md already exist in the
   feature directory,
   **When** `/unleash` is re-run,
   **Then** it skips clarify, plan, and tasks steps
   and resumes at spec review.

2. **Given** all tasks in tasks.md are marked `[x]`
   and tests pass,
   **When** `/unleash` is re-run,
   **Then** it skips implementation and resumes at
   code review.

3. **Given** no spec artifacts exist beyond spec.md,
   **When** `/unleash` is run for the first time,
   **Then** it starts from the clarify step.

---

### Edge Cases

- What happens when `/unleash` is run on `main`? The
  command refuses with an error: "Must be on a Speckit
  feature branch (`NNN-*`). Run `/speckit.specify`
  first."
- What happens when `/unleash` is run on an `opsx/*`
  branch? The command refuses with an error: "/unleash
  is for Speckit strategic specs only. Use `/opsx:apply`
  for OpenSpec tactical changes."
- What happens when spec.md does not exist? The command
  refuses with an error: "No spec.md found. Run
  `/speckit.specify` first."
- What happens when Dewey is not available during
  clarify? The command falls back to presenting all
  questions to the human (same behavior as if Dewey
  couldn't answer any of them).
- What happens when Gaze is not installed during code
  review? The code review proceeds without Gaze quality
  data (informational note, not blocking).
- What happens when Swarm worktree creation fails? The
  command falls back to sequential execution for that
  phase and reports the worktree failure as a warning.
- What happens when the build fails at a phase
  checkpoint? The command stops implementation, reports
  the build failure, and exits to the human.
- What happens when all tasks are already complete but
  code review hasn't run? The command skips to code
  review.
- What happens when hivemind is not available for the
  retrospective? The retrospective step is skipped with
  an informational note. This is non-blocking.
- What happens when the session is interrupted during
  parallel worktree execution? On every `/unleash`
  invocation, a startup cleanup step runs
  `swarm_worktree_list` and `swarm_worktree_cleanup`
  for any stale worktrees from previous runs before
  proceeding. The cleanup only removes orphaned
  worktrees -- it does NOT touch spec.md, plan.md,
  tasks.md, or task checkboxes. Resumability is
  preserved because those artifacts live in the main
  working directory, not in worktrees.
- What happens when worktree merges produce conflicts?
  `swarm_worktree_merge` uses cherry-pick. If no
  conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`)
  remain after the cherry-pick, auto-resolution
  succeeded and the pipeline continues (after
  build+test checkpoint). If conflict markers remain,
  the command exits to the human with conflict details
  and the affected files.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `/unleash` MUST execute the full Speckit
  pipeline autonomously: clarify, plan, tasks, spec
  review, implement, code review, retrospective, demo.
- **FR-002**: `/unleash` MUST only run on Speckit
  feature branches (`NNN-*` pattern). It MUST refuse
  to run on `main`, `opsx/*`, or unrecognized branches.
- **FR-003**: `/unleash` MUST require spec.md to exist
  in the feature directory before proceeding.
- **FR-004**: During the clarify step, `/unleash` MUST
  use Dewey semantic search to attempt to answer
  ambiguities in the spec before presenting them to the
  human. The orchestrating agent formulates targeted
  search queries from the question text and surrounding
  spec context (not raw `[NEEDS CLARIFICATION]` text
  verbatim).
- **FR-005**: Dewey-resolved answers MUST be
  auto-accepted silently and written to the spec
  without human confirmation. The orchestrating agent
  uses its judgment to determine whether Dewey's
  search results sufficiently answer the question
  (no numeric similarity threshold -- agent reads
  the results and decides). Each auto-accepted answer
  MUST include a provenance annotation in the
  Clarifications section: `(Dewey-resolved from
  [page/block reference])` so humans can audit which
  answers came from Dewey vs. human input.
- **FR-006**: If any ambiguities remain after Dewey
  resolution, `/unleash` MUST exit and present all
  remaining questions to the human at once.
- **FR-007**: `/unleash` MUST be resumable -- re-running
  after an exit MUST detect completed steps from
  filesystem artifacts and skip forward to the next
  incomplete step.
- **FR-008**: Step completion MUST be detected from
  filesystem state: plan.md exists (plan done),
  tasks.md exists (tasks done), spec review passed
  (a `<!-- spec-review: passed -->` marker exists in
  tasks.md, written by the spec review step), all
  tasks `[x]` (implementation done), build+test pass
  (code review candidate).
- **FR-009**: During implementation, `/unleash` MUST
  execute `[P]`-marked tasks in parallel using Swarm
  workers with dedicated worktrees. The maximum number
  of concurrent workers MUST be limited to 4 (default).
  If a phase has more than 4 `[P]` tasks, the remaining
  tasks are batched and executed after the first batch
  completes.
- **FR-010**: During implementation, non-`[P]` tasks
  MUST execute sequentially via a single agent.
- **FR-011**: If any parallel worker fails, `/unleash`
  MUST stop spawning new workers, wait for
  already-running workers to complete or fail, clean
  up all worktrees via `swarm_worktree_cleanup`, and
  exit to the human with error context. The command
  cannot forcibly terminate running agents -- "stop"
  means best-effort (no new spawns + cleanup).
- **FR-011a**: Worktree merges use
  `swarm_worktree_merge` (cherry-pick). Auto-resolution
  succeeds when no conflict markers (`<<<<<<<`,
  `=======`, `>>>>>>>`) remain in any file after the
  cherry-pick. If conflict markers remain, `/unleash`
  MUST exit to the human with conflict details. After
  successful auto-resolution, the phase checkpoint
  (build+test) MUST run to verify the merged result.
- **FR-012**: Each implementation phase MUST have a
  build+test checkpoint before proceeding to the next
  phase. The build and test commands MUST be derived
  from `.github/workflows/` (the CI workflow files are
  the source of truth), not hardcoded to a specific
  language. This is the same pattern used by
  `/review-council` Phase 1a.
- **FR-013**: The spec review step MUST use the
  `/review-council` in spec review mode, auto-fixing
  LOW/MEDIUM findings and exiting on HIGH/CRITICAL.
  This step subsumes the `/speckit.analyze` and
  `/speckit.checklist` phases from the canonical
  Speckit pipeline -- the review council's 5 Divisor
  agents provide equivalent coverage (consistency
  analysis + quality validation) in a single pass.
- **FR-014**: The code review step MUST use the
  `/review-council` in code review mode (CI hard gate
  + Gaze + Divisor agents), retrying up to 3
  iterations.
- **FR-015**: The retrospective step MUST store at
  least one learning in semantic memory via
  `hivemind_store` describing patterns, gotchas, or
  decisions from the session.
- **FR-016**: The demo step MUST present: what was
  built, how to verify, key files changed, test
  results summary, and next-step options (`/finale`
  or `/speckit.clarify`).
- **FR-017**: `/unleash` MUST gracefully degrade when
  optional tools are unavailable: Dewey (fall back to
  human questions), Gaze (skip quality analysis),
  Swarm worktrees (fall back to sequential), Hivemind
  (skip retrospective).
- **FR-018**: The command MUST be deployed as a
  tool-owned scaffold asset at
  `.opencode/command/unleash.md`, deployable via
  `uf init`.

### Key Entities

- **Pipeline Step**: One of the 8 steps in the
  `/unleash` pipeline (clarify, plan, tasks,
  spec-review, implement, code-review, retrospective,
  demo). Each step has a detection condition for
  resumability.
- **Exit Point**: A condition that causes `/unleash`
  to return control to the human with accumulated
  context and next-step instructions.
- **Swarm Worker**: A parallel execution unit spawned
  via `swarm_spawn_subtask` with a dedicated git
  worktree for file isolation.
- **Learning**: A semantic memory entry stored via
  `hivemind_store` with tags linking it to the feature
  branch and session.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can go from spec.md to
  demo-ready code with a single `/unleash` invocation
  when no human-required clarifications exist.
- **SC-002**: Re-running `/unleash` after answering
  clarification questions resumes from the correct
  step without re-executing completed steps.
- **SC-003**: Parallel `[P]` tasks in a phase complete
  faster than sequential execution by utilizing
  multiple Swarm workers simultaneously.
- **SC-004**: The retrospective stores at least one
  learning per `/unleash` session that can be retrieved
  by future sessions via `hivemind_find`.
- **SC-005**: The demo instructions are sufficient for
  a developer to verify the implementation without
  reading the full spec or plan.
- **SC-006**: All exit points present actionable
  next-step instructions that a developer can follow
  without additional context.
- **SC-007**: `/unleash` degrades gracefully when
  optional tools (Dewey, Gaze, Swarm worktrees,
  Hivemind) are unavailable -- the pipeline completes
  with reduced functionality rather than failing.

## Clarifications

### Session 2026-03-29

- Q: How does `/unleash` decide whether Dewey's search
  results sufficiently answer a clarification question?
  → A: Agent judgment -- the orchestrating agent reads
  Dewey results and decides (no numeric similarity
  threshold).
- Q: How should `/unleash` query Dewey for clarification
  answers? → A: Agent formulates a semantic search query
  from the question text + surrounding spec context (not
  raw NEEDS CLARIFICATION text verbatim).
- Q: How should worktree merge conflicts be handled
  after parallel workers complete? → A: Attempt
  auto-resolution (accept both changes), continue if
  no semantic conflicts. Exit to human only if
  auto-resolution fails.

### Session 2026-03-29 (Review Council)

- Q: How should orphaned worktrees from interrupted
  sessions be handled? → A: Startup cleanup step runs
  `swarm_worktree_list` + `swarm_worktree_cleanup`
  on every invocation. Does not affect resumability
  (artifacts are in main working directory).
- Q: What is the relationship between `/unleash` and
  the `/workflow` system (Specs 008/012/016)? → A:
  `/unleash` operates at the Speckit pipeline level,
  not the hero lifecycle workflow level. No
  WorkflowInstance objects created.
- Q: Should Dewey auto-resolved answers include
  provenance? → A: Yes, add `(Dewey-resolved from
  [page/block])` annotation for auditability.
- Q: What merge strategy for worktree conflicts? → A:
  Cherry-pick (what `swarm_worktree_merge` uses).
  Success = no conflict markers remain. Failure =
  conflict markers present → exit to human.
- Q: How should spec review completion be detected for
  resumability? → A: Explicit `<!-- spec-review:
  passed -->` marker in tasks.md, not inferred from
  task checkbox state.
- Q: Can `/unleash` actually stop running parallel
  workers? → A: No. Revise to "stop spawning new
  workers, wait for running ones, cleanup worktrees."
- Q: Should parallel workers have a concurrency limit?
  → A: Yes, max 4 concurrent workers (default). Batch
  remaining `[P]` tasks.
- Q: Should build/test commands be Go-specific? → A:
  No, derive from `.github/workflows/` (CI files are
  source of truth), same pattern as `/review-council`.
- Q: Are `/speckit.analyze` and `/speckit.checklist`
  missing from the pipeline? → A: Subsumed by the
  spec review step (review council provides equivalent
  coverage).

## Assumptions

- The developer has already run `/speckit.specify` to
  create the spec.md and feature branch before running
  `/unleash`.
- Dewey MCP tools are the primary mechanism for
  answering clarification questions. Dewey's semantic
  search across the knowledge base (specs, code,
  documentation) provides sufficient context to answer
  most domain questions autonomously.
- The Swarm plugin is installed and `swarm_spawn_subtask`
  and `swarm_worktree_create` tools are available for
  parallel execution. If not, sequential fallback is
  used.
- The `/review-council` command is available and its
  Code Review Mode (Phase 1a CI + Phase 1b Gaze +
  Divisor agents) is the quality gate for implementation.
- Hivemind (`hivemind_store`/`hivemind_find`) is
  available for the retrospective step. If not, the
  step is skipped.
- The `/unleash` command is a Markdown instruction file
  (`.opencode/command/unleash.md`), not Go code. It
  orchestrates existing slash commands and Swarm tools.
- `/unleash` operates at the Speckit pipeline level,
  NOT the hero lifecycle workflow level. It does not
  create or advance `WorkflowInstance` objects from
  Spec 008. The 8 `/unleash` steps map to the Speckit
  pipeline (clarify → plan → tasks → review →
  implement) with added code review and retrospective
  steps. The `/workflow` system (Specs 008/012/016) is
  a separate orchestration layer for multi-hero
  artifact handoff.
- The command file is tool-owned and auto-updated by
  `uf init` (same as `/finale`, `/cobalt-crush`, etc.).
- Parallel execution uses Swarm's `swarm_worktree_create`
  for git isolation and `swarmmail_reserve` for file
  locking. Worktrees are merged back via
  `swarm_worktree_merge` after all parallel workers in
  a phase complete.
