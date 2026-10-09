---
spec_id: "005"
title: "The Divisor Architecture (PR Reviewer Council)"
phase: 1
status: complete
depends_on:
  - "[[specs/001-org-constitution/spec]]"
  - "[[specs/002-hero-interface-contract/spec]]"
---

# Feature Specification: The Divisor Architecture

**Feature Branch**: `005-the-divisor-architecture`
**Created**: 2026-02-24
**Status**: Complete
**Input**: User description: "Design the architecture for The
Divisor, the PR Reviewer Council hero. The Divisor is the
Architectural Conscience and Code Integrity Guardian, realized by
a council of three personas: The Guard, The Architect, and The
Adversary. The Gaze repository contains a prototype deployment.
The Divisor must be a standalone, reusable framework that produces
project-specific deployments like the Gaze prototype."

The original three-persona intent remains historical context. The
March 2026 design expanded the council to five review personas. The
current contract recognizes nine known personas: six review-capable
personas and three content-only personas.

## Clarifications

### Session 2026-02-24

- Q: Are the Gaze reviewer agents The Divisor project? A: They are
  a prototype deployment. The Divisor defines the reusable
  framework; Gaze is a project-specific instance.
- Q: How does The Divisor handle project-specific conventions?
  A: It uses pluggable language and framework convention packs.
- Q: Is The Divisor a CLI, plugin, or agent configuration? A: It is
  primarily agent configuration with CLI deployment support.

### Session 2026-03-19

- Q: How is The Divisor distributed? A: The existing `unbound`
  binary distributes it. `unbound init` deploys all assets, while
  `unbound init --divisor` deploys the Divisor subset.
- Q: Which pattern discovers Divisor personas? A: The command scans
  `.opencode/agents/divisor-*.md`.
- Q: When are convention packs loaded? A: Personas load them at
  review time so pack updates do not require re-scaffolding.
- Q: Where is the JSON decision contract defined? A: Spec 009 owns
  the shared schema. This spec owns its production semantics.
- Q: How is the old `reviewer-*` naming migrated? A: New
  `divisor-*` files deploy alongside old files. Users remove old
  files after verification because the scaffold does not delete.

### Session 2026-10-02

- Q: Which known personas can review? A: Adversary, Architect,
  Curator, Guard, SRE, and Testing are review-capable. Envoy,
  Herald, and Scribe are content-only.
- Q: What determines eligibility? A: The closed
  `.uf/reviewer-capabilities.yaml` manifest is authoritative.
- Q: Can a persona have multiple model runs? A: Yes. A validated
  dynamic plan MAY include one or more runs per included persona.
- Q: Which artifact drives downstream decisions? A:
  `review-verdict` version 2 is canonical. `review-dispatch` adds
  execution and provenance data and does not replace it.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Dynamic Review Protocol (Priority: P1)

The Divisor defines a project-agnostic review protocol. The council
discovers `divisor-*.md` files, classifies them through the closed
reviewer-capabilities manifest, creates a deterministic review plan,
and executes every included run. The known persona set contains six
review-capable personas and three content-only personas.

Review-capable personas are Adversary, Architect, Curator, Guard,
SRE, and Testing. Content-only personas are Envoy, Herald, and
Scribe. Content-only personas remain discoverable but never dispatch
as reviewers.

**Why this priority**: The protocol is the core intellectual
property of The Divisor. Without a formal protocol, each deployment
would reinvent discovery, eligibility, dispatch, and decisions.

**Independent Test**: Present a sample change to a validated plan
and verify that every included run produces structured output that
can be consolidated into one council decision.

**Acceptance Scenarios**:

1. **Given** the protocol specification, **When** a reviewer
   inspects it, **Then** it defines dynamic discovery, all nine known
   personas, their capabilities and scopes, verdict format, and
   council decision rules.
2. **Given** a code change, **When** included plan runs review it,
   **Then** each successful run produces a structured verdict with
   persona, model provenance, findings, and summary.
3. **Given** successful included runs, **When** the council decision
   is computed, **Then** any blocking verdict blocks approval and
   absent or skipped personas do not vote.
4. **Given** confirmed fixes, **When** the next iteration begins,
   **Then** the council recomputes and validates the plan and reruns
   every included plan run, including runs that did not previously
   block.
5. **Given** the three-iteration limit is reached, **When** findings
   remain, **Then** the council escalates to human review with all
   unresolved findings.
