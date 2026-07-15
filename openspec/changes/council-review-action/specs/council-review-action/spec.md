# Council Review Action Specification

## ADDED Requirements

### Requirement: Action discovers Divisor agents dynamically

The action SHALL discover Divisor persona definitions by globbing
`.opencode/agents/divisor-*.md` in the checked-out repository. It
SHALL construct `--agents` JSON from the discovered files.

#### Scenario: Multiple personas discovered

- **WHEN** the repo contains 5+ `divisor-*.md` files
- **THEN** the action builds an `--agents` JSON object with one
  entry per persona and invokes `claude -p --agents`

#### Scenario: Zero personas discovered

- **WHEN** the repo contains no `divisor-*.md` files
- **THEN** the action falls back to single-agent mode, logs a
  `::notice::`, and invokes `claude -p` without `--agents`

### Requirement: Action pre-fetches PR context

The action SHALL pre-fetch CI check results, existing reviews,
inline comments, and linked issues using `gh` commands. Claude
SHALL read these as JSON files via `Read` tool, not via Shell.

#### Scenario: CI checks available

- **WHEN** the PR has CI check results
- **THEN** `pr-checks.json` contains check name, state, description

#### Scenario: No linked issues

- **WHEN** the PR body contains no issue references
- **THEN** `pr-linked-issues.json` is an empty array `[]`

### Requirement: Diff content is file-based, not interpolated

The action SHALL NOT interpolate diff content into the prompt
string. The diff SHALL remain in a file that Claude reads via
its `Read` tool.

#### Scenario: Large diff

- **WHEN** the diff exceeds `max-diff-lines`
- **THEN** the diff file is truncated and a truncation note is
  included in the prompt

### Requirement: Action outputs structured JSON

The action SHALL output a JSON file with `summary` (string) and
`inline_comments` (array of objects with `path`, `line`, `body`).

#### Scenario: Structured output

- **WHEN** Claude produces valid JSON matching the schema
- **THEN** `review-mode` output is `inline` and `review-json`
  points to the validated file

#### Scenario: Unstructured output

- **WHEN** Claude produces text that is not valid JSON
- **THEN** `review-mode` output is `comment` and `review-json`
  points to the raw output file

### Requirement: Tool access is read-only

The action SHALL restrict Claude to `Read` and `Glob` tools for
subagents, and `Read`, `Glob`, and `Agent` for the parent. No
agent SHALL have `Shell`, `Write`, or `Edit` access.

### Requirement: Action does not handle authentication or posting

The action SHALL NOT perform WIF authentication, fork-safe workflow
orchestration, or PR comment posting. These are the consumer's
responsibility.
