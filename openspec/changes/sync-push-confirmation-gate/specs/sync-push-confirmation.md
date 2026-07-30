## ADDED Requirements

### Requirement: FR-001 Dry-run preview mode

The `sync-push` CLI command MUST accept a `--dry-run` flag.
When `--dry-run` is set, the command MUST report all pending
actions (CREATE or UPDATE) without executing any GitHub API
calls. The output MUST include the action type, backlog item
ID, item title, and (for updates) the existing GitHub Issue
number.

#### Scenario: Dry-run with items to create

- **GIVEN** the backlog contains item BI-001 with no
  associated GitHub Issue number
- **WHEN** the user runs `sync-push --dry-run`
- **THEN** the output MUST include a line indicating
  `CREATE` for BI-001 with its title
- **AND** no GitHub Issue MUST be created

#### Scenario: Dry-run with items to update

- **GIVEN** the backlog contains item BI-002 with
  GitHubIssueNumber set to 42
- **WHEN** the user runs `sync-push --dry-run`
- **THEN** the output MUST include a line indicating
  `UPDATE` for BI-002 referencing Issue #42
- **AND** no GitHub Issue MUST be modified

#### Scenario: Dry-run with single item filter

- **GIVEN** the backlog contains items BI-001 and BI-002
- **WHEN** the user runs `sync-push --dry-run BI-001`
- **THEN** the output MUST include only BI-001
- **AND** BI-002 MUST NOT appear in the output

#### Scenario: Dry-run with empty backlog

- **GIVEN** the backlog contains no items
- **WHEN** the user runs `sync-push --dry-run`
- **THEN** the output MUST indicate that no items are
  pending sync

#### Scenario: Dry-run with non-existent item

- **GIVEN** the backlog does not contain item BI-999
- **WHEN** the user runs `sync-push --dry-run BI-999`
- **THEN** the command MUST return an error
- **AND** no GitHub Issue MUST be created or modified

#### Scenario: Dry-run does not require gh CLI

- **GIVEN** the `gh` CLI is not installed or not in PATH
- **WHEN** the user runs `sync-push --dry-run`
- **THEN** the preview MUST still be generated
  successfully
- **AND** no error about `gh` MUST be reported

When `--dry-run` is set, the command MUST NOT invoke the
`GHRunner` or any GitHub API calls. Dry-run MUST
propagate backlog read errors identically to normal mode.

### Requirement: FR-002 Confirmation gate in command file

The `/muti-mind.sync-push` command file MUST include a
mandatory confirmation step using the AskUserQuestion tool
before invoking the Go backend for actual execution.

The confirmation step MUST:
1. Invoke the Go backend with `--dry-run` to obtain the
   preview
2. Present the preview output to the user
3. Use AskUserQuestion with options including at minimum
   "Yes -- sync to GitHub" and "No -- abort"
4. Proceed with execution ONLY if the user selects the
   affirmative option

#### Scenario: User confirms sync

- **GIVEN** the agent has presented the dry-run preview
- **WHEN** the user selects "Yes -- sync to GitHub"
- **THEN** the agent MUST invoke the Go backend without
  `--dry-run` to execute the sync
- **AND** the results MUST be displayed to the user

#### Scenario: User aborts sync

- **GIVEN** the agent has presented the dry-run preview
- **WHEN** the user selects "No -- abort"
- **THEN** the agent MUST NOT invoke the Go backend
- **AND** the agent MUST inform the user that the sync
  was cancelled

#### Scenario: No items to sync

- **GIVEN** the dry-run preview reports no items pending
- **WHEN** the preview is presented to the user
- **THEN** the agent SHOULD skip the confirmation step
- **AND** inform the user that there is nothing to sync

### Requirement: FR-003 Dry-run summary format

The dry-run output MUST use a structured text format with
one line per item. Each line MUST follow the pattern:

```
<ACTION>  <ITEM_ID>  <TITLE>  [Issue #<NUMBER>]
```

Where ACTION is `CREATE` or `UPDATE`, and the Issue number
is included only for UPDATE actions.

The output MUST end with a summary line:

```
Total: <N> item(s) (<C> to create, <U> to update)
```

#### Scenario: Mixed create and update output

- **GIVEN** the backlog contains BI-001 (no issue) and
  BI-002 (issue #42)
- **WHEN** the user runs `sync-push --dry-run`
- **THEN** the output MUST contain:
  - A CREATE line for BI-001
  - An UPDATE line for BI-002 referencing Issue #42
  - A summary line: "Total: 2 item(s) (1 to create, 1 to update)"

## MODIFIED Requirements

### Requirement: FR-004 Sync-push command file instructions

Previously: The command file contained a single step that
directly invoked the Go backend without confirmation.

The command file instructions MUST now follow this sequence:
1. Invoke Go backend with `--dry-run` to get preview
2. Present preview to user
3. Gate on AskUserQuestion confirmation
4. Execute actual sync only on confirmation

## REMOVED Requirements

None.
