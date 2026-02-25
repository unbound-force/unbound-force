# Feature Specification: The Divisor Architecture (PR Reviewer Council)

**Feature Branch**: `005-the-divisor-architecture`
**Created**: 2026-02-24
**Status**: Draft
**Input**: User description: "Design the architecture for The Divisor, the PR Reviewer Council hero. The Divisor is the Architectural Conscience and Code Integrity Guardian, realized by a council of three personas: The Guard (intent and cohesion), The Architect (structure and sustainability), and The Adversary (resilience and security). The Gaze repository contains a prototype deployment of The Divisor's review agents. The Divisor must be a standalone, reusable framework that produces project-specific deployments like the Gaze prototype."

## Clarifications

### Session 2026-02-24

- Q: The Gaze repo has reviewer agents (`reviewer-guard.md`, `reviewer-architect.md`, `reviewer-adversary.md`) and a `/review-council` command. Are these The Divisor project or a deployment of it? A: These are a prototype deployment of The Divisor. The Divisor project defines the framework; the Gaze agents are an instance configured for a Go static analysis tool.
- Q: How should The Divisor handle project-specific coding conventions? The Gaze deployment hardcodes Go-specific checks (gofmt, GoDoc, Go error wrapping). A: The Divisor framework must define convention packs — pluggable sets of language/framework-specific rules that are injected into the review personas. The Gaze Go convention pack is the first implementation.
- Q: Should The Divisor be a CLI tool, an OpenCode plugin, or agent configurations? A: Primarily agent configurations with a CLI tool for generating project-specific deployments (similar to `gaze init`). The CLI generates the agent files configured for the target project.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Framework Core: Three-Persona Review Protocol (Priority: P1)

The Divisor defines a formal three-persona review protocol that any project can deploy. The protocol specifies how The Guard (intent), The Architect (structure), and The Adversary (resilience) each evaluate a code change, how their individual verdicts are combined into a council decision, and how iteration works when changes are requested.

**Why this priority**: P1 because the review protocol is the core intellectual property of The Divisor. Without a formal, project-agnostic protocol, each deployment would reinvent the review process.

**Independent Test**: Can be tested by presenting a sample code change to each persona template and verifying the protocol produces structured verdicts that can be combined into a council decision.

**Acceptance Scenarios**:

1. **Given** the review protocol specification, **When** a reviewer inspects it, **Then** it defines: three persona roles (Guard, Architect, Adversary), their distinct focus areas, verdict format (APPROVE/REQUEST CHANGES/COMMENT), and the council decision rules.
2. **Given** a code change, **When** all three personas review it, **Then** each produces a structured verdict containing: persona name, verdict, findings[] (each with severity, category, file, line, description, recommendation), and a summary.
3. **Given** three individual verdicts, **When** the council decision is computed, **Then** the change is APPROVED only if no persona has issued REQUEST CHANGES. Any REQUEST CHANGES verdict blocks the merge.
4. **Given** a REQUEST CHANGES verdict, **When** the developer addresses the findings, **Then** the iteration protocol re-runs only the persona(s) that issued REQUEST CHANGES (up to a configurable maximum of iterations, default 3).
5. **Given** the maximum iteration count is reached, **When** unresolved findings remain, **Then** the council escalates to manual review with a summary of all unresolved findings.

---

### User Story 2 - Convention Packs: Language and Framework Adaptation (Priority: P1)

The Divisor supports convention packs — pluggable configurations that define language-specific and framework-specific coding conventions, architectural patterns, and security checks. A convention pack is injected into the review personas at deployment time, allowing the same review protocol to evaluate Go code, TypeScript code, Python code, or any other stack.

**Why this priority**: P1 because the Gaze prototype is hardcoded for Go. Without convention packs, The Divisor cannot be deployed to non-Go projects, making it a single-project tool rather than a framework.

**Independent Test**: Can be tested by creating two convention packs (Go and TypeScript), deploying The Divisor with each, and verifying that the Architect persona checks for the correct language-specific conventions in each deployment.

**Acceptance Scenarios**:

