# Feature Specification: Swarm Orchestration

**Feature Branch**: `008-swarm-orchestration`
**Created**: 2026-02-24
**Status**: Draft
**Input**: User description: "Define how the five heroes (Muti-Mind, Cobalt-Crush, Gaze, The Divisor, Mx F) work together as a swarm. Specify end-to-end workflows, artifact handoff protocols, the Swarm plugin integration, learning loops, and failure modes."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Feature Lifecycle Workflow: Idea to Deployment (Priority: P1)

A product team uses the Unbound Force swarm to take a feature from initial idea to deployed code. The workflow passes through all five heroes in a defined sequence: Muti-Mind defines the work, Cobalt-Crush implements it, Gaze validates it, The Divisor reviews it, and Mx F monitors the flow. Each handoff is an artifact exchange conforming to the inter-hero artifact protocol.

**Why this priority**: P1 because this is the primary value proposition of the Unbound Force swarm. If the heroes cannot work together in a defined workflow, they are just independent tools.

**Independent Test**: Can be tested by executing the full workflow on a small feature (e.g., "add a health check endpoint") in a test repository with all five heroes deployed, verifying each hero produces its expected artifact and the next hero consumes it.

**Acceptance Scenarios**:

1. **Given** Muti-Mind has a prioritized backlog item "Add health check endpoint" (BI-042), **When** the workflow begins, **Then** Muti-Mind initiates the speckit pipeline: `/specify` produces a spec, `/clarify` resolves ambiguities, `/plan` produces a plan, `/tasks` produces a task list.
2. **Given** a completed spec with tasks.md, **When** Cobalt-Crush begins implementation, **Then** it consumes the tasks.md, implements code phase by phase, and produces code artifacts (source files, tests) while consuming Gaze feedback after each phase.
3. **Given** Cobalt-Crush submits a PR, **When** Gaze runs, **Then** it produces a `quality-report` artifact with CRAP scores, contract coverage, and test results. The PR description references the originating backlog item (BI-042).
4. **Given** Gaze's quality-report shows acceptable quality, **When** The Divisor runs, **Then** it produces a `review-verdict` artifact. If APPROVED, the PR is ready to merge. If REQUEST_CHANGES, Cobalt-Crush addresses findings and iterates.
5. **Given** the PR is merged, **When** Muti-Mind reviews the increment, **Then** it produces an `acceptance-decision` artifact (accept/reject/conditional) based on the backlog item's acceptance criteria.
6. **Given** the workflow completes, **When** Mx F collects data, **Then** it records the full lifecycle metrics: time in each stage, iteration counts, quality scores, and updates the sprint state.

---

### User Story 2 - Artifact Handoff Protocol (Priority: P1)

Heroes exchange artifacts through a defined protocol. Each artifact conforms to the inter-hero envelope (Spec 002), is produced by exactly one hero, and can be consumed by one or more heroes. The protocol defines where artifacts are stored, how they are discovered, and how consumers handle missing or incompatible artifacts.

**Why this priority**: P1 because the artifact protocol is the communication backbone of the swarm. Without it, heroes cannot exchange information reliably.

**Independent Test**: Can be tested by having each hero produce its primary artifact type, storing it in the defined location, and verifying every consuming hero can discover and parse it.

**Acceptance Scenarios**:

1. **Given** the artifact storage convention, **When** a hero produces an artifact, **Then** it is written to `.unbound-force/artifacts/{artifact_type}/{timestamp}-{hero}.json` in the project repository.
2. **Given** an artifact in storage, **When** a consuming hero looks for artifacts of a type, **Then** it discovers all artifacts of that type sorted by timestamp (most recent first).
3. **Given** a Gaze `quality-report` artifact, **When** Mx F, Muti-Mind, and Cobalt-Crush each consume it, **Then** each extracts the information relevant to its role: Mx F extracts metrics, Muti-Mind extracts acceptance evidence, Cobalt-Crush extracts quality issues to address.
4. **Given** a hero that expects an artifact type that does not exist (e.g., Mx F looking for Gaze reports in a project without Gaze), **When** the consumer queries, **Then** it receives an empty result set and logs a debug-level notice (not an error).
5. **Given** an artifact with `schema_version: 2` and a consumer that only understands `schema_version: 1`, **When** the consumer reads it, **Then** it reports a version mismatch warning and either parses the compatible subset or skips the artifact gracefully.

