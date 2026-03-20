---
spec_id: "005"
title: "The Divisor Architecture (PR Reviewer Council)"
phase: 1
status: complete
depends_on:
  - "[[specs/001-org-constitution/spec]]"
  - "[[specs/002-hero-interface-contract/spec]]"
---

# Feature Specification: The Divisor Architecture (PR Reviewer Council)

**Feature Branch**: `005-the-divisor-architecture`
**Created**: 2026-02-24
**Status**: Complete
**Input**: User description: "Design the architecture for The Divisor, the PR Reviewer Council hero. The Divisor is the Architectural Conscience and Code Integrity Guardian, realized by a council of three personas: The Guard (intent and cohesion), The Architect (structure and sustainability), and The Adversary (resilience and security). The Gaze repository contains a prototype deployment of The Divisor's review agents. The Divisor must be a standalone, reusable framework that produces project-specific deployments like the Gaze prototype." *(Note: Since clarified to five canonical personas — Guard, Architect, Adversary, SRE, Testing — with dynamic discovery. See Session 2026-03-19 clarifications.)*

## Clarifications

### Session 2026-02-24

- Q: The Gaze repo has reviewer agents (`reviewer-guard.md`, `reviewer-architect.md`, `reviewer-adversary.md`) and a `/review-council` command. Are these The Divisor project or a deployment of it? A: These are a prototype deployment of The Divisor. The Divisor project defines the framework; the Gaze agents are an instance configured for a Go static analysis tool.
- Q: How should The Divisor handle project-specific coding conventions? The Gaze deployment hardcodes Go-specific checks (gofmt, GoDoc, Go error wrapping). A: The Divisor framework must define convention packs — pluggable sets of language/framework-specific rules that are injected into the review personas. The Gaze Go convention pack is the first implementation.
- Q: Should The Divisor be a CLI tool, an OpenCode plugin, or agent configurations? A: Primarily agent configurations with a CLI tool for generating project-specific deployments (similar to `gaze init`). The CLI generates the agent files configured for the target project.

### Session 2026-03-19

- Q: How should The Divisor CLI be structured for distribution? A: The Divisor is distributed through the existing `unbound` binary, not a standalone repo or binary. `unbound init` deploys everything (speckit + openspec + Divisor agents). `unbound init --divisor` deploys only Divisor agents and commands (subset deployment). No separate `unbound-force/the-divisor` repo is needed.
- Q: What file pattern should the review-council command scan for to discover Divisor personas? A: Scan for `divisor-*.md` in `.opencode/agents/`. The existing `reviewer-*` prototype files will be renamed to `divisor-*` as part of implementation. This cleanly separates Divisor personas from other agents.
- Q: When should convention pack content be injected into persona agents? A: Review-time (dynamic). Agents reference a convention pack file path and load it at review time. This keeps agents thin and allows pack updates without re-scaffolding. Convention packs are deployed as separate files alongside the agents.
- Q: Should this spec define the `review-verdict` JSON schema or defer to Spec 009? A: Produce a Markdown report now. Defer JSON artifact envelope to Spec 009 (Shared Data Model). FR-017 becomes SHOULD until Spec 009 is complete.
- Q: How should the `reviewer-*` to `divisor-*` migration be handled? A: Deploy `divisor-*` files alongside existing `reviewer-*` files. Old files are left in place for manual cleanup (the scaffold engine does not delete files). Migration is documented in release notes.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Framework Core: Dynamic Review Protocol (Priority: P1)

The Divisor defines a formal review protocol with dynamic persona discovery that any project can deploy. The `/review-council` command discovers personas at runtime by scanning for `divisor-*.md` files in `.opencode/agents/`. Five canonical personas ship as defaults — Guard (intent), Architect (structure), Adversary (resilience), SRE (operations), Testing (test quality) — but users may add or remove personas freely. The protocol specifies how discovered personas each evaluate a code change, how their individual verdicts are combined into a council decision, and how iteration works when changes are requested.

**Why this priority**: P1 because the review protocol is the core intellectual property of The Divisor. Without a formal, project-agnostic protocol, each deployment would reinvent the review process.

