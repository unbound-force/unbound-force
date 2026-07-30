## ADDED Requirements

### Requirement: Pre-conditions subsection ordering

The `/cobalt-crush` command file MUST place branch safety
pre-conditions in a `### Pre-conditions` subsection
immediately after the `## Instructions` heading and before
any instruction subsections that perform branch operations.

#### Scenario: Agent processes instructions sequentially

- **GIVEN** an agent reads `cobalt-crush.md` from top to
  bottom
- **WHEN** the agent reaches the first branch operation
  in `### When no arguments are provided`
- **THEN** the agent SHALL have already encountered and
  processed the branch safety pre-conditions in the
  `### Pre-conditions` subsection

## MODIFIED Requirements

### Requirement: Branch safety guardrails location

The branch safety guardrails content MUST appear within
the `## Instructions` section as a `### Pre-conditions`
subsection, not as a separate `## Branch Safety Guardrails`
section at the end of the file.

Previously: Branch safety guardrails were in a standalone
`## Branch Safety Guardrails` section at lines 97-111,
after all instruction content.

#### Scenario: Guardrail content preserved after relocation

- **GIVEN** the branch safety guardrails have been moved
  to `### Pre-conditions`
- **WHEN** the content is compared to the original
  `## Branch Safety Guardrails` section
- **THEN** all four guardrail rules SHALL be present
  with identical wording

#### Scenario: No duplicate guardrail sections

- **GIVEN** the guardrails have been moved to
  `### Pre-conditions`
- **WHEN** the file is inspected
- **THEN** there SHALL NOT be a remaining
  `## Branch Safety Guardrails` section at the end of
  the file

## REMOVED Requirements

### Requirement: Standalone Branch Safety Guardrails section

The standalone `## Branch Safety Guardrails` section at the
end of the file is removed. Its content is relocated to the
`### Pre-conditions` subsection within `## Instructions`.
Reason: T2 weakness -- CRITICAL rules MUST NOT appear after
the actions they govern.
