## MODIFIED Requirements

### Requirement: Forwarded Environment Variables

The `forwardedAPIKeys` list MUST include
`ANTHROPIC_VERTEX_PROJECT_ID` and
`CLAUDE_CODE_USE_VERTEX` in addition to the existing
6 entries.

#### Scenario: Vertex AI user starts sandbox

- **GIVEN** `ANTHROPIC_VERTEX_PROJECT_ID` and
  `CLAUDE_CODE_USE_VERTEX` are set in the host env
- **WHEN** the engineer runs `uf sandbox start`
- **THEN** both vars are forwarded to the container
  via `-e` flags

#### Scenario: Non-Vertex user unaffected

- **GIVEN** `ANTHROPIC_VERTEX_PROJECT_ID` is NOT set
- **WHEN** the engineer runs `uf sandbox start`
- **THEN** the var is not included in the `-e` flags
  (existing conditional check handles this)

## REMOVED Requirements

None.
