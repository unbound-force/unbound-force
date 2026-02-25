# Feature Specification: Muti-Mind Architecture (Product Owner)

**Feature Branch**: `004-muti-mind-architecture`
**Created**: 2026-02-24
**Status**: Draft
**Input**: User description: "Design the architecture for Muti-Mind, the Product Owner hero. Muti-Mind is the Vision Keeper and Prioritization Engine — the voice of the user within the swarm. It includes an AI persona, a backlog management CLI tool, GitHub Issues/Projects integration, spec management integration with speckit, and acceptance authority capabilities."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - AI Persona and Decision Framework (Priority: P1)

A developer or agent working on a feature consults Muti-Mind for product decisions. Muti-Mind operates as an AI agent persona with a clearly defined decision-making framework, communication style, and knowledge base. When asked "Why are we building this?" or "What should the acceptance criteria be?", Muti-Mind provides authoritative, consistent answers grounded in the documented product vision and user requirements.

**Why this priority**: P1 because the persona is the foundation of Muti-Mind. Without a well-defined AI persona, Muti-Mind is just tooling without judgment. The persona drives how all other capabilities are used.

**Independent Test**: Can be tested by presenting Muti-Mind with a set of product questions (feature priority, scope decisions, acceptance criteria clarification) and verifying the answers are consistent with the documented product vision and do not contradict prior decisions.

**Acceptance Scenarios**:

1. **Given** a Muti-Mind agent configuration deployed in a project, **When** a developer asks "Should we add feature X?", **Then** Muti-Mind evaluates the request against the product backlog priorities, product goals, and constitution principles, providing a reasoned recommendation with citations to relevant backlog items or specs.
2. **Given** a product goal document and backlog, **When** Muti-Mind is asked to clarify acceptance criteria for a user story, **Then** it produces Given/When/Then acceptance scenarios that are consistent with the documented product vision.
3. **Given** conflicting inputs (e.g., a developer wants to add scope, but the sprint goal is narrow), **When** Muti-Mind is consulted, **Then** it recommends deferring the scope addition to the backlog with a clear rationale, citing prioritization principles.
4. **Given** a Muti-Mind deployment in project A and a separate deployment in project B, **When** both are queried, **Then** each provides answers grounded in its own project's context (not contaminated by the other project's data).

---

### User Story 2 - Backlog Management CLI (Priority: P1)

A product owner or developer manages the product backlog through a Muti-Mind CLI tool. The CLI supports creating, reading, updating, and deleting backlog items (user stories, epics, tasks). Items are stored locally in a structured format and can be synced with GitHub Issues/Projects.

**Why this priority**: P1 because the backlog is the primary artifact Muti-Mind manages. Without a backlog management tool, there is no structured way to capture, prioritize, or communicate product requirements.

**Independent Test**: Can be tested by creating a backlog with several items, prioritizing them, and verifying the output reflects the correct priority order and item details.

**Acceptance Scenarios**:

1. **Given** an empty project, **When** `muti-mind backlog init` is run, **Then** it creates a `.muti-mind/backlog/` directory with an empty backlog index file.
2. **Given** an initialized backlog, **When** `muti-mind backlog add --type story --title "User login" --priority P1 --description "..."` is run, **Then** a new backlog item file is created with a unique ID, the specified attributes, and a creation timestamp.
3. **Given** a backlog with 10 items of varying priorities, **When** `muti-mind backlog list` is run, **Then** items are displayed in priority order (P1 first) with: ID, title, type, priority, status, and sprint assignment (if any).
4. **Given** a backlog item, **When** `muti-mind backlog update BI-003 --priority P2 --sprint "Sprint 3"` is run, **Then** the item's priority and sprint assignment are updated and the change is logged.
5. **Given** a backlog item, **When** `muti-mind backlog show BI-003` is run, **Then** the full item details are displayed including: title, description, type, priority, status, acceptance criteria, sprint, creation date, last modified date, and related specs.

---

### User Story 3 - Priority Scoring Engine (Priority: P2)

Muti-Mind provides an AI-assisted priority scoring algorithm that evaluates backlog items across multiple dimensions: business value, risk, dependencies, urgency, and effort. The scoring produces a ranked backlog and provides transparency into why each item received its score.

