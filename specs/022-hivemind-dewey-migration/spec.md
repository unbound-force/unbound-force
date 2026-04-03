# Feature Specification: Hivemind-to-Dewey Memory Migration

**Feature Branch**: `022-hivemind-dewey-migration`  
**Created**: 2026-04-03  
**Status**: Ready  
**Input**: User description: "issue #76 — Migrate /unleash and Divisor agents from Hivemind to Dewey (Spec 021 FR-012–FR-015)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Unified Learning Storage in Autonomous Pipeline (Priority: P1)

When an engineer runs the autonomous pipeline (`/unleash`), the
retrospective step stores learnings — patterns discovered, decisions
made, failures encountered — in Dewey, the project's unified memory
system. Today this step uses Hivemind, a separate tool, which
fragments the knowledge base across two systems. After this change,
all learnings flow into the same system that already handles semantic
search and knowledge retrieval, eliminating the split.

**Why this priority**: The autonomous pipeline is the primary
workflow that generates organizational learnings. If learnings are
stored in Hivemind while retrieval happens through Dewey, knowledge
written during `/unleash` is invisible to agents that search
Dewey — creating a silent data loss of institutional memory.

**Independent Test**: Run the autonomous pipeline end-to-end on a
small feature. Verify the retrospective step stores learnings in
Dewey. Confirm the stored learnings are retrievable via Dewey's
semantic search in a subsequent session.

**Acceptance Scenarios**:

1. **Given** an engineer completes an `/unleash` run that finishes
   the retrospective step, **When** the retrospective stores a
   learning, **Then** the learning is persisted in Dewey and
   retrievable via semantic search.
2. **Given** an engineer completes an `/unleash` run, **When** a
   Divisor agent later reviews related code, **Then** the prior
   learnings from the retrospective appear in the agent's prior
   learnings step.
3. **Given** the retrospective step attempts to store a learning,
   **When** Dewey's learning storage capability is not yet available
   (dependency not landed), **Then** the pipeline completes without
   error and logs a warning that learning storage was skipped.

---

### User Story 2 — Unified Learning Retrieval in Review Agents (Priority: P1)

When a Divisor review agent (any of the five personas — Guard,
Architect, Adversary, SRE, Testing) prepares for a code review, it
searches for prior learnings relevant to the files and patterns under
review. Today this step uses Hivemind's find capability. After this
change, all five agents retrieve prior learnings through Dewey's
semantic search, which already indexes the project's knowledge base.

**Why this priority**: The five Divisor agents run on every review.
If they search Hivemind while new learnings are stored in Dewey
(per US-1), the agents will miss all recently captured knowledge.
This is a data consistency issue that directly degrades review
quality.

**Independent Test**: Trigger a review with each of the five Divisor
personas on a file with known prior learnings stored in Dewey.
Verify each agent surfaces the relevant learnings in its Prior
Learnings step.

**Acceptance Scenarios**:

1. **Given** a Divisor agent begins a code review, **When** prior
   learnings relevant to the reviewed files exist in Dewey,
   **Then** the agent retrieves and displays those learnings in its
   Prior Learnings output.
2. **Given** a Divisor agent begins a code review, **When** Dewey
   is unavailable, **Then** the agent proceeds with the review
   without error, noting that prior learnings could not be
   retrieved.
3. **Given** a Divisor agent begins a code review, **When** no
   prior learnings match the reviewed files, **Then** the agent
   proceeds normally without displaying a Prior Learnings section.

---

### User Story 3 — Graceful Degradation (Priority: P2)

