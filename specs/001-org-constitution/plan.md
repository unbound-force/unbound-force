# Implementation Plan: Unbound Force Organization Constitution

**Branch**: `001-org-constitution` | **Date**: 2026-02-25 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-org-constitution/spec.md`

## Summary

Ratify the Unbound Force organization constitution with three core principles (Autonomous Collaboration, Composability First, Observable Quality), a governance model, a development workflow, and a hero constitution alignment section. Implement an OpenCode agent that performs agent-assisted alignment checking of hero constitutions against the org constitution.

This spec produces two primary deliverables:
1. The ratified constitution document (`.specify/memory/constitution.md`) -- already complete as v1.0.0.
2. An OpenCode agent (`constitution-check`) and command (`/constitution-check`) that validates hero constitutions against the org constitution, producing a structured alignment report.

## Technical Context

**Language/Version**: Markdown (constitution document) + OpenCode agent configuration (Markdown agent files)
**Primary Dependencies**: OpenCode (agent runtime), speckit (pipeline integration)
**Storage**: Filesystem only -- constitution at `.specify/memory/constitution.md`, agent files in `.opencode/agents/` and `.opencode/command/`
**Testing**: Manual validation against success criteria SC-001 through SC-006. Agent output validated by running against Gaze and Website constitutions.
**Target Platform**: Any project repository using OpenCode
**Project Type**: Governance document + OpenCode agent configuration
**Performance Goals**: N/A (document + agent, not runtime software)
**Constraints**: Constitution document MUST be under 500 lines (SC-006). Agent MUST produce structured output parseable by other heroes.
**Scale/Scope**: Governs 5 hero repositories + 1 website repository. Agent runs per-hero, not at scale.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Check

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Autonomous Collaboration | PASS | The constitution is a standalone artifact. The alignment agent produces self-describing output (structured report with provenance). No runtime coupling between the org constitution and hero constitutions -- alignment is checked asynchronously via artifact comparison. |
| II. Composability First | PASS | The constitution is independently usable -- any repo can read and follow it without installing any hero. The alignment agent is an optional enhancement (alignment can be checked manually without it). No mandatory dependencies. |
| III. Observable Quality | PASS | The alignment agent produces machine-parseable output (structured report with hero_name, org_constitution_version, findings[], overall_status). The constitution itself is versioned (semver) with ratification and amendment dates. Quality claims (alignment status) are backed by the agent's structured findings, not subjective judgment. |

**Gate result**: PASS -- all three principles satisfied. Proceeding to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/001-org-constitution/
├── spec.md              # Feature specification (complete)
├── plan.md              # This file
├── research.md          # Phase 0: research on alignment agent patterns
├── data-model.md        # Phase 1: alignment report data model
└── quickstart.md        # Phase 1: how to use the constitution and agent
```

### Source Code (repository root)

```text
.specify/
└── memory/
    └── constitution.md          # The ratified constitution (US1 deliverable)

.opencode/
├── agents/
│   └── constitution-check.md    # Alignment checking agent (US2 deliverable)
└── command/
    └── constitution-check.md    # /constitution-check command (US2 deliverable)
```

**Structure Decision**: No traditional `src/` or `tests/` directories. This feature produces governance documents and OpenCode agent configurations, not compiled software. The constitution lives in `.specify/memory/` per speckit convention. The alignment agent lives in `.opencode/` per OpenCode convention.

## Phase 0: Research

### Research Tasks

1. **Alignment agent design patterns**: How should an OpenCode agent compare two constitution documents and produce a structured finding report? What existing patterns (from the Gaze reviewer agents) can be adapted?
2. **Constitution Check integration**: How does the alignment agent integrate with the speckit pipeline's Constitution Check gate (FR-008)? Should the `/constitution-check` command be separate from or integrated into `/speckit.plan`?
3. **Structured report format**: What format should the alignment report use? Should it conform to the inter-hero artifact envelope (Spec 002/009) or use a simpler format since those specs are not yet finalized?

### Research Decisions

See `research.md` for full details.

## Phase 1: Design

### Deliverables

1. **data-model.md**: Alignment report structure, finding categories, status values
2. **quickstart.md**: How to use the constitution and run alignment checks
3. **Agent files**: `constitution-check.md` agent and command definitions

### Constitution Check (Post-Design Re-evaluation)

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Autonomous Collaboration | PASS | The alignment agent operates as a read-only subagent that produces a self-describing Markdown report. No runtime coupling -- it reads two files and outputs findings. The report includes provenance (hero name, versions, timestamp). |
| II. Composability First | PASS | The alignment agent is optional. Constitution alignment can be verified manually without the agent. The agent has no dependencies on other heroes. It works in any repo with OpenCode installed. |
| III. Observable Quality | PASS | The data model defines structured output (Alignment Check, Alignment Finding entities) with enumerated status values (ALIGNED/GAP/CONTRADICTION). The report includes measurable summary counts. The agent uses deterministic settings (temperature 0.1). |

**Post-design gate result**: PASS -- all three principles remain satisfied after design phase. No regressions from pre-design check.

## Complexity Tracking

No constitution violations to justify. All three principles pass cleanly.