**Why this priority**: P2 because manual prioritization works initially, but as backlogs grow, an automated scoring engine ensures consistency and surfaces non-obvious priority orderings.

**Independent Test**: Can be tested by creating a backlog with items that have known relative priorities and verifying the scoring engine produces a ranking that matches the expected order.

**Acceptance Scenarios**:

1. **Given** a backlog item with value, risk, dependency, urgency, and effort attributes, **When** `muti-mind prioritize` is run, **Then** each item receives a composite priority score and the backlog is re-ranked by score.
2. **Given** a prioritized backlog, **When** a user inspects an item's score, **Then** the score breakdown shows the contribution of each dimension (e.g., "value: 8/10, risk: 3/10, dependencies: 2 blocking items, urgency: high, effort: medium -> composite: 82").
3. **Given** two items with similar scores, **When** the prioritization runs, **Then** Muti-Mind flags them for manual tiebreaking and provides a recommendation based on dependency analysis.
4. **Given** a backlog item that blocks other items, **When** prioritization runs, **Then** the blocking item's score is boosted proportional to the aggregate value of the items it blocks.

---

### User Story 4 - GitHub Issues and Projects Synchronization (Priority: P2)

Muti-Mind synchronizes the local backlog with GitHub Issues and GitHub Projects. Items can be pushed to GitHub (creating or updating Issues), pulled from GitHub (importing existing Issues), and two-way synced to maintain consistency. Labels, milestones, and project board columns map to backlog attributes.

**Why this priority**: P2 because GitHub is the collaboration platform for the Unbound Force organization. Without sync, the backlog exists only locally and cannot be shared with team members or other heroes.

**Independent Test**: Can be tested by creating a local backlog item, syncing it to GitHub, verifying the Issue is created, modifying the Issue on GitHub, syncing back, and verifying the local item is updated.

**Acceptance Scenarios**:

1. **Given** a local backlog item, **When** `muti-mind sync push` is run, **Then** a GitHub Issue is created with: title from backlog title, body from description and acceptance criteria, labels from type and priority (e.g., `type:story`, `priority:P1`), and milestone from sprint.
2. **Given** a GitHub Issue not in the local backlog, **When** `muti-mind sync pull` is run, **Then** the Issue is imported as a new backlog item with attributes mapped from labels and milestone.
3. **Given** a synced item modified on both sides, **When** `muti-mind sync` is run, **Then** conflicts are detected, listed, and the user is prompted to choose the local or remote version for each conflicting field.
4. **Given** a GitHub Project board, **When** `muti-mind sync project --project "Sprint Board"` is run, **Then** backlog items are mapped to project columns based on status (e.g., "To Do", "In Progress", "Done").
5. **Given** a synced backlog, **When** `muti-mind sync status` is run, **Then** it reports: items in sync, items modified locally, items modified remotely, items only local, items only remote.

---

### User Story 5 - Speckit Integration and Acceptance Authority (Priority: P3)

Muti-Mind integrates with the speckit pipeline to drive the specification and acceptance phases. As the Product Owner, Muti-Mind is responsible for initiating the `specify` and `clarify` phases, validating specs against the product vision, and serving as the acceptance authority when Gaze reports test results.

**Why this priority**: P3 because this represents the full integration of Muti-Mind into the swarm workflow. It depends on the persona (US1), backlog (US2), and GitHub sync (US4) being functional.

**Independent Test**: Can be tested by having Muti-Mind review a completed spec against the product backlog and produce an acceptance/rejection decision with rationale.

**Acceptance Scenarios**:

1. **Given** a backlog item tagged for implementation, **When** Muti-Mind initiates the speckit pipeline, **Then** it invokes `/specify` with the backlog item's description, acceptance criteria, and priority as input context.
2. **Given** a completed spec, **When** Muti-Mind runs `/clarify`, **Then** it asks clarification questions grounded in the product vision and backlog priorities (not just technical ambiguity).
3. **Given** a Gaze quality report for a completed feature, **When** Muti-Mind reviews it, **Then** it produces an acceptance decision (accept/reject/conditionally accept) based on whether the acceptance criteria from the originating backlog item are satisfied.
4. **Given** an accepted feature, **When** Muti-Mind updates the backlog, **Then** the backlog item is marked as "Done" and the acceptance report is linked to the item.
5. **Given** a rejected feature, **When** Muti-Mind produces a rejection report, **Then** it specifies which acceptance criteria were not met, with references to the Gaze report findings.