1. **Given** a convention pack for Go, **When** The Architect reviews a Go PR, **Then** it checks for: gofmt compliance, GoDoc on exported symbols, error wrapping with `%w`, import grouping (stdlib/external/internal), and no global mutable state.
2. **Given** a convention pack for TypeScript, **When** The Architect reviews a TypeScript PR, **Then** it checks for: ESLint compliance, JSDoc on exported functions, proper error handling, import organization, and no `any` type usage.
3. **Given** a project with no convention pack configured, **When** The Divisor is deployed, **Then** the personas use a language-agnostic default pack that checks universal principles (SOLID, DRY, error handling, test coverage).
4. **Given** a convention pack, **When** a maintainer inspects its structure, **Then** it is a structured document (YAML or Markdown) with sections for: coding_style, architectural_patterns, security_checks, testing_conventions, and documentation_requirements.
5. **Given** a convention pack, **When** a project needs a custom rule not in the pack, **Then** the pack supports a `custom_rules[]` extension section where project-specific checks can be added without modifying the pack itself.

---

### User Story 3 - Project-Aware Review Context (Priority: P2)

The Divisor personas are project-aware: they read the target project's constitution, active spec, and AGENTS.md to inform their review criteria. The Guard validates intent alignment against the spec. The Architect validates structural compliance against the constitution and AGENTS.md conventions. The Adversary validates security and resilience against the spec's edge cases and the constitution's constraints.

**Why this priority**: P2 because project awareness transforms The Divisor from a generic linter into a context-sensitive reviewer that understands what the code is supposed to do, not just how it's structured.

**Independent Test**: Can be tested by deploying The Divisor in a project with a spec and constitution, submitting a PR that violates a spec acceptance criterion, and verifying The Guard detects the intent drift.

**Acceptance Scenarios**:

1. **Given** a project with an active spec containing acceptance criteria, **When** The Guard reviews a PR, **Then** it verifies the PR's changes are aligned with the spec's user stories and flags changes that appear unrelated to the active spec ("intent drift").
2. **Given** a project with a ratified constitution, **When** The Architect reviews a PR, **Then** it verifies the code adheres to the principles defined in the constitution (e.g., if the constitution says "Library-First," it checks that new code is structured as a library).
3. **Given** a project with edge cases defined in the spec, **When** The Adversary reviews a PR, **Then** it verifies the implementation handles the documented edge cases and flags any that appear unaddressed.
4. **Given** a project with no constitution or spec, **When** The Divisor reviews a PR, **Then** it falls back to convention-pack-only review and notes that project context was unavailable.

---

### User Story 4 - Deployment Generator CLI (Priority: P2)

The Divisor provides a CLI tool (`divisor init`) that generates project-specific OpenCode agent files and the `/review-council` command file. The generator reads the target project's context (language, framework, constitution) and produces agent files pre-configured with the appropriate convention pack and project references.

**Why this priority**: P2 because the deployment generator is the distribution mechanism. It depends on the protocol (US1) and convention packs (US2) being defined first.

**Independent Test**: Can be tested by running `divisor init` in a Go project and verifying it produces agent files that match (or improve upon) the existing Gaze prototype agents.

**Acceptance Scenarios**:

1. **Given** a Go project with a constitution, **When** `divisor init` is run, **Then** it creates `.opencode/agents/divisor-guard.md`, `.opencode/agents/divisor-architect.md`, `.opencode/agents/divisor-adversary.md`, and `.opencode/command/review-council.md` with Go convention pack content injected.
2. **Given** a TypeScript project, **When** `divisor init --lang typescript` is run, **Then** the generated agents contain TypeScript-specific convention checks instead of Go checks.
3. **Given** a project with an existing Divisor deployment, **When** `divisor init` is run without `--force`, **Then** existing files are skipped with a warning.
4. **Given** `divisor init --force`, **When** run in a project with existing deployment files, **Then** all files are overwritten and a summary is printed.
5. **Given** the generated agents, **When** a developer compares them to the Gaze prototype agents, **Then** the generated agents follow the same structural pattern but with convention-pack-driven content instead of hardcoded Go checks.

---

### User Story 5 - Review Report Artifact (Priority: P3)

The Divisor produces a standardized review report artifact (conforming to the inter-hero artifact envelope from Spec 002) that other heroes can consume. Mx F uses review reports for metrics. Muti-Mind uses them to understand implementation quality. Cobalt-Crush uses past reports to avoid repeating mistakes.

**Why this priority**: P3 because the review report artifact enables swarm integration. The core review functionality works without it, but cross-hero learning requires structured output.

**Independent Test**: Can be tested by running a review council session and verifying the output JSON validates against the `review-verdict` artifact type schema.

**Acceptance Scenarios**:

1. **Given** a completed review council session, **When** the report is generated, **Then** it conforms to the artifact envelope: `hero: "the-divisor"`, `artifact_type: "review-verdict"`, `payload` containing the three persona verdicts and the council decision.
2. **Given** a review report, **When** Mx F parses it, **Then** it can extract: number of findings per severity, categories of findings, iteration count, and final verdict.
3. **Given** a history of review reports, **When** Mx F analyzes trends, **Then** it can identify recurring finding categories (e.g., "The Architect frequently requests error wrapping improvements" -> suggests training or convention enforcement).

---

### Edge Cases

- What happens when `divisor init` is run and the project language cannot be auto-detected? The CLI MUST prompt the user to specify the language or use `--lang` flag. If neither is provided, it falls back to the language-agnostic default pack.
- What happens when a persona's review takes too long (e.g., very large PR)? The review protocol SHOULD define a timeout per persona (configurable, default 5 minutes for agent execution) and report partial results if a timeout occurs.
- What happens when two personas produce contradictory findings? The council report MUST include both findings. The protocol does not resolve contradictions — the developer addresses each finding independently.
- What happens when a convention pack has no security checks? The Adversary MUST still perform universal security checks (hardcoded secrets, SQL injection patterns, etc.) regardless of the convention pack content.
- What happens when the target project has no tests and The Adversary checks for test coverage? The Adversary MUST flag the absence of tests as a finding but MUST NOT block the review solely for missing tests (that is Gaze's domain).
- What happens when the `/review-council` command is run on a draft PR? The protocol MUST still execute but the report SHOULD note it is a draft review and the final review will occur when the PR is marked ready.
- What happens when `divisor init` is run in a project that already has non-Divisor review agents? The Divisor MUST NOT overwrite or interfere with existing agents from other heroes. Its agents use the `divisor-` prefix to avoid collisions.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The Divisor MUST define a formal three-persona review protocol with: Guard (intent/cohesion), Architect (structure/sustainability), Adversary (resilience/security).
- **FR-002**: Each persona MUST produce a structured verdict: persona_name, verdict (APPROVE/REQUEST_CHANGES/COMMENT), findings[], summary.
- **FR-003**: Each finding MUST include: severity (critical/major/minor/info), category (string), file (path), line (number, optional), description, recommendation.
- **FR-004**: The council decision MUST be APPROVE only when no persona has issued REQUEST_CHANGES.
- **FR-005**: The iteration protocol MUST re-run only personas that issued REQUEST_CHANGES, up to a configurable maximum (default 3 iterations).
- **FR-006**: The Divisor MUST support convention packs — pluggable configurations defining language/framework-specific review criteria.
- **FR-007**: Convention packs MUST be structured documents with sections: coding_style, architectural_patterns, security_checks, testing_conventions, documentation_requirements, and custom_rules[].
- **FR-008**: The Divisor MUST ship with at least two convention packs: Go (matching the Gaze prototype) and a language-agnostic default.
- **FR-009**: The Divisor MUST be project-aware: personas MUST read the target project's constitution, active spec, and AGENTS.md when available.
- **FR-010**: The Guard MUST validate PR changes against the active spec's user stories and acceptance criteria ("intent drift detection").
- **FR-011**: The Architect MUST validate code structure against the project's constitution principles and the convention pack's architectural patterns.
- **FR-012**: The Adversary MUST check for: security vulnerabilities, performance anti-patterns, error handling gaps, and resilience issues, informed by the spec's edge cases and the convention pack's security checks.
- **FR-013**: The Divisor MUST provide a CLI tool (`divisor init`) that generates project-specific OpenCode agent and command files.
- **FR-014**: `divisor init` MUST auto-detect the project language (from go.mod, package.json, pyproject.toml, etc.) or accept a `--lang` flag.
- **FR-015**: Generated agent files MUST follow the Hero Interface Contract naming convention: `divisor-guard.md`, `divisor-architect.md`, `divisor-adversary.md`.
- **FR-016**: The generated `/review-council` command MUST orchestrate the three personas in parallel, collect verdicts, compute the council decision, and handle iteration.
- **FR-017**: The Divisor MUST produce a `review-verdict` artifact type conforming to the inter-hero artifact envelope (Spec 002).
- **FR-018**: The review report MUST include: all three persona verdicts, the council decision, iteration history, and metadata (PR URL, review timestamp, convention pack used).
- **FR-019**: The Divisor MUST conform to the Hero Interface Contract (Spec 002): standard repo structure, hero manifest, speckit integration, OpenCode agent/command standards.
- **FR-020**: The Adversary MUST perform universal security checks (hardcoded secrets, injection patterns) regardless of the convention pack.
- **FR-021**: The Guard MUST enforce the Zero-Waste Mandate: PRs should not introduce code that is not connected to the active spec or a documented backlog item.
- **FR-022**: The Architect MUST enforce the Neighborhood Rule: changes must not negatively impact adjacent modules not covered by the PR's scope.

### Key Entities

- **Review Protocol**: The formal process governing a review council session. Attributes: personas[] (3), voting_rules, iteration_max (int), timeout_per_persona (duration), escalation_policy.
- **Convention Pack**: Pluggable review criteria for a specific language/framework. Attributes: pack_id, language, framework (optional), coding_style{}, architectural_patterns{}, security_checks{}, testing_conventions{}, documentation_requirements{}, custom_rules[].
- **Persona Verdict**: One reviewer persona's evaluation of a PR. Attributes: persona (guard/architect/adversary), verdict (APPROVE/REQUEST_CHANGES/COMMENT), findings[], summary, reviewed_at, iteration_number.
- **Review Finding**: A single issue identified during review. Attributes: id (F-NNN), severity (critical/major/minor/info), category, file, line (optional), description, recommendation, persona_source.
- **Council Decision**: The aggregate outcome of a review session. Attributes: decision (APPROVED/CHANGES_REQUESTED/ESCALATED), persona_verdicts[], iteration_count, unresolved_findings[], reviewed_at, convention_pack_used, pr_url.
- **Deployment Configuration**: Settings for generating a project-specific Divisor deployment. Attributes: target_dir, language, framework, convention_pack_id, project_constitution_path, project_spec_path, force_overwrite (bool).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The review protocol is formally documented and covers: three personas, verdict format, council decision rules, iteration protocol, and escalation.
- **SC-002**: The Go convention pack produces review behavior equivalent to the existing Gaze prototype agents (verified by comparing review findings on the same sample PR).
- **SC-003**: The language-agnostic default pack produces meaningful findings on a project in any language (verified by reviewing a Python or TypeScript PR).
- **SC-004**: `divisor init` in a Go project produces agent files structurally equivalent to the Gaze prototype (same sections, same behavioral constraints, convention-pack-driven content).
- **SC-005**: `divisor init --lang typescript` produces agent files with TypeScript-specific convention checks.
- **SC-006**: The review report JSON validates against the `review-verdict` artifact type schema.
- **SC-007**: The iteration protocol correctly re-runs only the requesting personas and terminates after the maximum iteration count.
- **SC-008**: Project-aware review (with constitution and spec available) produces more targeted findings than convention-pack-only review (measured by relevance of findings on a sample PR).

## Dependencies

### Prerequisites

- **Spec 001** (Org Constitution): The Divisor must align with org principles.
- **Spec 002** (Hero Interface Contract): The Divisor must conform to the hero manifest, artifact envelope, and naming conventions.

### Downstream Dependents

- **Spec 006** (Cobalt-Crush Architecture): Cobalt-Crush consumes Divisor review feedback.
- **Spec 007** (Mx F Architecture): Mx F consumes review-verdict artifacts for metrics.
- **Spec 008** (Swarm Orchestration): The Divisor is a gate in the "feature to deployment" workflow.
- **Spec 009** (Shared Data Model): Defines the `review-verdict` JSON schema.

### Reference Implementation

- **Gaze Prototype**: The agents in `/Users/jflowers/Projects/github/unbound-force/gaze/.opencode/agents/` (`reviewer-guard.md`, `reviewer-architect.md`, `reviewer-adversary.md`) and the `/review-council` command (`review-council.md`) serve as the prototype deployment. The Divisor framework must produce equivalent (or improved) output for Go projects.

```
The Divisor Framework
┌────────────────────────────────────────────────┐
│ Review Protocol (formal spec)                  │
│ Convention Packs (Go, TS, Python, default)     │
│ Deployment Generator CLI (divisor init)        │
│ Artifact Producer (review-verdict envelope)    │
└───────────┬──────────────────────┬─────────────┘
            │ generates            │ generates
            ▼                      ▼
┌───────────────────┐  ┌────────────────────────┐
│ Gaze Deployment   │  │ Project X Deployment   │
│ (Go convention)   │  │ (TS convention)        │
│ divisor-guard.md  │  │ divisor-guard.md       │
│ divisor-arch.md   │  │ divisor-arch.md        │
│ divisor-adv.md    │  │ divisor-adv.md         │
│ review-council.md │  │ review-council.md      │
└───────────────────┘  └────────────────────────┘
```