---

### User Story 3 - Swarm Plugin Integration (Priority: P2)

The Unbound Force swarm integrates with the OpenCode Swarm plugin (swarmtools.ai) to enable multi-agent collaboration within a single OpenCode session. The Swarm plugin orchestrates hero agents, routes tasks to the appropriate hero, and facilitates cross-hero communication in real-time.

**Why this priority**: P2 because the Swarm plugin is an enhancement over manual hero invocation. The heroes work without Swarm (each invoked independently), but Swarm enables seamless collaboration.

**Independent Test**: Can be tested by configuring the Swarm plugin with all five hero agents and verifying it routes a product question to Muti-Mind, a coding task to Cobalt-Crush, a quality check to Gaze, a review request to The Divisor, and a process question to Mx F.

**Acceptance Scenarios**:

1. **Given** the Swarm plugin configured with hero agents, **When** a user asks "What should we build next?", **Then** Swarm routes the question to Muti-Mind (product owner).
2. **Given** the Swarm plugin, **When** a user asks "Implement the login feature from spec 003", **Then** Swarm routes the task to Cobalt-Crush (developer) with the spec context.
3. **Given** the Swarm plugin, **When** Cobalt-Crush completes implementation and a user says "Run quality checks", **Then** Swarm routes to Gaze with the changed files as context.
4. **Given** the Swarm plugin, **When** a user asks "Review this PR", **Then** Swarm invokes The Divisor's review council (three personas in parallel).
5. **Given** the Swarm plugin, **When** a user asks "How are we doing this sprint?", **Then** Swarm routes to Mx F for a metrics summary.
6. **Given** the Swarm plugin with learning enabled, **When** the swarm completes a feature lifecycle, **Then** the learning from the cycle (review patterns, quality trends, velocity data) improves future agent effectiveness.

---

### User Story 4 - Learning Loop and Continuous Improvement (Priority: P3)

The swarm implements a learning loop where completed cycles feed back into future work. Mx F's retrospective findings inform Muti-Mind's prioritization. Divisor review patterns inform Cobalt-Crush's coding conventions. Gaze's quality trends inform testing strategy. This loop is the mechanism for the swarm to become more effective over time.

**Why this priority**: P3 because the learning loop is what makes the swarm an "intelligent" system rather than a static tool chain. It depends on all heroes being functional and producing artifacts (US1, US2).

**Independent Test**: Can be tested by running two complete feature lifecycles and verifying that Mx F detects a pattern from the first cycle and it influences behavior in the second cycle.

**Acceptance Scenarios**:

1. **Given** Mx F identifies from Divisor data that "The Architect frequently requests error wrapping improvements" over 3 review cycles, **When** Mx F produces a coaching record, **Then** it recommends updating Cobalt-Crush's convention pack to include mandatory error wrapping.
2. **Given** the convention pack is updated based on Mx F's recommendation, **When** Cobalt-Crush writes new code, **Then** it proactively includes error wrapping without waiting for Divisor feedback.
3. **Given** Gaze reports showing declining contract coverage over 2 sprints, **When** Mx F flags this in a retrospective, **Then** Muti-Mind factors "improve test coverage for module X" into the next sprint's prioritization.
4. **Given** Muti-Mind's acceptance decisions showing a pattern of conditional acceptances due to missing edge case handling, **When** the pattern is surfaced, **Then** Cobalt-Crush's implementation workflow adds an explicit edge case review step after initial implementation.
5. **Given** the learning loop has operated for 5 sprints, **When** metrics are compared between sprint 1 and sprint 5, **Then** measurable improvements are visible in at least two metrics (e.g., review iterations decreased, quality scores increased).

---

### User Story 5 - Failure Modes and Graceful Degradation (Priority: P3)

The swarm handles failures gracefully: a hero being unavailable, producing unexpected output, or disagreeing with another hero does not halt the entire workflow. The orchestration defines fallback behavior for each failure mode.

