# Council Review Action Specification

## ADDED Requirements

### Requirement: Action discovers Divisor agents dynamically

The action MUST discover Divisor persona definitions by globbing
`.opencode/agents/divisor-*.md` in the checked-out repository.
OpenCode auto-discovers these agents via its native `.opencode/`
context loading.

#### Scenario: Multiple personas discovered

- **WHEN** the repo contains 5+ `divisor-*.md` files
- **THEN** the action invokes `opencode run` which auto-discovers
  the agents and applies each persona's review criteria

#### Scenario: Zero personas discovered

- **WHEN** the repo contains no `divisor-*.md` files
- **THEN** the action falls back to single-agent mode, logs a
  `::notice::`, and invokes `opencode run` without agent context

### Requirement: Action pre-fetches PR context

The action MUST pre-fetch CI check results, existing reviews,
inline comments, and linked issues using `gh` commands. OpenCode
MUST read these as JSON files via `Read` tool, not via Shell.

#### Scenario: CI checks available

- **WHEN** the PR has CI check results
- **THEN** `pr-checks.json` contains check name, state, description

#### Scenario: No linked issues

- **WHEN** the PR body contains no issue references
- **THEN** `pr-linked-issues.json` is an empty array `[]`

### Requirement: Diff content is file-based, not interpolated

The action MUST NOT interpolate diff content into the prompt
string. The diff MUST remain in a file that OpenCode reads via
its `Read` tool.

#### Scenario: Large diff (deferred)

Diff truncation via `max-diff-lines` is not currently
implemented. The noise filtering in `prepare-diff.sh`
reduces diff size sufficiently for typical PRs. Truncation
MAY be added in a future iteration if large diffs cause
token budget issues.

### Requirement: Action outputs structured JSON

The action MUST output a JSON file with `summary` (string) and
`inline_comments` (array of objects with `path`, `line`, `body`).

#### Scenario: Structured output

- **WHEN** OpenCode produces valid JSON matching the schema
- **THEN** `review-mode` output is `inline` and `review-json`
  points to the validated file

#### Scenario: Unstructured output

- **WHEN** OpenCode produces text that is not valid JSON
- **THEN** `review-mode` output is `comment` and `review-json`
  points to the raw output file

### Requirement: Tool access is read-only

The action MUST restrict OpenCode to `Read` and `Glob` tools
for subagents, and `Read`, `Glob`, and `Agent` for the parent.
No agent MUST have `Shell`, `Write`, or `Edit` access.

### Requirement: Action does not handle authentication or posting

The action MUST NOT perform WIF authentication, fork-safe workflow
orchestration, or PR comment posting. These are the consumer's
responsibility.
