## ADDED Requirements

_None._

## MODIFIED Requirements

### Requirement: Secrets file confirmation gate

The `/finale` command's Step 2 ("Check for Changes to
Commit") MUST use the **AskUserQuestion tool** with
structured options when potential secret files are
detected, instead of prose-only confirmation.

When potential secret files are found, the agent MUST
present the user with the **AskUserQuestion tool** using
options:
- "Yes -- stage all files including flagged ones"
- "No -- stop and let me handle it manually"

If the user selects "No", the agent MUST NOT run
`git add .` and MUST stop execution.

Previously: "Ask for confirmation. If the user declines,
stop and let them handle it manually." (prose-only, no
tool call specified)

#### Scenario: Secret files detected and user approves

- **GIVEN** the working tree contains `.env.local` and
  `credentials.json` alongside modified source files
- **WHEN** `/finale` runs Step 2 and detects these files
  match secret patterns
- **THEN** the agent MUST display the warning message
  listing the flagged files AND present the
  **AskUserQuestion tool** with options "Yes -- stage
  all files including flagged ones" and "No -- stop and
  let me handle it manually"
- **AND** if the user selects "Yes", the agent MUST
  proceed to `git add .`

#### Scenario: Secret files detected and user declines

- **GIVEN** the working tree contains `secrets.json`
  alongside modified source files
- **WHEN** `/finale` runs Step 2, detects the file, and
  presents the AskUserQuestion tool
- **THEN** if the user selects "No -- stop and let me
  handle it manually", the agent MUST NOT run `git add .`
- **AND** the agent MUST stop execution and report that
  the user chose to handle staging manually

#### Scenario: No secret files detected

- **GIVEN** the working tree contains only `.go`, `.md`,
  and `.yaml` files with no secret-pattern matches
- **WHEN** `/finale` runs Step 2
- **THEN** the agent MUST proceed directly to
  `git add .` without presenting the secrets
  confirmation prompt

### Requirement: Scaffold asset byte-identity

The file `internal/scaffold/assets/opencode/commands/finale.md`
MUST be byte-identical to `.opencode/commands/finale.md`
after all edits.

Previously: This requirement already exists and is
enforced by `TestEmbeddedAssets_MatchSource`. No change
to the requirement itself, but the scaffold asset MUST
be updated to reflect the modified command file.

#### Scenario: Scaffold drift detection

- **GIVEN** `.opencode/commands/finale.md` has been
  modified with the AskUserQuestion gate
- **WHEN** `go test ./internal/scaffold/... -count=1`
  is run
- **THEN** `TestEmbeddedAssets_MatchSource` MUST pass,
  confirming the scaffold asset is byte-identical

## REMOVED Requirements

_None._
<!-- scaffolded by uf vdev -->