When Dewey is temporarily unavailable — not installed, not running,
or missing the learning storage capability (e.g., dependency
dewey#25 has not yet landed) — all learning storage and retrieval
operations degrade gracefully. The pipeline and review agents
continue operating without error, warning the engineer that the
memory system is unavailable.

**Why this priority**: Dewey availability cannot be guaranteed in
all environments (CI, new contributor machines, offline use). The
agent ecosystem must remain functional without Dewey, producing
warnings rather than failures.

**Independent Test**: Temporarily disable Dewey (remove from PATH
or stop the server). Run `/unleash` and trigger a Divisor review.
Verify both complete successfully with appropriate warnings.

**Acceptance Scenarios**:

1. **Given** Dewey is not installed or not in PATH, **When** the
   autonomous pipeline retrospective attempts to store a learning,
   **Then** the pipeline completes successfully and logs a warning
   that learning storage was skipped.
2. **Given** Dewey is not installed or not in PATH, **When** a
   Divisor agent attempts to retrieve prior learnings, **Then** the
   agent proceeds with its review and notes that prior learnings
   are unavailable.
3. **Given** Dewey is installed but the learning storage tool has
   not been implemented yet, **When** the retrospective attempts
   to store a learning, **Then** the pipeline completes
   successfully and logs a warning that the tool is unavailable.

---

### User Story 4 — Documentation Accuracy (Priority: P2)

Project documentation accurately describes Dewey as the unified
memory layer for the agent ecosystem. The prior framing — "Dewey
complements Hivemind" (from Spec 020) — is superseded by "Dewey is
the single memory system." Engineers reading the documentation
understand that all learning storage and retrieval flows through
Dewey.

**Why this priority**: Accurate documentation prevents engineers
from manually using Hivemind for tasks that should go through
Dewey, avoiding further knowledge fragmentation.

**Independent Test**: Review the project documentation and verify
that Dewey is described as the unified (not complementary) memory
layer, with no residual references to Hivemind as an active
component of the learning workflow.

**Acceptance Scenarios**:

1. **Given** an engineer reads the project documentation, **When**
   they look for how learnings are stored and retrieved, **Then**
   the documentation describes Dewey as the unified memory system.
2. **Given** the prior documentation described Dewey as
   "complementing" Hivemind, **When** this change lands, **Then**
   all such references are updated to reflect Dewey as the single
   memory layer.

---

### User Story 5 — Scaffold Consistency (Priority: P3)

When the scaffold engine deploys agent and command files to a target
repository via `uf init`, the deployed files reflect the
Dewey-unified memory configuration. The scaffold asset copies remain
synchronized with the canonical source files modified in this
change.

**Why this priority**: Without synchronized scaffold assets, newly
initialized repositories would receive stale agent files that still
reference Hivemind, creating drift between existing and new
projects.

**Independent Test**: Run `uf init` in a fresh temporary directory.
Verify the scaffolded agent and command files reference Dewey for
learning storage and retrieval, with no Hivemind references in
those flows.

**Acceptance Scenarios**:

1. **Given** the canonical agent and command files have been updated
   to use Dewey, **When** the scaffold engine deploys these files,
   **Then** the deployed copies match the canonical sources exactly.
2. **Given** an engineer runs `uf init` after this change, **When**
   the scaffold completes, **Then** the deployed Divisor agent files
   and autonomous pipeline command reference Dewey, not Hivemind,
   for learning operations.

---

### Edge Cases

- What happens when Dewey is installed but its semantic search
  returns an error (not "unavailable" but a runtime failure)?
  The agent SHOULD catch the error, warn the user, and proceed
  without prior learnings rather than aborting the review.
- What happens when the Dewey learning storage tool exists but
  returns a write error? The retrospective step SHOULD warn and
  continue rather than failing the entire pipeline.
- What happens when a scaffold drift detection test detects
  mismatches between canonical files and scaffold assets? The test
  SHOULD fail the build, forcing the engineer to synchronize before
  merging.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The autonomous pipeline retrospective step MUST store
  learnings using Dewey's learning storage capability instead of
  Hivemind's store capability (per Spec 021 FR-012: "Learning
  storage MUST use `dewey_store_learning`").
- **FR-002**: All five Divisor review agents' Prior Learnings step
  MUST retrieve learnings using Dewey's semantic search instead of
  Hivemind's find capability (per Spec 021 FR-013: "Learning
  retrieval MUST use `dewey_semantic_search`").
- **FR-003**: All learning storage operations MUST gracefully
  degrade when Dewey is unavailable — completing without error
  and warning the user (per Spec 021 FR-014: "All Dewey
  operations MUST degrade gracefully when unavailable").
- **FR-004**: All learning retrieval operations MUST gracefully
  degrade when Dewey is unavailable — proceeding without prior
  learnings and noting the absence (per Spec 021 FR-014: "All
  Dewey operations MUST degrade gracefully when unavailable").
- **FR-005**: Project documentation MUST describe Dewey as the
  unified memory layer, superseding the "Dewey complements
  Hivemind" framing from Spec 020 (per Spec 021 FR-015:
  "Documentation MUST describe Dewey as the unified memory
  layer").
- **FR-006**: All modified agent and command files MUST have their
  scaffold asset copies synchronized to prevent drift between
  source files and deployed copies.
- **FR-007**: Existing drift detection tests MUST continue to pass
  after scaffold assets are updated.

### Key Entities

- **Learning**: A knowledge artifact produced during a
  retrospective — captures patterns, decisions, failure modes,
  and contextual metadata. Produced by the autonomous pipeline,
  consumed by review agents.
- **Prior Learnings**: A set of semantically relevant learnings
  retrieved before a code review, providing historical context
  to review agents.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero Hivemind tool references remain in the
  autonomous pipeline's retrospective step and all five Divisor
  review agents' prior learnings step — verified by a text search
  across all agent and command files.
- **SC-002**: Learnings stored during an autonomous pipeline run
  are retrievable by a Divisor review agent in a subsequent
  session within 30 seconds of storage.
- **SC-003**: When Dewey is unavailable, 100% of autonomous
  pipeline runs and Divisor reviews complete without error —
  verified by running both workflows with Dewey removed from PATH.
- **SC-004**: Project documentation contains zero references to
  Hivemind as an active component of the learning workflow.
- **SC-005**: All scaffold drift detection tests pass after the
  migration — verified by running the project's full test suite.
- **SC-006**: Newly scaffolded repositories (via `uf init`)
  contain zero Hivemind references in learning-related agent and
  command files.

## Dependencies & Assumptions

### Dependencies

- **Dewey learning storage** (dewey#25): Dewey must implement the
  learning storage tool for US-1 to be fully functional. Until
  then, the graceful degradation path (US-3) activates
  automatically.
- **Dewey semantic search**: Already available. US-2 (Divisor agent
  retrieval) can land immediately without waiting for dewey#25.
- **Spec 021 — Dewey Unified Memory**: This spec implements Phase 2
  requirements (FR-012 through FR-015) from the parent spec.

### Assumptions

- Dewey's semantic search returns results in the same relevance
  quality as Hivemind's find for equivalent queries. If semantic
  search quality differs, agents will need tuned queries — this
  is out of scope for this spec but noted as a risk.
- The graceful degradation pattern follows the existing 3-tier
  pattern documented in AGENTS.md: Full Dewey, Graph-only,
  No Dewey.
- Scaffold drift detection tests already exist and will
  automatically catch any unsynchronized assets.