---

### User Story 6 - User Story Generation (Priority: P3)

Muti-Mind generates user stories from high-level product goals or feature descriptions. Given a brief description (e.g., "users need to export data"), Muti-Mind produces structured user stories with acceptance criteria, priority recommendations, and effort estimates.

**Why this priority**: P3 because this is an AI-enhanced productivity feature. Manual story writing works, but AI generation accelerates the refinement process.

**Independent Test**: Can be tested by providing a high-level feature description and verifying the generated stories are well-formed, independently testable, and include Given/When/Then acceptance criteria.

**Acceptance Scenarios**:

1. **Given** a high-level goal "users need to export their data in CSV and PDF formats", **When** `muti-mind generate stories --goal "..."` is run, **Then** it produces at least two user stories (one per format) with titles, descriptions, acceptance criteria, and priority recommendations.
2. **Given** a generated set of stories, **When** a reviewer inspects them, **Then** each story follows the speckit user story format: title, priority, independent test description, and Given/When/Then acceptance scenarios.
3. **Given** a product backlog with existing items, **When** stories are generated, **Then** Muti-Mind checks for overlap with existing items and flags potential duplicates.

---

### Edge Cases

- What happens when `muti-mind sync push` is run without GitHub credentials configured? The command MUST fail with a clear error message directing the user to configure authentication (e.g., `gh auth login` or `GITHUB_TOKEN`).
- What happens when a backlog item has no acceptance criteria? Muti-Mind MUST warn that the item is not "Ready" for implementation and recommend running the `clarify` phase.
- What happens when prioritization is run on an empty backlog? The command MUST report "no items to prioritize" and exit cleanly.
- What happens when GitHub sync encounters rate limiting? The command MUST detect HTTP 429 responses, report the rate limit, and suggest retrying after the reset time.
- What happens when two users sync conflicting changes simultaneously? The second sync MUST detect the conflict and refuse to overwrite, presenting a merge resolution prompt.
- What happens when a backlog item references a spec that no longer exists? Muti-Mind MUST warn about the broken reference but not delete the backlog item.
- What happens when a story is generated but the user rejects it? Generated stories MUST be presented as proposals that require explicit user approval before being added to the backlog.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Muti-Mind MUST provide an AI agent persona with a documented decision-making framework, communication style, and behavioral constraints.
- **FR-002**: The agent persona MUST be deployable as an OpenCode agent file (`muti-mind-po.md`) installable via `muti-mind init`.
- **FR-003**: Muti-Mind MUST provide a CLI tool (`muti-mind`) for backlog management with subcommands: `init`, `backlog` (add/list/show/update/delete), `prioritize`, `sync`, `generate`, and `accept`.
- **FR-004**: Backlog items MUST be stored as individual files in `.muti-mind/backlog/` in a human-readable format (YAML front matter + Markdown body).
- **FR-005**: Each backlog item MUST have a unique identifier (BI-NNN), type (epic/story/task/bug), priority (P1-P5), status (draft/ready/in-progress/review/done/cancelled), and timestamps (created, modified).
- **FR-006**: The priority scoring engine MUST evaluate items across at least five dimensions: business value, risk, dependencies, urgency, and effort.
- **FR-007**: The priority score MUST be transparent: each dimension's contribution to the composite score MUST be visible.
- **FR-008**: GitHub sync MUST support push (local -> GitHub Issues), pull (GitHub Issues -> local), and bidirectional sync with conflict detection.
- **FR-009**: GitHub sync MUST map backlog attributes to GitHub primitives: type -> labels, priority -> labels, sprint -> milestones, status -> project board columns.
- **FR-010**: Muti-Mind MUST integrate with the speckit pipeline: it MUST be able to invoke `/specify` and `/clarify` with backlog item context.
- **FR-011**: Muti-Mind MUST serve as the acceptance authority: given a Gaze quality report and the originating backlog item's acceptance criteria, it MUST produce an accept/reject/conditional decision.
- **FR-012**: The acceptance decision MUST be machine-parseable (JSON artifact conforming to the inter-hero artifact envelope from Spec 002).
- **FR-013**: Muti-Mind MUST generate user stories from high-level goals, producing output in the speckit user story format.
- **FR-014**: Generated stories MUST be proposals requiring explicit user approval before addition to the backlog.
- **FR-015**: Muti-Mind MUST produce a `backlog-item` artifact type (per Spec 002) for consumption by other heroes (Mx F for metrics, The Divisor for intent verification).
- **FR-016**: Muti-Mind MUST produce an `acceptance-decision` artifact type for consumption by other heroes.
- **FR-017**: The CLI MUST support `--format json` and `--format text` for all output commands (per Org Constitution Principle III).
- **FR-018**: Muti-Mind MUST conform to the Hero Interface Contract (Spec 002): standard repo structure, hero manifest, speckit integration, OpenCode agent/command standards.