**Why this priority**: P3 because graceful degradation ensures the swarm is robust in real-world usage where not all heroes may be deployed or functioning at all times.

**Independent Test**: Can be tested by removing one hero at a time from the workflow and verifying the remaining heroes continue to function with appropriate warnings.

**Acceptance Scenarios**:

1. **Given** Gaze is not deployed in a project, **When** the workflow reaches the validation stage, **Then** the orchestration skips Gaze validation, notes "quality validation unavailable," and proceeds to The Divisor review (which notes the absence of quality data in its review).
2. **Given** The Divisor review issues REQUEST_CHANGES but Cobalt-Crush has already iterated 3 times, **When** the maximum iteration count is reached, **Then** the orchestration escalates to manual review with a summary of unresolved findings.
3. **Given** Muti-Mind rejects an increment, **When** the rejection occurs, **Then** the workflow creates a new backlog item for the rejected work with the rejection rationale, and the original item remains in "review" status.
4. **Given** Mx F's metrics collection fails for one data source, **When** reporting occurs, **Then** Mx F reports metrics from available sources and notes which sources are missing.
5. **Given** two heroes produce contradictory guidance (e.g., Muti-Mind says "ship quickly" and Gaze says "quality is insufficient"), **When** the conflict occurs, **Then** the orchestration does not auto-resolve — it surfaces the conflict to the human operator with both perspectives and supporting data.

---

### Edge Cases

- What happens when the swarm workflow is started without any hero deployed? The orchestration MUST report "no heroes available" and provide installation guidance for each hero.
- What happens when artifacts from a previous workflow version are present but incompatible with current hero versions? Artifacts MUST be versioned and consumers MUST handle version mismatches (skip or parse compatible subset).
- What happens when the Swarm plugin is not installed? Each hero MUST be invocable independently via its own `/command`. The swarm workflow can be executed manually step by step.
- What happens when two workflows are running concurrently on different features? Each workflow MUST be scoped to its feature branch and backlog item. Artifacts MUST include the branch/item context to prevent cross-contamination.
- What happens when a hero produces an artifact that no other hero consumes? The artifact is still stored per the protocol. Unused artifacts do not cause errors.
- What happens when the learning loop produces a recommendation that the team disagrees with? Mx F's recommendations are proposals, not mandates. The team may dismiss a recommendation with documented rationale.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The swarm orchestration MUST define the complete feature lifecycle workflow: Muti-Mind (define) -> Cobalt-Crush (implement) -> Gaze (validate) -> The Divisor (review) -> Muti-Mind (accept) -> Mx F (measure).
- **FR-002**: Each workflow stage MUST produce and/or consume artifacts conforming to the inter-hero artifact envelope (Spec 002).
- **FR-003**: Artifact storage MUST follow the convention: `.unbound-force/artifacts/{artifact_type}/{timestamp}-{hero}.json`.
- **FR-004**: Artifact discovery MUST support querying by type, hero, and time range, returning results sorted by timestamp.
- **FR-005**: The orchestration MUST define a Swarm plugin configuration that maps hero agents to their roles and routing rules.
- **FR-006**: The Swarm plugin configuration MUST support natural language routing: product questions -> Muti-Mind, coding tasks -> Cobalt-Crush, quality checks -> Gaze, review requests -> The Divisor, process questions -> Mx F.
- **FR-007**: The learning loop MUST define how Mx F's coaching records feed back into: Muti-Mind's prioritization, Cobalt-Crush's convention packs, Gaze's testing strategy, and The Divisor's review criteria.
- **FR-008**: The orchestration MUST define failure modes and fallback behavior for: hero unavailable, artifact missing, maximum iteration reached, acceptance rejected, and inter-hero contradiction.
- **FR-009**: Each hero MUST function independently (per Org Constitution Principle II) — the swarm orchestration is additive, not required.
- **FR-010**: Concurrent workflows on different feature branches MUST be isolated — artifacts MUST include branch/backlog-item context.
- **FR-011**: The orchestration MUST define escalation protocols for unresolvable conflicts: surface to human operator with full context from all involved heroes.
- **FR-012**: The workflow MUST be executable both through the Swarm plugin (automated routing) and manually (hero-by-hero invocation via individual `/commands`).
- **FR-013**: The orchestration MUST produce a `workflow-record` artifact that captures the complete lifecycle of a feature: all stages, all artifacts produced, all decisions made, total elapsed time.
- **FR-014**: The workflow-record MUST be consumable by Mx F for lifecycle metrics and by Muti-Mind for velocity tracking.

