# Permission Rules

## ADDED Requirements

### FR-001: Allow Read-Only `gh api` Calls

Read-only `gh api` calls MUST execute without
prompting for approval.

#### Scenario: Default (implicit GET) gh api call
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs `gh api /repos/owner/repo`
- **THEN** the command MUST execute without prompting

#### Scenario: Explicit GET gh api call
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo -X GET`
- **THEN** the command MUST execute without prompting

### FR-002: Gate Mutating `gh api` Methods

Mutating HTTP methods MUST prompt for approval.
All flag forms MUST be covered: `-X` (spaced), `-X`
(glued), `--method` (spaced), `--method=` (equals).
Both uppercase and lowercase method names MUST be
covered due to case-sensitive matching on Linux/macOS.

#### Scenario: gh api with -X POST
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues -X POST`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with -X DELETE
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues/1 -X DELETE`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with -X PATCH
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues/1 -X PATCH`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with -X PUT
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/merges -X PUT`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with --method POST (long form)
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues --method POST`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with --method DELETE (long form)
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues/1 --method DELETE`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with lowercase -X post
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues -X post`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with glued -XPOST (no space)
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues -XPOST`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with glued lowercase -Xpost
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues -Xpost`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with --method=POST (equals form)
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues --method=POST`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with --method=delete (lowercase)
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues/1 --method=delete`
- **THEN** the command MUST prompt for approval

#### Known Limitation: Flag-before-endpoint ordering
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api -X POST /repos/owner/repo`
- **THEN** the command MAY execute without prompting
  (pattern may not match reversed flag ordering;
  mitigated by `<protect>` tags from PR #499)

### FR-003: Gate Data-Sending Flags

Commands with data-sending flags MUST prompt for
approval regardless of HTTP method. Both short forms
(`-f`, `-F`) and long forms (`--field`, `--raw-field`)
MUST be covered, including `=` syntax variants.

#### Scenario: gh api with -f flag
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues -f title=bug`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with -F flag
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues -F body=@file.md`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with --field (long form of -F)
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues --field body=text`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with --field= (equals form)
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues --field=body=text`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with --raw-field (long form of -f)
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues --raw-field title=bug`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with --raw-field= (equals form)
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo --raw-field=query=val`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with --input flag
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo/issues --input data.json`
- **THEN** the command MUST prompt for approval

#### Scenario: gh api with GET and data-sending flag
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the agent runs
  `gh api /repos/owner/repo -X GET -f query=val`
- **THEN** the command MUST prompt for approval

### FR-004: Preserve Original Mutation Guards

All original non-api mutation guards from the
upstream permission block MUST be preserved unchanged.

The following 13 ask rules MUST be present:
- `"gh issue create*": "ask"`
- `"gh issue edit*": "ask"`
- `"gh issue close*": "ask"`
- `"gh issue comment*": "ask"`
- `"gh pr create*": "ask"`
- `"gh pr merge*": "ask"`
- `"gh pr close*": "ask"`
- `"gh pr comment*": "ask"`
- `"gh pr edit*": "ask"`
- `"gh pr review*": "ask"`
- `"git push*": "ask"`
- `"git commit*": "ask"`
- `"rm *": "ask"`

The global default `"*": "allow"` MUST also be present.

#### Scenario: Original rules preserved
- **GIVEN** a `permission.bash` block in `opencode.json`
- **WHEN** the block is inspected
- **THEN** all 13 original ask rules and the global
  default MUST be present and unchanged

## MODIFIED Requirements

No requirements are modified.

## REMOVED Requirements

No requirements are removed.