### Key Entities

- **Backlog Item**: A unit of product work. Attributes: id (BI-NNN), title, description (markdown), type (epic/story/task/bug), priority (P1-P5), status (draft/ready/in-progress/review/done/cancelled), acceptance_criteria[] (Given/When/Then), sprint (string, optional), effort_estimate (t-shirt size or story points), dependencies[] (other BI IDs), related_specs[] (spec paths), github_issue_number (int, optional), created_at, modified_at.
- **Priority Score**: Composite ranking of a backlog item. Attributes: item_id, business_value (0-10), risk (0-10), dependency_weight (computed), urgency (low/medium/high/critical), effort (XS/S/M/L/XL), composite_score (0-100), rank (position in ordered backlog).
- **Acceptance Decision**: Product Owner verdict on a completed increment. Attributes: item_id, decision (accept/reject/conditional), rationale (markdown), criteria_met[] (which acceptance criteria passed), criteria_failed[] (which failed), gaze_report_ref (path to quality report), decided_at.
- **Sync State**: Tracking of local-to-GitHub synchronization. Attributes: item_id, github_issue_number, last_synced_at, local_hash, remote_hash, sync_status (synced/local-modified/remote-modified/conflict/local-only/remote-only).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The Muti-Mind agent persona produces consistent, vision-aligned answers to 10 sample product questions with zero contradictions.
- **SC-002**: A backlog with 20 items can be created, prioritized, and listed in under 30 seconds via the CLI.
- **SC-003**: Priority scoring produces a ranking for 20 items with transparent dimension breakdowns for each item.
- **SC-004**: GitHub sync round-trips a backlog item (create local -> push -> modify on GitHub -> pull) with zero data loss.
- **SC-005**: Speckit integration successfully invokes `/specify` with backlog item context and produces a valid spec draft.
- **SC-006**: The acceptance authority correctly identifies which acceptance criteria pass/fail given a Gaze quality report, producing a machine-parseable decision artifact.
- **SC-007**: User story generation from a high-level goal produces at least two well-formed stories with Given/When/Then criteria.
- **SC-008**: All CLI output commands support `--format json` and the JSON output validates against the artifact envelope schema.

## Dependencies

### Prerequisites

- **Spec 001** (Org Constitution): Muti-Mind's persona and decisions must align with org principles.
- **Spec 002** (Hero Interface Contract): Muti-Mind must conform to the hero manifest, artifact envelope, and naming conventions.
- **Spec 003** (Speckit Framework): Muti-Mind integrates with the speckit pipeline for spec management.

### Downstream Dependents

- **Spec 007** (Mx F Architecture): Mx F consumes backlog items and acceptance decisions for metrics.
- **Spec 008** (Swarm Orchestration): Muti-Mind drives the "feature from idea to deployment" workflow.
- **Spec 009** (Shared Data Model): Defines the `backlog-item` and `acceptance-decision` JSON schemas.

### External Dependencies

- **GitHub API**: GitHub Issues, Projects, Labels, Milestones APIs for sync functionality.
- **GitHub CLI (`gh`)**: May be used as a dependency for GitHub API interaction.
