# Feature Specification: Cobalt-Crush Architecture (Developer)

**Feature Branch**: `006-cobalt-crush-architecture`
**Created**: 2026-02-24
**Status**: Draft
**Input**: User description: "Design the architecture for Cobalt-Crush, the Developer hero. Cobalt-Crush is the Engineering Core and Adaptive Implementation Engine. It includes an AI agent persona with coding conventions, templates, and integration with Gaze (test feedback) and The Divisor (review feedback) feedback loops."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - AI Persona and Coding Philosophy (Priority: P1)

A developer using OpenCode deploys the Cobalt-Crush agent persona to guide their coding workflow. Cobalt-Crush operates as an AI developer agent with a clearly defined engineering philosophy: clean code principles, SOLID adherence, test-driven awareness, and a relentless focus on shipping quality code within the CI/CD paradigm. When asked to implement a feature, Cobalt-Crush consults the active spec, follows the tasks.md checklist, and produces code that is designed to pass Gaze's validation and The Divisor's review.

**Why this priority**: P1 because the persona defines how Cobalt-Crush approaches every engineering decision. Without it, Cobalt-Crush is an undifferentiated code generator.

**Independent Test**: Can be tested by deploying the Cobalt-Crush agent in a project and asking it to implement a feature from a spec, verifying the output follows clean code principles, addresses acceptance criteria, and produces testable code.

**Acceptance Scenarios**:

1. **Given** a Cobalt-Crush agent deployed in a project with an active spec, **When** asked to implement a user story, **Then** it reads the spec's acceptance criteria, consults the tasks.md for the implementation plan, and produces code that addresses each criterion.
2. **Given** a coding task, **When** Cobalt-Crush writes code, **Then** it follows the project's coding conventions (from the convention pack or AGENTS.md), including naming, formatting, error handling, and documentation patterns.
3. **Given** a complex implementation decision (e.g., choosing between two design patterns), **When** Cobalt-Crush makes the decision, **Then** it documents the rationale in a code comment or design document and cites the relevant principle (e.g., "Chose Strategy pattern over switch statement per SOLID Open/Closed Principle").
4. **Given** Cobalt-Crush produces code, **When** the code is inspected, **Then** it includes appropriate test hooks (exported test helpers, interface abstractions, dependency injection) to facilitate Gaze's validation.

---

### User Story 2 - Coding Standards Framework (Priority: P1)

Cobalt-Crush defines and enforces a coding standards framework with language-agnostic principles and language-specific convention packs. The framework is a structured document that Cobalt-Crush consults during implementation and that The Divisor references during review, ensuring consistency between what is written and what is approved.

**Why this priority**: P1 because coding standards are the shared language between the developer (Cobalt-Crush) and the reviewer (The Divisor). They must be defined identically to prevent perpetual review-rework cycles.

**Independent Test**: Can be tested by verifying that Cobalt-Crush's coding standards for a given language are identical to The Divisor's convention pack for that language (same rules, same expectations).

**Acceptance Scenarios**:

1. **Given** Cobalt-Crush's coding standards framework, **When** a maintainer inspects it, **Then** it includes: language-agnostic principles (clean code, SOLID, DRY, YAGNI, separation of concerns), and language-specific convention packs that are compatible with The Divisor's convention packs.
2. **Given** a Go project, **When** Cobalt-Crush writes code, **Then** it adheres to the Go convention pack: gofmt, GoDoc on exported symbols, error wrapping with `%w`, import grouping (stdlib/external/internal), no global mutable state, table-driven tests.
3. **Given** a TypeScript project, **When** Cobalt-Crush writes code, **Then** it adheres to the TypeScript convention pack: ESLint/Prettier compliance, JSDoc on exported functions, proper async/await error handling, barrel exports, and type safety (no `any`).
4. **Given** a new language not yet covered by a convention pack, **When** Cobalt-Crush works in that language, **Then** it applies the language-agnostic principles and notes that a language-specific pack should be created.

---

### User Story 3 - Gaze Feedback Loop Integration (Priority: P2)

Cobalt-Crush integrates with Gaze's testing feedback in a tight, continuous loop. After writing code, Cobalt-Crush consumes Gaze's test results (unit tests, CRAP scores, contract coverage) and immediately addresses failures or quality gaps. Cobalt-Crush treats test feedback as an integral part of the development process, not a separate phase.