6. **Given** a content-only persona is discovered, **When** planning
   occurs, **Then** it is reported but no review run is created.
7. **Given** a change has no documentation or user-facing scope,
   **When** planning occurs, **Then** Curator MAY be pruned with a
   deterministic reason.

---

### User Story 2 - Convention Packs (Priority: P1)

The Divisor supports pluggable convention packs for language and
framework conventions, architectural patterns, security checks,
testing practices, and documentation requirements. Packs deploy as
separate files and load at review time.

**Why this priority**: The Gaze prototype is Go-specific. Packs make
the framework reusable across languages and projects.

**Independent Test**: Deploy Go and TypeScript packs, then verify
that Architect evaluates each project against its active pack.

**Acceptance Scenarios**:

1. **Given** a Go convention pack, **When** Architect reviews Go,
   **Then** it checks formatting, GoDoc, wrapped errors, imports, and
   mutable global state.
2. **Given** a TypeScript pack, **When** Architect reviews
   TypeScript, **Then** it checks ESLint, JSDoc, errors, imports, and
   prohibited `any` usage.
3. **Given** no language pack, **When** review begins, **Then** the
   personas use the language-agnostic default pack.
4. **Given** a convention pack, **When** a maintainer inspects it,
   **Then** it contains coding, architecture, security, testing, and
   documentation sections.
5. **Given** a project-specific rule, **When** the pack is extended,
   **Then** the custom rule does not require changing the canonical
   pack.

---

### User Story 3 - Project-Aware Context (Priority: P2)

Review personas read the target constitution, active spec, and
`AGENTS.md` when available. Guard validates intent. Architect
validates structure. Adversary validates security and resilience.

**Why this priority**: Project context lets the council assess what
the code should do, not only how it is written.

**Independent Test**: Review a change that violates an acceptance
criterion and verify that Guard reports the intent drift.

**Acceptance Scenarios**:

1. **Given** an active spec, **When** Guard reviews a change,
   **Then** it checks alignment with user stories and criteria.
2. **Given** a ratified constitution, **When** Architect reviews,
   **Then** it checks the implementation against its principles.
3. **Given** documented edge cases, **When** Adversary reviews,
   **Then** it reports unhandled cases.
4. **Given** no constitution or spec, **When** review begins,
   **Then** the council falls back to convention packs and reports
   the missing context.

---

### User Story 4 - Deployment via `unbound init` (Priority: P2)

The existing `unbound` binary deploys The Divisor. Full setup
deploys all scaffold assets. `unbound init --divisor` deploys only
Divisor assets. Language detection selects the convention pack.

**Why this priority**: Deployment depends on the protocol and packs
being defined first.

**Independent Test**: Run `unbound init --divisor` in a Go project
and verify that it deploys the expected agents, command, manifest,
and convention pack.

**Acceptance Scenarios**:

1. **Given** a Go project, **When** full initialization runs,
   **Then** all known Divisor agents, the review command, the closed
   reviewer manifest, and the Go pack are deployed.
2. **Given** a Go project, **When** Divisor-only initialization runs,
   **Then** only the Divisor subset is deployed.
3. **Given** a TypeScript project, **When** initialization specifies
   TypeScript, **Then** the TypeScript pack is deployed.
4. **Given** an existing deployment, **When** initialization runs
   without `--force`, **Then** user-owned files are skipped with a
   warning.
5. **Given** generated agents, **When** compared with the Gaze
   prototype, **Then** they preserve the structural intent while
   loading project conventions dynamically.

---

### User Story 5 - Review Decision Artifacts (Priority: P3)

The Divisor produces a canonical `review-verdict` version 2
artifact under the Hero Interface Contract. It also produces an
additive `review-dispatch` artifact containing the validated plan,
model runs, execution outcomes, and provenance.

Mx F uses canonical verdicts for metrics. Muti-Mind uses them for
acceptance decisions. Cobalt-Crush uses them to address findings and
avoid repeated mistakes. Version 1 verdicts remain readable as
historical data.

**Why this priority**: Structured decisions enable cross-hero
integration. Additive dispatch data makes execution auditable
without changing which artifact controls downstream decisions.

**Independent Test**: Complete a council review and validate the
Markdown report, `review-verdict` version 2 artifact, and additive
`review-dispatch` artifact against their schemas.

**Acceptance Scenarios**:

