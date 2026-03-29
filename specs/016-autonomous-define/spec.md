---
spec_id: "016"
title: "Autonomous Define with Dewey"
phase: 3
status: complete
depends_on:
  - "[[specs/012-swarm-delegation/spec]]"
  - "[[specs/014-dewey-architecture/spec]]"
  - "[[specs/015-dewey-integration/spec]]"
---

# Feature Specification: Autonomous Define with Dewey

**Feature Branch**: `016-autonomous-define`
**Created**: 2026-03-26
**Status**: Draft
**Input**: User description: "Enable the define stage
to run as swarm mode so Muti-Mind can autonomously
draft specifications using Dewey context, reducing
human checkpoints from 2 to 1"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configurable Define Stage Execution Mode (Priority: P1)

A team lead configures the hero lifecycle workflow so
that the define stage runs as `[swarm]` instead of
`[human]`. When a developer seeds a feature (provides
a brief intent description), the swarm takes over
immediately -- Muti-Mind drafts the specification
autonomously using Dewey's cross-repo context, then
the swarm continues through implementation, validation,
and review. The workflow pauses only at the accept
stage for human review.

**Why this priority**: This is the core mechanism. Without
configurable execution modes, the define stage is always
human-driven and the workflow requires two human
checkpoints. Making the mode configurable enables the
single-checkpoint workflow (seed + accept) that is the
stated goal of the Dewey initiative.

**Independent Test**: Start a workflow with the define
stage configured as `[swarm]`. Provide a one-sentence
seed. Verify the workflow advances through define
without human intervention, and pauses at accept.

**Acceptance Scenarios**:

1. **Given** a workflow configuration with define mode
   set to `swarm`, **When** a workflow is started with
   a backlog item, **Then** the define stage executes
   autonomously and the workflow advances to implement
   without requiring `/workflow advance`.
2. **Given** a workflow configuration with define mode
   set to `human` (the default), **When** a workflow is
   started, **Then** the define stage behaves exactly as
   before Spec 012 -- the human must advance manually.
3. **Given** a configurable execution mode map,
   **When** the operator changes the define stage from
   `human` to `swarm`, **Then** only the define stage
   behavior changes -- all other stages retain their
   configured modes.
4. **Given** a workflow with define in `swarm` mode,
   **When** all swarm stages complete (define through
   review), **Then** the workflow pauses at
   `awaiting_human` before the accept stage, exactly
   as in Spec 012's checkpoint behavior.

---

### User Story 2 - Muti-Mind Autonomous Specification (Priority: P1)

The Product Owner agent (Muti-Mind) autonomously drafts
a feature specification when the define stage runs in
swarm mode. It uses Dewey's semantic search to retrieve
cross-repo context (related issues, past specs, toolstack
docs), generates acceptance criteria informed by
historical patterns, and self-clarifies ambiguities by
querying Dewey instead of asking the human.

**Why this priority**: Without autonomous specification
drafting, configurable execution modes (US1) are
meaningless -- the swarm can't execute the define stage
without Muti-Mind knowing how to draft a spec on its
own. This is tied with US1 because both are needed for
the feature to work.

**Independent Test**: Trigger Muti-Mind in autonomous
mode with a seed description like "add CSV export to
the dashboard." Verify the produced specification
includes acceptance criteria, functional requirements,
and that the criteria reference real cross-repo context
(not hallucinated).

**Acceptance Scenarios**:

1. **Given** a seed description ("add CSV export to the
   dashboard") and Dewey is running with indexed
   content, **When** Muti-Mind executes the define stage
   autonomously, **Then** it produces a specification
   with acceptance criteria that reference real
   patterns found in the Dewey index.
2. **Given** an ambiguous seed description, **When**
   Muti-Mind encounters a clarification point, **Then**
   it queries Dewey for context (e.g., "what
   authentication method does this project use?") and
   resolves the ambiguity without asking the human.
3. **Given** at least 3 completed workflow records with
   learning feedback, **When** Muti-Mind drafts a new
   spec, **Then** it references relevant lessons learned
   (e.g., "past features with vague performance criteria
   were rejected 60% of the time") to produce more
   precise criteria.
4. **Given** Dewey is unavailable (Tier 1 degradation),
   **When** Muti-Mind attempts autonomous specification,
   **Then** it falls back to using local backlog items
   and convention packs, producing a less contextual
   but still valid specification.

---

### User Story 3 - Optional Spec Review Checkpoint (Priority: P2)

For high-stakes features, the operator can configure
an optional specification review checkpoint between the
autonomous define stage and the implement stage. When
enabled, the workflow pauses after Muti-Mind produces
the spec, allowing the human to scan the acceptance
criteria for intent alignment before the swarm invests
effort in implementation.

**Why this priority**: This is a risk mitigation
mechanism. The primary workflow (US1+US2) works without
it, but high-stakes features benefit from a lightweight
human review of the spec before implementation begins.
This is P2 because the MVP can ship without it.

**Independent Test**: Configure a workflow with the spec
review checkpoint enabled. Seed a feature. Verify the
workflow pauses after define (before implement) with
`awaiting_human` status and a message indicating spec
review is available.

**Acceptance Scenarios**:

1. **Given** a workflow with the spec review checkpoint
   enabled, **When** Muti-Mind completes the autonomous
   define stage, **Then** the workflow pauses with
   `awaiting_human` status and a message indicating the
   spec is ready for review.
2. **Given** a paused workflow at the spec review
   checkpoint, **When** the human runs
   `/workflow advance`, **Then** the workflow resumes
   and the implement stage begins.
3. **Given** a workflow WITHOUT the spec review
   checkpoint (default), **When** Muti-Mind completes
   the define stage, **Then** the workflow advances
   directly to implement without pausing.

---

### User Story 4 - Seed Workflow Command (Priority: P2)

The operator can start a workflow with a one-sentence
seed description using a single command, rather than
the current multi-step process of creating a backlog
item, running `/speckit.specify`, and then advancing.
The seed command combines intent expression with
workflow initiation.

**Why this priority**: This reduces the friction of
starting a new feature from "create a backlog item,
specify, clarify, advance" to a single command. It's
P2 because the autonomous define (US1+US2) works with
the existing `/workflow start` command -- the seed
command is a UX improvement.

**Independent Test**: Run the seed command with a
feature description. Verify a workflow starts with the
define stage in swarm mode and Muti-Mind begins
autonomous specification.

**Acceptance Scenarios**:

1. **Given** the operator types a seed command with a
   description, **When** the command executes, **Then**
   a new workflow starts with the define stage in swarm
   mode and the seed description is passed to Muti-Mind
   as the feature intent.
2. **Given** the operator provides an empty seed
   description, **When** the command executes, **Then**
   it prompts the operator for a description before
   proceeding.

---

### User Story 5 - Updated Documentation and SKILL.md (Priority: P3)

The SKILL.md, workflow command docs, AGENTS.md, and
website are updated to document the seed-to-accept
workflow. Operators reading the documentation understand
that the define stage can be configured as swarm-driven,
that Muti-Mind uses Dewey for context, and that a spec
review checkpoint is available for high-stakes features.

**Why this priority**: Documentation is a polish step
that makes the feature discoverable. The feature works
without updated docs, but adoption depends on it.

**Independent Test**: Read the updated SKILL.md and
Common Workflows page. Verify they describe the
seed-to-accept workflow with execution mode
configuration.

**Acceptance Scenarios**:

1. **Given** the SKILL.md for the hero lifecycle,
   **When** an operator reads it, **Then** it describes
   the configurable define stage and the seed-to-accept
   workflow.
2. **Given** the Common Workflows page on the website,
   **When** a visitor reads it, **Then** it describes
   both the 2-checkpoint workflow (default) and the
   1-checkpoint workflow (autonomous define).

---

### Edge Cases

- What happens when Dewey is unavailable during
  autonomous define? Muti-Mind falls back to Tier 1
  (local file reads + convention packs). The spec is
  less contextual but still valid. The workflow does
  not block waiting for Dewey.
- What happens when the seed description is too vague
  for Muti-Mind to produce a useful spec? Muti-Mind
  produces a spec with broader acceptance criteria and
  notes the ambiguity. The spec review checkpoint (if
  enabled) catches this. Without the checkpoint, the
  Divisor review stage may flag intent drift.
- What happens when the execution mode configuration
  is invalid (e.g., a stage set to an unknown mode)?
  The configuration is validated at workflow start.
  Invalid modes are rejected with a clear error
  message.
- What happens when the spec review checkpoint is
  enabled but the define stage is in `human` mode?
  The checkpoint is irrelevant -- the human is already
  involved in the define stage. The checkpoint is
  silently skipped.
- What happens when multiple workflows are started with
  autonomous define concurrently? Each workflow has its
  own state. Concurrent workflows are independent (per
  Constitution Principle I -- Autonomous Collaboration).

## Requirements *(mandatory)*

### Functional Requirements

**Configurable Execution Modes:**

- **FR-001**: The execution mode map MUST be
  configurable, allowing any stage to be set to
  `human` or `swarm` independently.
- **FR-002**: The default execution mode for the define
  stage MUST remain `human` for backward compatibility.
- **FR-003**: Changing the define stage to `swarm` MUST
  NOT affect the execution mode of any other stage.
- **FR-004**: The execution mode configuration MUST be
  validated at workflow start. Invalid modes MUST be
  rejected with a clear error message.

**Autonomous Specification:**

- **FR-005**: When the define stage runs in swarm mode,
  Muti-Mind MUST autonomously produce a feature
  specification with acceptance criteria, functional
  requirements, and success criteria.
- **FR-006**: Muti-Mind MUST use Dewey's semantic search
  to retrieve relevant cross-repo context when
  available (Tier 3 or Tier 2).
- **FR-007**: Muti-Mind MUST resolve ambiguities by
  querying Dewey for context rather than asking the
  human operator.
- **FR-008**: Muti-Mind MUST reference historical
  learning feedback from past workflows when producing
  acceptance criteria, if sufficient workflow records
  exist (3 or more).
- **FR-009**: When Dewey is unavailable (Tier 1),
  Muti-Mind MUST fall back to local backlog items and
  convention packs without blocking the workflow.

**Spec Review Checkpoint:**

- **FR-010**: The workflow MUST support an optional
  spec review checkpoint between the define and
  implement stages.
- **FR-011**: When enabled AND the define stage's
  execution mode is `swarm`, the spec review checkpoint
  MUST pause the workflow with `awaiting_human` status
  after the define stage completes. When the define
  stage is in `human` mode, the checkpoint is silently
  skipped (the human was already involved).
- **FR-012**: When the human advances past the spec
  review checkpoint, the workflow MUST resume and the
  implement stage MUST begin.
- **FR-013**: The spec review checkpoint MUST be
  disabled by default (the workflow advances directly
  from define to implement without pausing).

**Seed Command:**

- **FR-014**: A seed command MUST exist that starts a
  workflow with a one-sentence feature description and
  the define stage in swarm mode.
- **FR-015**: The seed command MUST create a backlog
  item from the seed description and start the workflow
  in one operation.
- **FR-016**: If no seed description is provided, the
  command MUST prompt the operator for one.

### Key Entities

- **Execution Mode Configuration**: A mapping of stage
  names to execution modes (`human` or `swarm`). Can
  be persisted as part of the workflow record or
  provided at workflow start time.
- **Spec Review Checkpoint**: A configurable pause
  point between the define and implement stages. When
  enabled, the workflow transitions to `awaiting_human`
  after define completes, allowing the human to review
  the specification before implementation begins.
- **Seed**: A one-sentence feature description that
  serves as input to Muti-Mind's autonomous
  specification workflow. Contains the operator's
  intent expressed in natural language.

## Assumptions

- Dewey is installed and indexed in the project (per
  Spec 015 -- `uf setup` handles this). The autonomous
  define degrades gracefully without Dewey (Tier 1
  fallback).
- The Swarm delegation workflow (Spec 012) is
  implemented -- execution modes, `StatusAwaitingHuman`,
  and checkpoint logic already exist.
- The `StageExecutionModeMap()` function currently
  returns hardcoded defaults. This spec makes it
  configurable.
- Muti-Mind's agent file (`.opencode/agents/muti-mind-po.md`)
  already includes Dewey tool usage and 3-tier
  degradation instructions (per Spec 015).
- The spec review checkpoint reuses the existing
  `StatusAwaitingHuman` mechanism from Spec 012. No new
  workflow status is introduced.
- The seed command is a convenience wrapper. It does
  not replace the existing `/workflow start` command
  -- both coexist.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A workflow with the define stage in swarm
  mode completes the define stage autonomously -- zero
  human interactions between seed and the accept
  checkpoint.
- **SC-002a**: When `NewWorkflow()` is called with
  `overrides={"define": "swarm"}`, the workflow
  instance's define stage has `execution_mode=swarm`
  and the workflow advances through define without
  requiring human intervention (Go-testable).
- **SC-002b**: The Muti-Mind agent file includes an
  "Autonomous Specification Workflow" section with
  steps for seed intake, Dewey context retrieval,
  self-clarification, and spec output (verified by
  manual review or drift-detection test on agent file
  section headings).
- **SC-003**: A complete end-to-end workflow (seed
  through reflect) completes with exactly 1 human
  decision point (accept) when the define stage is in
  swarm mode and the spec review checkpoint is disabled.
- **SC-004**: When the spec review checkpoint is
  enabled, the workflow pauses exactly twice: once after
  define (spec review) and once before accept.
- **SC-005**: The default configuration (define=human)
  produces identical behavior to pre-change workflows
  -- zero regressions.
- **SC-006**: All existing orchestration tests pass
  with zero changes when using the default execution
  mode configuration.
