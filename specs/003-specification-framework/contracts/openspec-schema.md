# Contract: OpenSpec Custom Schema (`unbound-force`)

**Spec**: 003-specification-framework
**Date**: 2026-03-08

## Overview

The `unbound-force` custom OpenSpec schema extends the
built-in `spec-driven` schema with constitution alignment
requirements. It is the standard schema for all Unbound
Force repositories.

## Schema Location

```text
openspec/schemas/unbound-force/
+-- schema.yaml
+-- templates/
    +-- proposal.md
    +-- spec.md
    +-- design.md
    +-- tasks.md
```

## schema.yaml

```yaml
name: unbound-force
description: >
  Unbound Force specification schema with constitution
  alignment. Extends the spec-driven workflow with
  mandatory governance checks.

artifacts:
  - id: proposal
    generates: proposal.md
    description: >
      Change proposal with constitution alignment
    template: proposal.md
    instruction: >
      Create a change proposal. You MUST include a
      Constitution Alignment section assessing this
      change against all three org constitution
      principles. Read the constitution from
      .specify/memory/constitution.md for the full
      principle definitions.
    requires: []

  - id: specs
    generates: "specs/**/*.md"
    description: >
      Delta specs describing requirement changes
    template: spec.md
    instruction: >
      Write delta specs using ADDED/MODIFIED/REMOVED
      sections. Use RFC 2119 language (MUST/SHALL/
      SHOULD/MAY) for all requirements. Include
      Given/When/Then scenarios.
    requires: [proposal]

  - id: design
    generates: design.md
    description: >
      Technical design and architecture decisions
    template: design.md
    instruction: >
      Document the technical approach. Reference the
      constitution alignment from the proposal. Note
      any design decisions that relate to Autonomous
      Collaboration, Composability First, or
      Observable Quality.
    requires: [proposal]

  - id: tasks
    generates: tasks.md
    description: >
      Implementation task checklist
    template: tasks.md
    instruction: >
      Break the design into implementable tasks with
      checkboxes. Group related tasks. Include a task
      for verifying constitution alignment if the
      proposal identified relevant principles.
    requires: [specs, design]

apply:
  requires: [tasks]
  tracks: tasks.md
  instruction: >
    Implement tasks from tasks.md. Check off each
    task as you complete it. Verify that the
    implementation maintains constitution alignment
    as documented in the proposal.
```

## Templates

### proposal.md

```markdown
## Why

<!-- Motivation for this change -->

## What Changes

<!-- Specific changes being proposed -->

## Capabilities

### New Capabilities
- `<name>`: <description>

### Modified Capabilities
- `<existing-name>`: <what changes>

### Removed Capabilities
- `<name>`: <reason for removal>

## Impact

<!-- Affected systems, files, or behaviors -->

## Constitution Alignment

Assessed against Unbound Force org constitution v1.0.0.

### I. Autonomous Collaboration

**Assessment**: PASS | N/A

<!-- How does this change affect artifact-based
communication? Does it maintain self-describing
outputs? -->

### II. Composability First

**Assessment**: PASS | N/A

<!-- Does this change maintain standalone
functionality? Does it avoid introducing mandatory
dependencies? -->

### III. Observable Quality

**Assessment**: PASS | N/A

<!-- Does this change produce machine-parseable
output? Does it maintain provenance metadata? -->
```

### spec.md

```markdown
## ADDED Requirements

### Requirement: <!-- name -->

<!-- requirement text using RFC 2119 language -->

#### Scenario: <!-- name -->
- **GIVEN** <!-- precondition -->
- **WHEN** <!-- action -->
- **THEN** <!-- expected outcome -->

## MODIFIED Requirements

### Requirement: <!-- name -->

<!-- new text (note: "Previously: <old text>") -->

## REMOVED Requirements

### Requirement: <!-- name -->

<!-- reason for removal -->
```

### design.md

```markdown
## Context

<!-- Current state and motivation -->

## Goals / Non-Goals

### Goals
- <!-- what this design achieves -->

### Non-Goals
- <!-- what is explicitly out of scope -->

## Decisions

<!-- Key technical decisions with rationale -->

## Risks / Trade-offs

<!-- Known risks and accepted trade-offs -->
```

### tasks.md

```markdown
## 1. <!-- Task Group -->

- [ ] 1.1 <!-- task description -->
- [ ] 1.2 <!-- task description -->

## 2. <!-- Task Group -->

- [ ] 2.1 <!-- task description -->
```

## Default Configuration

`openspec/config.yaml` (installed alongside schema):

```yaml
schema: unbound-force

context: |
  This project follows the Unbound Force org constitution
  v1.0.0. Three core principles govern all work:

  I. Autonomous Collaboration: Heroes collaborate through
     well-defined artifacts, not runtime coupling.
  II. Composability First: Every hero is independently
      installable and usable alone.
  III. Observable Quality: Every hero produces
       machine-parseable output with provenance.

  All changes MUST align with these principles.
  Constitution violations are CRITICAL severity.

  Full constitution: .specify/memory/constitution.md

rules:
  proposal:
    - MUST include Constitution Alignment section
    - Each principle needs PASS or N/A with justification
  specs:
    - Use RFC 2119 language (MUST/SHALL/SHOULD/MAY)
    - Use Given/When/Then for all scenarios
  design:
    - Reference constitution alignment from proposal
  tasks:
    - Include verification task for constitution alignment
```