### Key Entities

- **Workflow Instance**: A single execution of the feature lifecycle. Attributes: workflow_id, feature_branch, backlog_item_id, stages[], current_stage, started_at, completed_at, status (active/completed/failed/escalated).
- **Workflow Stage**: One step in the lifecycle. Attributes: stage_name (define/implement/validate/review/accept/measure), hero, status (pending/active/completed/skipped/failed), artifacts_produced[], artifacts_consumed[], started_at, completed_at.
- **Swarm Plugin Configuration**: Routing rules for the Swarm plugin. Attributes: heroes[] (name, agent_file, role, routing_patterns[]), default_hero, escalation_rules.
- **Learning Feedback**: A recommendation from one hero to another based on pattern analysis. Attributes: source_hero, target_hero, pattern_observed, recommendation, supporting_data{}, status (proposed/accepted/rejected/implemented).
- **Workflow Record**: The complete history of a feature lifecycle. Attributes: workflow_id, backlog_item_id, stages[], artifacts[], decisions[], total_elapsed_time, outcome (shipped/rejected/abandoned).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The feature lifecycle workflow is documented with all six stages, their hero assignments, and the artifacts exchanged at each handoff.
- **SC-002**: A full feature lifecycle can be executed manually (hero-by-hero) in a test repository, producing the expected artifacts at each stage.
- **SC-003**: The Swarm plugin configuration correctly routes 5 different natural language queries to the appropriate hero.
- **SC-004**: Removing one hero from the swarm produces a degraded workflow with appropriate warnings but does not halt the remaining heroes.
- **SC-005**: The learning loop produces at least one actionable recommendation after 3 completed feature lifecycles.
- **SC-006**: The workflow-record artifact captures the complete lifecycle and is parseable by Mx F for metrics.
- **SC-007**: Concurrent workflows on different branches produce isolated artifacts with no cross-contamination.
- **SC-008**: The orchestration handles the acceptance-rejection failure mode correctly: creates a new backlog item, preserves the original item state.

## Dependencies

### Prerequisites

- **Spec 001** (Org Constitution): Defines the three principles that govern swarm behavior (Autonomous Collaboration, Composability First, Observable Quality).
- **Spec 002** (Hero Interface Contract): Defines the artifact envelope and inter-hero protocol used by the orchestration.
- **Specs 004-007** (Hero Architectures): Defines each hero's capabilities, artifact types, and integration points.

### Downstream Dependents

- **Spec 009** (Shared Data Model): Defines the `workflow-record` and `learning-feedback` JSON schemas.

```
Feature Lifecycle Workflow

   Muti-Mind          Cobalt-Crush         Gaze            The Divisor        Muti-Mind          Mx F
  (Define)            (Implement)        (Validate)        (Review)           (Accept)          (Measure)
┌──────────┐       ┌──────────────┐    ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────┐
│ Backlog  │       │ Code +       │    │ Quality      │  │ Review       │  │ Acceptance   │  │ Metrics  │
│ Item     │──────>│ Tests        │───>│ Report       │─>│ Verdict      │─>│ Decision     │─>│ Snapshot │
│ + Spec   │       │              │    │              │  │              │  │              │  │          │
└──────────┘       └──────┬───────┘    └──────────────┘  └──────┬───────┘  └──────────────┘  └──────────┘
                          │                                      │
                          │◄─────────── REQUEST CHANGES ─────────┘
                          │          (iteration loop)
                          │
                          │◄─── Gaze feedback ───┘
                              (quality loop)

Learning Loop (Mx F → All Heroes):
  Mx F patterns ──> Muti-Mind (prioritization)
  Mx F patterns ──> Cobalt-Crush (conventions)
  Mx F patterns ──> Gaze (test strategy)
  Mx F patterns ──> The Divisor (review criteria)
```
