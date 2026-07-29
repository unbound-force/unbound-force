## ADDED Requirements

### Requirement: Structured secrets confirmation gate

The `/finale` command's secrets check in Step 2 MUST use the
**AskUserQuestion tool** with predefined options when potential
secret files are detected. The agent MUST NOT proceed to
`git add .` without receiving an explicit structured response
from the user.

Options presented MUST be:
1. "Yes -- stage all files including flagged ones"
2. "No -- stop and let me handle it manually"

If the user selects "No", the agent MUST stop immediately.
The agent MUST NOT fall back to prose-based confirmation or
proceed without user selection.

#### Scenario: Secret files detected and user approves

- **GIVEN** the working tree contains files matching secret
  patterns (`.env`, `*.key`, `*.pem`, `credentials.json`,
  `secrets.json`)
- **WHEN** the agent reaches the secrets check in Step 2
- **THEN** the agent MUST present the **AskUserQuestion tool**
  with the two predefined options
- **AND** if the user selects "Yes -- stage all files
  including flagged ones", the agent MUST proceed to
  `git add .`

#### Scenario: Secret files detected and user declines

- **GIVEN** the working tree contains files matching secret
  patterns
- **WHEN** the agent presents the **AskUserQuestion tool**
  and the user selects "No -- stop and let me handle it
  manually"
- **THEN** the agent MUST stop immediately
- **AND** the agent MUST NOT run `git add .`
- **AND** the agent MUST NOT attempt any alternative staging
  approach

#### Scenario: No secret files detected

- **GIVEN** the working tree contains only non-secret files
- **WHEN** the agent reaches the secrets check in Step 2
- **THEN** the agent MUST proceed directly to `git add .`
  without presenting the AskUserQuestion prompt

## MODIFIED Requirements

### Requirement: Secrets check confirmation mechanism

The secrets check confirmation MUST use the
**AskUserQuestion tool** instead of prose-based instructions.

Previously: "Ask for confirmation. If the user declines,
stop and let them handle it manually." (free-text prose)

### Requirement: Scaffold asset synchronization

The scaffold asset at
`internal/scaffold/assets/opencode/commands/finale.md`
MUST be updated with the identical change to maintain
byte-identical copies. The `TestEmbeddedAssets_MatchSource`
test MUST verify this.

## REMOVED Requirements

None.
<!-- scaffolded by uf vdev -->
