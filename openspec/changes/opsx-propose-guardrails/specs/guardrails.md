## ADDED Requirements

### Requirement: /opsx-propose Guardrails

The `/uf-init` slash command MUST include a
customization step that injects a `## Guardrails`
section into `.opencode/command/opsx-propose.md` if
one does not already exist.

The guardrails MUST state:
- The command creates artifacts only
- The command must not implement code changes
- The command must not commit, push, or create PRs
- The command must stop after artifacts are complete
  and prompt the user

Injection is idempotent — if the section exists, skip.

#### Scenario: /uf-init adds guardrails to opsx-propose

- **GIVEN** `opsx-propose.md` exists without a
  `## Guardrails` section
- **WHEN** the engineer runs `/uf-init`
- **THEN** a `## Guardrails` section is appended

#### Scenario: /uf-init skips existing guardrails

- **GIVEN** `opsx-propose.md` already has a
  `## Guardrails` section
- **WHEN** the engineer runs `/uf-init`
- **THEN** the existing section is not duplicated

## REMOVED Requirements

None.