**Independent Test**: Can be tested by presenting a sample code change to each persona template and verifying the protocol produces structured verdicts that can be combined into a council decision.

**Acceptance Scenarios**:

1. **Given** the review protocol specification, **When** a reviewer inspects it, **Then** it defines: dynamic persona discovery via `divisor-*.md` file scanning, five canonical persona roles (Guard, Architect, Adversary, SRE, Testing), their distinct focus areas, verdict format (APPROVE/REQUEST CHANGES/COMMENT), and the council decision rules.
2. **Given** a code change, **When** all discovered personas review it, **Then** each produces a structured verdict containing: persona name, verdict, findings[] (each with severity, category, file, line, description, recommendation), and a summary.
3. **Given** the individual verdicts from all discovered personas, **When** the council decision is computed, **Then** the change is APPROVED only if no discovered persona has issued REQUEST CHANGES. Any REQUEST CHANGES verdict blocks the merge. Absent personas do not affect the verdict.
4. **Given** a REQUEST CHANGES verdict, **When** the developer addresses the findings, **Then** the iteration protocol re-runs only the persona(s) that issued REQUEST CHANGES (up to a configurable maximum of iterations, default 3).
5. **Given** the maximum iteration count is reached, **When** unresolved findings remain, **Then** the council escalates to manual review with a summary of all unresolved findings.

---

### User Story 2 - Convention Packs: Language and Framework Adaptation (Priority: P1)

The Divisor supports convention packs — pluggable configurations that define language-specific and framework-specific coding conventions, architectural patterns, and security checks. Convention packs are deployed as separate files to `.opencode/divisor/packs/` and loaded dynamically at review time by each persona, allowing the same review protocol to evaluate Go code, TypeScript code, Python code, or any other stack.

**Why this priority**: P1 because the Gaze prototype is hardcoded for Go. Without convention packs, The Divisor cannot be deployed to non-Go projects, making it a single-project tool rather than a framework.

**Independent Test**: Can be tested by creating two convention packs (Go and TypeScript), deploying The Divisor with each, and verifying that the Architect persona checks for the correct language-specific conventions in each deployment.

**Acceptance Scenarios**:

1. **Given** a Go convention pack deployed at `.opencode/divisor/packs/go.md`, **When** The Architect reviews a Go PR, **Then** it dynamically loads the pack and checks for: gofmt compliance, GoDoc on exported symbols, error wrapping with `%w`, import grouping (stdlib/external/internal), and no global mutable state.
2. **Given** a convention pack for TypeScript, **When** The Architect reviews a TypeScript PR, **Then** it checks for: ESLint compliance, JSDoc on exported functions, proper error handling, import organization, and no `any` type usage.
3. **Given** a project with no convention pack configured, **When** The Divisor is deployed, **Then** the personas use a language-agnostic default pack that checks universal principles (SOLID, DRY, error handling, test coverage).
4. **Given** a convention pack file at `.opencode/divisor/packs/{language}.md`, **When** a maintainer inspects its structure, **Then** it is a structured document (Markdown or YAML) with sections for: coding_style, architectural_patterns, security_checks, testing_conventions, and documentation_requirements.
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

### User Story 4 - Deployment via `unbound init` (Priority: P2)

The Divisor is distributed through the existing `unbound` binary. `unbound init` deploys all scaffold files including Divisor agents, the `/review-council` command, and the appropriate convention pack file. `unbound init --divisor` deploys only Divisor-specific files as a subset. The scaffold engine auto-detects the project language and deploys the matching convention pack to `.opencode/divisor/packs/`.

**Why this priority**: P2 because the deployment mechanism depends on the protocol (US1) and convention packs (US2) being defined first.

**Independent Test**: Can be tested by running `unbound init --divisor` in a Go project and verifying it produces agent files that match (or improve upon) the existing Gaze prototype agents.

**Acceptance Scenarios**:

1. **Given** a Go project with a constitution, **When** `unbound init` is run, **Then** it creates `.opencode/agents/divisor-guard.md`, `.opencode/agents/divisor-architect.md`, `.opencode/agents/divisor-adversary.md`, `.opencode/agents/divisor-sre.md`, `.opencode/agents/divisor-testing.md`, `.opencode/command/review-council.md`, and `.opencode/divisor/packs/go.md` (among all other scaffold files).
2. **Given** a Go project, **When** `unbound init --divisor` is run, **Then** it creates only the Divisor agent and command files (not speckit templates, openspec schema, etc.).
3. **Given** a TypeScript project, **When** `unbound init --divisor --lang typescript` is run, **Then** it deploys the same persona agents plus `.opencode/divisor/packs/typescript.md` containing TypeScript-specific convention checks.
4. **Given** a project with an existing Divisor deployment, **When** `unbound init --divisor` is run without `--force`, **Then** existing Divisor files are skipped with a warning.
5. **Given** the generated agents, **When** a developer compares them to the Gaze prototype agents, **Then** the generated agents follow the same structural pattern but with convention-pack-driven content instead of hardcoded Go checks.

---

### User Story 5 - Review Report Artifact (Priority: P3)

The Divisor produces a standardized review report artifact (conforming to the inter-hero artifact envelope from Spec 002) that other heroes can consume. Mx F uses review reports for metrics. Muti-Mind uses them to understand implementation quality. Cobalt-Crush uses past reports to avoid repeating mistakes.

**Why this priority**: P3 because the review report artifact enables swarm integration. The core review functionality works without it, but cross-hero learning requires structured output.

**Independent Test**: Can be tested by running a review council session and verifying the structured Markdown report contains all required sections (persona verdicts, discovery summary, council decision). JSON artifact validation is deferred to Spec 009.

**Acceptance Scenarios**:

1. **Given** a completed review council session, **When** the report is generated, **Then** it conforms to the artifact envelope: `hero: "the-divisor"`, `artifact_type: "review-verdict"`, `payload` containing all discovered persona verdicts, discovery summary, and the council decision.
2. **Given** a review report, **When** Mx F parses it, **Then** it can extract: number of findings per severity, categories of findings, iteration count, and final verdict.
3. **Given** a history of review reports, **When** Mx F analyzes trends, **Then** it can identify recurring finding categories (e.g., "The Architect frequently requests error wrapping improvements" -> suggests training or convention enforcement).

---

### Edge Cases

