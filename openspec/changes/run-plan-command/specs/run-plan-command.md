## ADDED Requirements

### Requirement: Run-Plan Command Definition

The system SHALL provide a `/uf.run-pipeline-plan` slash command that executes a prioritized work plan through the Unbound Force pipeline. The command definition SHALL exist at `.opencode/commands/uf.run-pipeline-plan.md` (canonical) and `internal/scaffold/assets/opencode/commands/uf.run-pipeline-plan.md` (embedded copy). The two files MUST be byte-identical.

#### Scenario: Command file exists in both locations
- **GIVEN** the implementation is complete
- **WHEN** `TestEmbeddedAssets_MatchSource` runs
- **THEN** the canonical and embedded copies of `uf.run-pipeline-plan.md` MUST be byte-identical, and the test MUST pass

#### Scenario: Scaffold drift test manifest includes new command
- **GIVEN** the implementation is complete
- **WHEN** `expectedAssetPaths` in `internal/scaffold/scaffold_test.go` is inspected
- **THEN** it MUST contain the entry `"opencode/commands/uf.run-pipeline-plan.md"`

### Requirement: Input Detection

The orchestrator MUST detect the input type from `$ARGUMENTS`. If `$ARGUMENTS` is an existing file path, it MUST be parsed as a plan file. Otherwise, it MUST be treated as a space-separated list of GitHub issue references (`#NNN` or full GitHub URLs).

#### Scenario: File path input
- **GIVEN** `$ARGUMENTS` is `sprint-42.md`
- **AND** `sprint-42.md` exists as a file in the working directory
- **WHEN** `/uf.run-pipeline-plan` is invoked
- **THEN** the orchestrator MUST parse `sprint-42.md` as a plan file

#### Scenario: Issue ref input
- **GIVEN** `$ARGUMENTS` is `#123 #456 #789`
- **AND** no file named `#123 #456 #789` exists
- **WHEN** `/uf.run-pipeline-plan` is invoked
- **THEN** the orchestrator MUST treat the input as three GitHub issue references

#### Scenario: Mixed URL input
- **GIVEN** `$ARGUMENTS` is `https://github.com/unbound-force/unbound-force/issues/42 #99`
- **WHEN** `/uf.run-pipeline-plan` is invoked
- **THEN** the orchestrator MUST parse both the URL and the `#NNN` ref as issue references

### Requirement: Plan File Contract

A plan file MUST be a numbered markdown list. Each item MUST carry: one or more issue references (`#NNN`), a kebab-case `Change:` name, and an optional one-line description. Non-mechanical items (investigations, policy decisions) MUST be marked with a tag (e.g., `[INVESTIGATION]` or `[POLICY]`).

#### Scenario: Valid plan file parsing
- **GIVEN** a plan file containing:
  ```
  1. #42 Change: add-run-plan-command — Create /uf.run-pipeline-plan slash command
  2. #99 Change: fix-scaffold-drift — Sync embedded asset copies
  ```
- **WHEN** the orchestrator parses the plan file
- **THEN** it MUST extract two work items with issue refs `#42` and `#99`, change names `add-run-plan-command` and `fix-scaffold-drift`, and their descriptions

#### Scenario: Non-mechanical item flagged
- **GIVEN** a plan file containing:
  ```
  1. #55 [INVESTIGATION] Investigate CI flakiness in TestEmbeddedAssets
  ```
- **WHEN** the orchestrator parses the plan file
- **THEN** it MUST flag this item as non-mechanical and surface it for human resolution

### Requirement: Phase 0 — Input Resolution and Confirmation

Phase 0 MUST run in the orchestrator context (no subagents). For issue-ref input, the orchestrator MUST resolve each issue via `gh issue view <N> --json title,body` to derive a kebab-case change name and description. For a bare issue list (no grouping information), the orchestrator MUST ask via `question` how to group issues into work items (default: one PR per issue). The orchestrator MUST present the full ordered work-item table via `question` for human confirmation. Non-mechanical items MUST be flagged for explicit human resolution before the pipeline runs.

#### Scenario: Issue ref resolution
- **GIVEN** input `#42 #99`
- **WHEN** Phase 0 executes
- **THEN** the orchestrator MUST run `gh issue view 42 --json title,body` and `gh issue view 99 --json title,body`
- **AND** derive kebab-case change names and descriptions from the issue metadata

#### Scenario: Work-item table confirmation
- **GIVEN** Phase 0 has resolved all issues and derived change names
- **WHEN** the orchestrator presents the work-item table
- **THEN** it MUST invoke `question` with the full ordered table
- **AND** the human MUST confirm before the pipeline proceeds

#### Scenario: Non-mechanical item blocking
- **GIVEN** a plan containing `[INVESTIGATION] Investigate CI flakiness`
- **WHEN** Phase 0 presents the work-item table
- **THEN** the investigation item MUST be flagged
- **AND** the human MUST explicitly resolve it (remove, convert to mechanical, or defer) before the pipeline runs

### Requirement: Per-Item Pipeline Protocol

For each work item, the orchestrator MUST execute three pipeline phases sequentially, each in a fresh `general` subagent:
1. Subagent 1: `/opsx-propose` with input = change name + description
2. Subagent 2: `/uf.unleash` (OpenSpec mode; Step 8 review council MUST NOT be skipped)
3. Subagent 3: `/uf.finale` (commit + PR MUST include `Closes #N`)