1. **Given** a completed council session, **When** artifacts are
   emitted, **Then** `review-verdict` version 2 contains the canonical
   downstream decision and `review-dispatch` contains additive
   execution and provenance data.
2. **Given** a version 1 verdict, **When** a migrated consumer reads
   it, **Then** historical reading remains supported.
3. **Given** no successful review run, **When** canonical output is
   produced, **Then** version 2 records `INCONCLUSIVE` or
   `UNAVAILABLE` and downstream automation remains blocked.
4. **Given** a review history, **When** Mx F analyzes it, **Then** it
   can identify recurring categories, iteration counts, and final
   decisions without treating dispatch data as canonical.

---

### Edge Cases

- If project language detection fails, the CLI MUST prompt for a
  language or use the default pack.
- Each run SHOULD have a configurable timeout. Partial successful
  results MUST remain available when another run times out.
- Contradictory findings MUST remain visible with their provenance.
- Adversary MUST perform universal security checks even when the
  active convention pack has no security section.
- Missing tests alone MUST NOT block solely through Adversary; test
  quality belongs to Testing and Gaze.
- Draft pull requests MUST still run, and the report SHOULD note the
  draft state.
- Existing non-Divisor agents MUST NOT be overwritten or invoked as
  Divisor personas.
- Legacy `reviewer-*` files MUST remain inert because discovery uses
  only `divisor-*`.
- Curator child runs MUST NOT create issues or perform curation
  actions. Only the parent MAY perform one deduplicated action after
  consolidation and an explicit human gate.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The Divisor MUST discover `divisor-*.md` files at
  runtime. The nine known personas MUST be Adversary, Architect,
  Curator, Envoy, Guard, Herald, Scribe, SRE, and Testing. Adversary,
  Architect, Curator, Guard, SRE, and Testing MUST be review-capable.
  Envoy, Herald, and Scribe MUST be content-only.
- **FR-002**: Each successful run MUST produce a structured verdict
  containing persona, verdict, findings, and summary.
- **FR-003**: Each finding MUST include severity, category, file,
  optional line, description, and recommendation.
- **FR-004**: The council MUST approve only when no successful
  included run requests changes. Absent, skipped, content-only, and
  unsuccessful runs MUST NOT vote. A no-success result MUST NOT
  become approval.
- **FR-005**: After the human confirms fixes, each iteration MUST
  recompute and validate the plan and rerun every included plan run.
  It MUST NOT rerun only previously blocking personas. The existing
  human fix gate and maximum of three iterations MUST remain.
- **FR-006**: The Divisor MUST support pluggable convention packs
  loaded dynamically at review time.
- **FR-007**: Convention packs MUST be structured documents in
  `.opencode/unbound/packs/`. They MUST cover coding, architecture,
  security, testing, documentation, and custom rules.
- **FR-008**: The Divisor MUST ship at least Go and default packs.
  Initialization MUST deploy the pack selected by language
  detection.
- **FR-009**: Review personas MUST read the target constitution,
  active spec, and `AGENTS.md` when available.
- **FR-010**: Guard MUST detect intent drift against active stories
  and acceptance criteria.
- **FR-011**: Architect MUST validate structure against the
  constitution and active convention packs.
- **FR-012**: Adversary MUST assess security, performance, error
  handling, and resilience against project context.
- **FR-013**: The existing `unbound` binary MUST distribute The
  Divisor. Full initialization and `--divisor` subset deployment
  MUST remain supported.
- **FR-014**: Divisor initialization MUST detect project language or
  accept a language flag.
- **FR-015**: Generated known persona files MUST use the
  `divisor-{function}.md` naming convention. The generated set MUST
  contain adversary, architect, curator, envoy, guard, herald,
  scribe, SRE, and testing files.
- **FR-016**: The review command MUST discover agents, obtain and
  validate a dynamic plan, execute every included run, consolidate
  findings, compute the decision, and handle iteration. A plan MAY
  contain one or more model runs for each included persona.
- **FR-017**: The Divisor MUST produce a structured Markdown report
  and a canonical `review-verdict` version 2 artifact. Producers MUST
  preserve version 1 historical reads. Version 2 MUST support
  `INCONCLUSIVE` and `UNAVAILABLE` no-success decisions.
- **FR-018**: The report MUST include discovered capabilities,
  included and skipped plan entries, run provenance, persona
  verdicts, decision, iteration history, PR metadata, and the active
  convention pack.