- What happens when `unbound init --divisor` is run and the project language cannot be auto-detected? The CLI MUST prompt the user to specify the language or use `--lang` flag. If neither is provided, it falls back to the language-agnostic default pack.
- What happens when a persona's review takes too long (e.g., very large PR)? The review protocol SHOULD define a timeout per persona (configurable, default 5 minutes for agent execution) and report partial results if a timeout occurs.
- What happens when two personas produce contradictory findings? The council report MUST include both findings. The protocol does not resolve contradictions — the developer addresses each finding independently.
- What happens when a convention pack has no security checks? The Adversary MUST still perform universal security checks (hardcoded secrets, SQL injection patterns, etc.) regardless of the convention pack content.
- What happens when the target project has no tests and The Adversary checks for test coverage? The Adversary MUST flag the absence of tests as a finding but MUST NOT block the review solely for missing tests (that is Gaze's domain).
- What happens when the `/review-council` command is run on a draft PR? The protocol MUST still execute but the report SHOULD note it is a draft review and the final review will occur when the PR is marked ready.
- What happens when `unbound init --divisor` is run in a project that already has non-Divisor review agents? The Divisor MUST NOT overwrite or interfere with existing agents from other heroes. Its agents use the `divisor-` prefix to avoid collisions.
- What happens when a project has existing `reviewer-*.md` files from a previous `unbound init`? The new `divisor-*.md` files are deployed alongside the old `reviewer-*` files. The scaffold engine MUST NOT delete the old files. The `/review-council` command scans only for `divisor-*.md`, so the old `reviewer-*` files become inert. Migration documentation MUST note that users should manually remove old `reviewer-*` files after verifying `divisor-*` equivalents work correctly.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The Divisor MUST define a formal review protocol with dynamic persona discovery. The `/review-council` command MUST scan `.opencode/agents/` for files matching `divisor-*.md` at runtime. Five canonical personas ship as defaults: Guard (intent/cohesion), Architect (structure/sustainability), Adversary (resilience/security), SRE (operational readiness), and Testing (test quality/testability). Users MAY add or remove personas freely.
- **FR-002**: Each persona MUST produce a structured verdict: persona_name, verdict (APPROVE/REQUEST_CHANGES/COMMENT), findings[], summary.
- **FR-003**: Each finding MUST include: severity (critical/major/minor/info), category (string), file (path), line (number, optional), description, recommendation.
- **FR-004**: The council decision MUST be APPROVE only when no discovered persona has issued REQUEST_CHANGES. Absent personas (not present as `divisor-*.md` files) MUST NOT affect the verdict.
- **FR-005**: The iteration protocol MUST re-run only personas that issued REQUEST_CHANGES, up to a configurable maximum (default 3 iterations).
- **FR-006**: The Divisor MUST support convention packs — pluggable configurations defining language/framework-specific review criteria. Convention packs are loaded dynamically at review time, not baked into agent files at scaffold time.
- **FR-007**: Convention packs MUST be structured documents (Markdown or YAML) deployed to `.opencode/divisor/packs/` with sections: coding_style, architectural_patterns, security_checks, testing_conventions, documentation_requirements, and custom_rules[]. Each persona agent MUST reference the active pack file path and load it when invoked.
- **FR-008**: The Divisor MUST ship with at least two convention packs: Go (matching the Gaze prototype) and a language-agnostic default. `unbound init` deploys the appropriate pack to `.opencode/divisor/packs/` based on language detection.
- **FR-009**: The Divisor MUST be project-aware: personas MUST read the target project's constitution, active spec, and AGENTS.md when available.
- **FR-010**: The Guard MUST validate PR changes against the active spec's user stories and acceptance criteria ("intent drift detection").
- **FR-011**: The Architect MUST validate code structure against the project's constitution principles and the convention pack's architectural patterns.
- **FR-012**: The Adversary MUST check for: security vulnerabilities, performance anti-patterns, error handling gaps, and resilience issues, informed by the spec's edge cases and the convention pack's security checks.
- **FR-013**: The Divisor MUST be distributed through the `unbound` binary. `unbound init` deploys all scaffold files including Divisor agents. `unbound init --divisor` deploys only Divisor agents and commands as a subset.
- **FR-014**: `unbound init --divisor` MUST auto-detect the project language (from go.mod, package.json, pyproject.toml, etc.) or accept a `--lang` flag.
- **FR-015**: Generated agent files MUST follow the `divisor-{function}.md` naming convention. The five canonical files are: `divisor-guard.md`, `divisor-architect.md`, `divisor-adversary.md`, `divisor-sre.md`, `divisor-testing.md`.
- **FR-016**: The generated `/review-council` command MUST discover all `divisor-*.md` agents dynamically, orchestrate them in parallel, collect verdicts, compute the council decision, and handle iteration.
- **FR-017**: The Divisor MUST produce a structured Markdown review report. The Divisor SHOULD produce a `review-verdict` JSON artifact conforming to the inter-hero artifact envelope (Spec 002) once Spec 009 (Shared Data Model) defines the envelope schema. JSON artifact output is deferred to Spec 009.
- **FR-018**: The review report MUST include: all discovered persona verdicts, the council decision, a discovery summary (invoked and absent personas), iteration history, and metadata (PR URL, review timestamp, convention pack used).
- **FR-019**: The Divisor MUST conform to the Hero Interface Contract (Spec 002) as an embedded hero: OpenCode agent/command standards and artifact envelope compliance. The Divisor does not require a standalone repo; it is distributed as part of the `unbound` binary's scaffold assets. A formal hero manifest JSON file is deferred to Spec 009 (Shared Data Model), which will define the schema for embedded heroes that lack standalone repos.
- **FR-020**: The Adversary MUST perform universal security checks (hardcoded secrets, injection patterns) regardless of the convention pack.
- **FR-021**: The Guard MUST enforce the Zero-Waste Mandate: PRs should not introduce code that is not connected to the active spec or a documented backlog item.
- **FR-022**: The Architect MUST enforce the Neighborhood Rule: changes must not negatively impact adjacent modules not covered by the PR's scope.

### Key Entities

- **Review Protocol**: The formal process governing a review council session. Attributes: discovery_pattern (`divisor-*.md`), canonical_personas[] (5: guard, architect, adversary, sre, testing), voting_rules, iteration_max (int), timeout_per_persona (duration), escalation_policy.
- **Convention Pack**: Pluggable review criteria for a specific language/framework, deployed to `.opencode/divisor/packs/` and loaded dynamically at review time. Attributes: pack_id, language, framework (optional), coding_style{}, architectural_patterns{}, security_checks{}, testing_conventions{}, documentation_requirements{}, custom_rules[]. File path: `.opencode/divisor/packs/{language}.md` (e.g., `go.md`, `typescript.md`, `default.md`).
- **Persona Verdict**: One reviewer persona's evaluation of a PR. Attributes: persona (guard/architect/adversary/sre/testing or any dynamically discovered persona), verdict (APPROVE/REQUEST_CHANGES/COMMENT), findings[], summary, reviewed_at, iteration_number.
- **Review Finding**: A single issue identified during review. Attributes: id (F-NNN), severity (critical/major/minor/info), category, file, line (optional), description, recommendation, persona_source.
- **Council Decision**: The aggregate outcome of a review session. Attributes: decision (APPROVED/CHANGES_REQUESTED/ESCALATED), discovered_personas[], absent_personas[], persona_verdicts[], iteration_count, unresolved_findings[], reviewed_at, convention_pack_used, pr_url.
- **Deployment Configuration**: Settings for a project-specific Divisor deployment via `unbound init --divisor`. Attributes: target_dir, language, framework, convention_pack_id, project_constitution_path, project_spec_path, force_overwrite (bool), divisor_only (bool, true when `--divisor` flag used).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The review protocol is formally documented and covers: dynamic persona discovery, five canonical personas, verdict format, council decision rules, iteration protocol, and escalation.
- **SC-002**: The Go convention pack produces review behavior equivalent to the existing Gaze prototype agents (verified by comparing review findings on the same sample PR).
- **SC-003**: The language-agnostic default pack produces meaningful findings on a project in any language (verified by reviewing a Python or TypeScript PR).
- **SC-004**: `unbound init --divisor` in a Go project produces agent files structurally equivalent to the Gaze prototype (same sections, same behavioral constraints, convention-pack-driven content).
- **SC-005**: `unbound init --divisor --lang typescript` produces agent files with TypeScript-specific convention checks.
- **SC-006**: The review report Markdown is structured and machine-parseable. JSON artifact validation against the `review-verdict` schema is deferred to Spec 009.
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

- **Gaze Prototype**: The agents in the Gaze repo at `.opencode/agents/` (`reviewer-guard.md`, `reviewer-architect.md`, `reviewer-adversary.md`) and the `/review-council` command (`review-council.md`) serve as the prototype deployment (see [github.com/unbound-force/gaze](https://github.com/unbound-force/gaze)). The Divisor framework must produce equivalent (or improved) output for Go projects.

```
The Divisor Framework (embedded in unbound binary)
┌────────────────────────────────────────────────┐
│ Review Protocol (formal spec)                  │
│ Convention Packs (Go, TS, Python, default)     │
│ Scaffold Assets (embed.FS in unbound binary)   │
│ Artifact Producer (review-verdict envelope)    │
└───────────┬──────────────────────┬─────────────┘
            │ unbound init         │ unbound init
            │                      │ --divisor --lang ts
            ▼                      ▼
┌───────────────────┐  ┌────────────────────────┐
│ Go Project        │  │ TS Project             │
│ (Go convention)   │  │ (TS convention)        │
│ divisor-guard.md  │  │ divisor-guard.md       │
│ divisor-arch.md   │  │ divisor-arch.md        │
│ divisor-adv.md    │  │ divisor-adv.md         │
│ + SRE, testing    │  │ + SRE, testing         │
│ review-council.md │  │ review-council.md      │
└───────────────────┘  └────────────────────────┘
```