**Why this priority**: P2 because the Gaze feedback loop is what enables Cobalt-Crush to deliver code that is validated before it reaches review. It depends on the persona (US1) and coding standards (US2) being defined first.

**Independent Test**: Can be tested by having Cobalt-Crush implement a feature, running Gaze on the result, feeding the Gaze report back to Cobalt-Crush, and verifying Cobalt-Crush addresses the identified issues.

**Acceptance Scenarios**:

1. **Given** Cobalt-Crush has written code, **When** Gaze produces a quality report identifying low contract coverage on function `calculateTotal`, **Then** Cobalt-Crush adds tests that assert on the contractual side effects of `calculateTotal`.
2. **Given** a Gaze report showing a CRAP score > 30 on function `processInput`, **When** Cobalt-Crush reviews the report, **Then** it refactors `processInput` to reduce cyclomatic complexity or increases test coverage, targeting a CRAP score < 30.
3. **Given** Gaze requests a testability improvement (e.g., "function `sendEmail` has side effects that are hard to test"), **When** Cobalt-Crush receives the request, **Then** it refactors to inject the email dependency as an interface, making the function testable without real side effects.
4. **Given** Gaze reports all tests passing with good coverage, **When** Cobalt-Crush reviews the report, **Then** it proceeds to submit the PR for Divisor review.

---

### User Story 4 - Divisor Feedback Loop Integration (Priority: P2)

Cobalt-Crush integrates with The Divisor's review feedback. When The Divisor issues REQUEST CHANGES findings, Cobalt-Crush addresses each finding systematically, re-runs Gaze validation after changes, and re-submits for review. Cobalt-Crush maintains a record of past review feedback to avoid repeating the same mistakes.

**Why this priority**: P2 because the Divisor feedback loop completes the quality cycle. Code that passes Gaze but fails review is still not shippable. This loop is what ensures code reaches the "merge-ready" state.

**Independent Test**: Can be tested by simulating a Divisor review with specific findings, having Cobalt-Crush address each finding, and verifying the resulting code resolves the issues without introducing new ones.

**Acceptance Scenarios**:

1. **Given** The Divisor issues a REQUEST CHANGES finding "Function `processData` has O(n^2) complexity in the inner loop" (Adversary, severity: major), **When** Cobalt-Crush addresses it, **Then** it refactors the algorithm, adds a comment explaining the new complexity, and re-runs Gaze to verify no regressions.
2. **Given** The Divisor issues a finding "Missing error handling on line 42" (Adversary, severity: critical), **When** Cobalt-Crush addresses it, **Then** it adds proper error handling that follows the project's error wrapping convention.
3. **Given** past review reports show The Architect frequently requests "add GoDoc to exported function," **When** Cobalt-Crush writes new exported functions, **Then** it proactively includes GoDoc comments without waiting for review feedback (learned behavior).
4. **Given** Cobalt-Crush addresses all findings, **When** the code is re-submitted, **Then** only the personas that issued REQUEST CHANGES re-review (per The Divisor's iteration protocol).

---

### User Story 5 - Task Consumption and Speckit Integration (Priority: P3)

Cobalt-Crush consumes the speckit `tasks.md` file to drive implementation. It processes tasks phase by phase, respecting dependencies, marking tasks as complete, and producing checkpoint summaries. This is the operational integration between the spec-driven workflow and the actual code writing.

**Why this priority**: P3 because this is the automation of the development workflow. Manual task consumption works, but integration with speckit's `/implement` command enables autonomous, spec-driven development.

**Independent Test**: Can be tested by providing Cobalt-Crush with a tasks.md containing a simple phase, having it execute the tasks, and verifying the correct files are created/modified and tasks are marked complete.

**Acceptance Scenarios**:

1. **Given** a `tasks.md` with Phase 1 tasks, **When** Cobalt-Crush runs the implementation workflow, **Then** it processes tasks in dependency order, respects `[P]` parallelization markers, and maps `[US1]` tags to the corresponding user story in the spec.
2. **Given** a task `[T003] [US1] Create internal/scorer/scorer.go with Score(fn Function) int`, **When** Cobalt-Crush executes it, **Then** it creates the file, implements the function, writes tests, and marks `[T003]` as `[x]` in tasks.md.
3. **Given** a phase checkpoint, **When** all tasks in the phase are complete, **Then** Cobalt-Crush runs the project's test suite and reports pass/fail before proceeding to the next phase.
4. **Given** a task depends on another task that is not yet complete, **When** Cobalt-Crush encounters it, **Then** it skips the task and returns to it after the dependency is resolved.

---

### User Story 6 - Deployment Generator (Priority: P3)

Cobalt-Crush provides a `cobalt-crush init` command that deploys the Cobalt-Crush agent configuration into a target project. The generated agent file is pre-configured with the project's language, coding standards, and integrations with Gaze and The Divisor.

**Why this priority**: P3 because the deployment generator enables adoption in any project. It depends on the persona, coding standards, and feedback loops being defined first.

**Independent Test**: Can be tested by running `cobalt-crush init` in a Go project and verifying the generated agent file contains Go-specific coding conventions and references to Gaze and Divisor integration.

**Acceptance Scenarios**:

1. **Given** a Go project, **When** `cobalt-crush init` is run, **Then** it creates `.opencode/agents/cobalt-crush-dev.md` with Go coding conventions, Gaze feedback loop instructions, and Divisor review preparation guidelines.
2. **Given** a TypeScript project, **When** `cobalt-crush init --lang typescript` is run, **Then** the generated agent uses TypeScript conventions.
3. **Given** a project with Gaze and Divisor already deployed, **When** `cobalt-crush init` is run, **Then** the generated agent references the existing Gaze and Divisor agents by name for the feedback loops.
4. **Given** an existing Cobalt-Crush deployment, **When** `cobalt-crush init` is run without `--force`, **Then** existing files are skipped with a warning.

---

### Edge Cases

- What happens when Cobalt-Crush is deployed in a project without Gaze? Cobalt-Crush MUST still function. The Gaze feedback loop is optional — the agent notes that quality validation is not available and recommends installing Gaze.
- What happens when Cobalt-Crush is deployed in a project without The Divisor? Cobalt-Crush MUST still function. The Divisor feedback loop is optional — the agent notes that automated review is not available.
- What happens when Cobalt-Crush encounters a task in tasks.md that requires a technology it has no convention pack for? Cobalt-Crush MUST apply language-agnostic principles and flag the gap.
- What happens when Gaze and Divisor feedback contradict (e.g., Gaze says "add more tests," Divisor says "reduce test complexity")? Cobalt-Crush MUST address both by finding a solution that satisfies both constraints (e.g., more focused tests rather than more numerous tests). If irreconcilable, it MUST escalate to Muti-Mind for prioritization.
- What happens when `cobalt-crush init` cannot auto-detect the project language? It MUST prompt the user or accept `--lang` flag, falling back to language-agnostic defaults.
- What happens when a convention pack is updated after Cobalt-Crush has already written code? The next development session SHOULD note the convention pack version has changed and offer to review existing code against the new conventions.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Cobalt-Crush MUST provide an AI agent persona with a documented engineering philosophy: clean code, SOLID, TDD awareness, CI/CD focus, and spec-driven development.
- **FR-002**: The agent persona MUST be deployable as an OpenCode agent file (`cobalt-crush-dev.md`) installable via `cobalt-crush init`.
- **FR-003**: Cobalt-Crush MUST define a coding standards framework with: language-agnostic principles and language-specific convention packs.
- **FR-004**: Convention packs MUST be compatible with The Divisor's convention packs (Spec 005) — same rules, same terminology, same categories — ensuring developer-reviewer alignment.
- **FR-005**: Cobalt-Crush MUST integrate with Gaze's quality reporting: consume quality reports (JSON), identify issues, and produce corrective code changes.
- **FR-006**: Cobalt-Crush MUST integrate with The Divisor's review feedback: consume review-verdict artifacts, address findings, and re-submit for review.
- **FR-007**: Cobalt-Crush MUST maintain awareness of past review feedback patterns and proactively apply learned conventions to new code.
- **FR-008**: Cobalt-Crush MUST integrate with the speckit pipeline: consume tasks.md, process tasks in dependency order, respect parallelization markers, and mark tasks complete.
- **FR-009**: Cobalt-Crush MUST provide a CLI tool (`cobalt-crush init`) that generates project-specific OpenCode agent files pre-configured with the correct language and integrations.
- **FR-010**: `cobalt-crush init` MUST auto-detect the project language (from go.mod, package.json, pyproject.toml, etc.) or accept a `--lang` flag.
- **FR-011**: The agent MUST follow the Hero Interface Contract naming convention: `cobalt-crush-dev.md`.
- **FR-012**: Cobalt-Crush MUST produce code with appropriate test hooks (interface abstractions, dependency injection, exported test helpers) to facilitate Gaze's validation.
- **FR-013**: Cobalt-Crush MUST produce code that adheres to the project's constitution principles.
- **FR-014**: Cobalt-Crush MUST include documentation generation: inline code comments, GoDoc/JSDoc for exported symbols, and design decision records when architectural choices are made.
- **FR-015**: Cobalt-Crush MUST provide phase checkpoint validation: run the project's test suite after each completed phase and report results before proceeding.
- **FR-016**: Cobalt-Crush MUST conform to the Hero Interface Contract (Spec 002): standard repo structure, hero manifest, speckit integration, OpenCode agent/command standards.
- **FR-017**: Cobalt-Crush SHOULD share convention packs with The Divisor from a single source of truth, preventing drift between developer conventions and reviewer expectations.

### Key Entities

- **Developer Persona Configuration**: The AI agent's behavioral framework. Attributes: engineering_philosophy (clean code, SOLID, etc.), communication_style, decision_framework (how to resolve ambiguity), learning_model (how past feedback informs future behavior).
- **Coding Standards Framework**: The complete set of coding rules. Attributes: universal_principles[] (SOLID, DRY, YAGNI, etc.), language_packs{} (keyed by language), active_pack (the pack in use for the current project).
- **Convention Pack** (shared with The Divisor): Language-specific coding rules. Attributes: pack_id, language, coding_style{}, architectural_patterns{}, testing_conventions{}, documentation_requirements{}.
- **Feedback Loop Record**: A record of a Gaze or Divisor feedback cycle. Attributes: source_hero (gaze/divisor), report_ref (path), findings_addressed[], changes_made[], validation_result (pass/fail).
- **Task Execution Context**: State tracked while processing tasks.md. Attributes: current_phase, tasks_completed[], tasks_remaining[], dependencies_satisfied[], phase_checkpoint_result.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The Cobalt-Crush agent persona, when deployed, produces code from a spec that adheres to the specified coding conventions (verified by running The Divisor on the output).
- **SC-002**: Convention packs are identical between Cobalt-Crush and The Divisor (verified by checksum or content comparison).
- **SC-003**: Given a Gaze quality report with 3 identified issues, Cobalt-Crush produces corrective changes that resolve all 3 issues (verified by re-running Gaze).
- **SC-004**: Given a Divisor review with 3 findings, Cobalt-Crush addresses all 3 findings without introducing new ones (verified by re-running The Divisor).
- **SC-005**: Task consumption processes a 5-task phase in dependency order, marks all tasks complete, and runs the test suite checkpoint.
- **SC-006**: `cobalt-crush init` generates a functional agent file for at least two languages (Go and TypeScript).
- **SC-007**: The Cobalt-Crush agent functions without Gaze or The Divisor installed (standalone capability per Principle II).

## Dependencies

### Prerequisites

- **Spec 001** (Org Constitution): Cobalt-Crush must align with org principles.
- **Spec 002** (Hero Interface Contract): Cobalt-Crush must conform to the hero manifest, artifact envelope, and naming conventions.
- **Spec 005** (The Divisor Architecture): Convention packs must be shared or compatible.

### Downstream Dependents

- **Spec 008** (Swarm Orchestration): Cobalt-Crush is the implementation engine in the swarm workflow.
- **Spec 009** (Shared Data Model): Convention pack schema must be shared with The Divisor.

### Collaboration Partners

- **Gaze** (Tester): Cobalt-Crush consumes quality reports, Gaze validates Cobalt-Crush's output.
- **The Divisor** (Reviewer): Cobalt-Crush consumes review verdicts, shares convention packs.
- **Muti-Mind** (Product Owner): Cobalt-Crush consumes backlog items and specs, Muti-Mind accepts or rejects the output.