- **FR-019**: The Divisor MUST conform to the Hero Interface
  Contract as an embedded hero. It MUST use standard OpenCode agents,
  commands, and artifact envelopes.
- **FR-020**: Adversary MUST perform universal security checks
  regardless of convention-pack contents.
- **FR-021**: Guard MUST enforce the Zero-Waste Mandate.
- **FR-022**: Architect MUST enforce the Neighborhood Rule.
- **FR-023**: The closed reviewer-capabilities manifest MUST be the
  only reviewer eligibility source. Content-only personas MUST NOT
  dispatch. Adversary and Guard MUST always be eligible. Other
  review personas MUST require deterministic scope intersection.
  Curator MUST require documentation or user-facing scope and MAY be
  pruned otherwise. Unknown agents MUST require an explicit valid
  review entry with intersecting scope.
- **FR-024**: `review-dispatch` MUST remain additive execution and
  provenance data. It MUST NOT replace or become the canonical
  downstream decision artifact. Child Curator runs MUST NOT create
  issues. Only the parent MAY perform one deduplicated curation
  action after consolidation and an explicit human gate.

### Key Entities

- **Review Protocol**: Discovery, manifest eligibility, planning,
  model runs, consolidation, iteration, and escalation.
- **Reviewer Capabilities Manifest**: Closed eligibility source with
  persona capability and ordered scopes.
- **Dispatch Plan**: Validated ordered entries that MAY contain one
  or more model runs for each included review persona.
- **Convention Pack**: Dynamically loaded language and project
  review rules.
- **Persona Verdict**: A consolidated persona assessment with all
  contributing run provenance.
- **Review Finding**: A severity-graded issue with location, cause,
  recommendation, and contributing runs.
- **Council Decision**: `APPROVED`, `CHANGES_REQUESTED`,
  `ESCALATED`, `INCONCLUSIVE`, or `UNAVAILABLE` in canonical
  `review-verdict` version 2.
- **Review Dispatch**: Additive execution and provenance data that
  does not replace the canonical verdict.
- **Deployment Configuration**: Project language, pack, context,
  force behavior, and Divisor-only selection.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The protocol documents all nine known personas, six
  review capabilities, three content-only capabilities, manifest
  eligibility, planning, decisions, iteration, and escalation.
- **SC-002**: The Go pack produces behavior equivalent to or better
  than the Gaze prototype on the same sample change.
- **SC-003**: The default pack produces meaningful findings in a
  project whose language has no dedicated pack.
- **SC-004**: Divisor-only Go initialization produces the expected
  agents, manifest, command, and convention pack.
- **SC-005**: TypeScript initialization deploys TypeScript-specific
  convention checks.
- **SC-006**: Markdown, canonical `review-verdict` version 2, and
  additive `review-dispatch` outputs validate. Version 1 verdicts
  remain readable, and no-success decisions block automation.
- **SC-007**: After confirmed fixes, iteration reruns every included
  validated plan run and stops at the unchanged three-iteration
  limit unless a human explicitly resolves the escalation.
- **SC-008**: Project-aware review produces more relevant findings
  than convention-pack-only review on the same sample change.

## Dependencies

### Prerequisites

- **Spec 001**: The Divisor MUST align with org principles.
- **Spec 002**: The Divisor MUST conform to artifact and naming
  contracts.

### Downstream Dependents

- **Spec 006**: Cobalt-Crush consumes canonical review decisions.
- **Spec 007**: Mx F consumes canonical review decisions.
- **Spec 008**: The Divisor acts as an orchestration gate.
- **Spec 009**: The shared model defines artifact schemas and
  compatibility.
- **Spec 026**: Curator owns documentation and content opportunity
  assessment under manifest-based eligibility.

### Reference Implementation

The Gaze prototype agents and review command remain the historical
reference. The Divisor framework MUST preserve or improve their Go
review behavior while using dynamic convention packs, manifest
eligibility, and validated review plans.

```text
The Divisor Framework (embedded in unbound binary)
┌──────────────────────────────────────────────────────┐
│ Review protocol and closed reviewer manifest        │
│ Convention packs and validated dispatch plans       │
│ Canonical review-verdict and additive dispatch data │
└──────────────────────────┬───────────────────────────┘
                           │ unbound init
                           ▼
┌──────────────────────────────────────────────────────┐
│ Project deployment                                  │
│ Six review-capable and three content-only personas  │
│ Review command, manifest, packs, and artifact rules │
└──────────────────────────────────────────────────────┘
```