The orchestrator MUST verify disk handoff (artifacts on disk, branch checked out) before spawning each next subagent. The orchestrator MUST confirm return to `main` + clean working tree between items.

#### Scenario: Fresh subagent per phase
- **GIVEN** work item "add-run-plan-command" is being executed
- **WHEN** the `/opsx-propose` phase completes
- **THEN** the subagent MUST terminate
- **AND** a NEW `general` subagent MUST be spawned for `/uf.unleash`
- **AND** the new subagent MUST NOT inherit context from the previous subagent

#### Scenario: Disk handoff verification
- **GIVEN** `/opsx-propose` has completed in its subagent
- **WHEN** the orchestrator prepares to spawn the `/uf.unleash` subagent
- **THEN** it MUST verify that spec artifacts exist on disk (proposal.md, design.md, specs/, tasks.md)
- **AND** it MUST verify the correct branch is checked out

#### Scenario: Clean tree between items
- **GIVEN** work item 1 has completed (PR created)
- **WHEN** the orchestrator prepares to start work item 2
- **THEN** the working directory MUST be on `main` with a clean working tree (`git status --short` is empty)

### Requirement: Gate Relay Protocol

When a subagent encounters a mandatory human confirmation gate, it MUST STOP execution and return a structured value: `GATE:<name>` + the exact question + full proposed text + options. The orchestrator MUST relay the gate through the `question` tool. After the human responds, the orchestrator MUST resume the SAME subagent via `task_id` with the human's answer. Subagents MUST NOT invoke `question` directly.

#### Scenario: Subagent encounters a gate
- **GIVEN** a `/uf.unleash` subagent is implementing tasks
- **AND** it encounters a MANDATORY human confirmation gate (e.g., spec review approval)
- **WHEN** the gate is reached
- **THEN** the subagent MUST STOP and return: `GATE:spec-review` + the exact question + full proposed text + options (e.g., `["Approve", "Request changes"]`)
- **AND** the subagent MUST NOT invoke `question` itself

#### Scenario: Orchestrator relays gate
- **GIVEN** a subagent has returned `GATE:spec-review` + question + text + options
- **WHEN** the orchestrator receives the gate return
- **THEN** it MUST invoke `question` with the exact question, full text, and options
- **AND** after the human responds, it MUST resume the same subagent via `task_id` with the answer

#### Scenario: Subagent forbidden from direct question invocation
- **GIVEN** a subagent is executing a pipeline phase
- **WHEN** it needs human input
- **THEN** it MUST NOT invoke the `question` tool directly
- **AND** it MUST return a `GATE:<name>` structured value instead

### Requirement: Guardrails

The command MUST enforce the following guardrails:
- **Scaffold drift sync**: Any canonical `.opencode/` or `openspec/` change MUST sync its `internal/scaffold/assets/...` copy, else `TestEmbeddedAssets_MatchSource` fails
- **Explicit staging**: The command MUST stage files explicitly. `git add -A` and `git add .` are forbidden
- **Stray file removal**: Any untracked `internal/scaffold/AGENTS.md` MUST be removed before staging (never commit it)
- **No force-push**: `git push --force` is forbidden. Only `git push --force-with-lease` is allowed
- **PR targeting**: PRs MUST target the parent `unbound-force/unbound-force` repo with head `jflowers:<branch>`
- **Closes linkage**: The commit message and PR body MUST include `Closes #N` for each issue in the work item
- **No auto-merge**: The PR MUST NOT be merged automatically; it stays open for human review

#### Scenario: Explicit staging enforced
- **GIVEN** the `/uf.finale` subagent is staging changes
- **WHEN** it prepares to commit
- **THEN** it MUST use explicit file paths in `git add` (e.g., `git add .opencode/commands/uf.run-pipeline-plan.md internal/scaffold/assets/opencode/commands/uf.run-pipeline-plan.md internal/scaffold/scaffold_test.go`)
- **AND** it MUST NOT use `git add -A` or `git add .`

#### Scenario: PR targeting parent repo
- **GIVEN** the `/uf.finale` subagent is creating a PR
- **WHEN** it invokes `gh pr create`
- **THEN** the PR MUST target `--repo unbound-force/unbound-force` with `--head jflowers:<branch>`
- **AND** the PR body MUST include `Closes #N`

#### Scenario: No force-push
- **GIVEN** the `/uf.finale` subagent is pushing
- **WHEN** it invokes `git push`
- **THEN** it MUST NOT use `--force`
- **AND** it MAY use `--force-with-lease` if needed

### Requirement: Progress Tracking

The orchestrator MUST maintain a `todowrite` tracker with one entry per work item. An entry MUST be marked complete only after the item's PR is created AND CI is green.

#### Scenario: Todowrite entry per item
- **GIVEN** a plan with 3 work items
- **WHEN** the orchestrator starts execution
- **THEN** `todowrite` MUST contain 3 entries, one per work item

#### Scenario: Entry marked complete
- **GIVEN** work item 1's PR has been created
- **AND** CI checks on the PR are green
- **WHEN** the orchestrator updates `todowrite`
- **THEN** the entry for work item 1 MUST be marked complete

#### Scenario: Entry not marked complete prematurely
- **GIVEN** work item 2's PR has been created
- **AND** CI checks are still pending
- **WHEN** the orchestrator checks `todowrite`
- **THEN** the entry for work item 2 MUST NOT be marked complete

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
