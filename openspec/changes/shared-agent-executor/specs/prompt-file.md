## ADDED Requirements

### Requirement: promptFile Parameter on invoke_agent

The `invoke_agent` tool MUST accept an optional
`promptFile` parameter — a string of at most 1024
characters specifying a file path containing the prompt
text.

Exactly one of `prompt` or `promptFile` MUST be provided.
If both are present, the tool MUST reject the input with
a validation error. If neither is present, the tool MUST
reject the input with a validation error.

#### Scenario: promptFile reads prompt from disk

- **GIVEN** a file at `/tmp/review-prompt.md` containing
  94KB of prompt text
- **WHEN** `invoke_agent` is called with
  `{ agent: "divisor-architect", promptFile: "/tmp/review-prompt.md" }`
- **THEN** the tool MUST read the file content via
  `deps.readText()` and deliver it as the prompt to the
  child session

#### Scenario: Mutual exclusivity enforced

- **GIVEN** a tool call with both `prompt: "hello"` and
  `promptFile: "/tmp/p.md"`
- **WHEN** input validation runs
- **THEN** the tool MUST reject with a validation error
  stating that `prompt` and `promptFile` are mutually
  exclusive

#### Scenario: Neither prompt nor promptFile provided

- **GIVEN** a tool call with only `agent: "divisor-guard"`
- **WHEN** input validation runs
- **THEN** the tool MUST reject with a validation error
  stating that one of `prompt` or `promptFile` is required

#### Scenario: promptFile read failure

- **GIVEN** `promptFile: "/nonexistent/path.md"`
- **WHEN** the tool attempts to read the file
- **THEN** the tool MUST return a failed result with a
  sanitized error message describing the read failure

#### Scenario: promptFile path length limit

- **GIVEN** `promptFile` with a string longer than 1024
  characters
- **WHEN** input validation runs
- **THEN** the tool MUST reject with a validation error

## MODIFIED Requirements

### Requirement: invoke_agent Input Schema

Previously: The `prompt` parameter was required (string,
max 128 KiB UTF-8).

Now: The `prompt` parameter is optional. Exactly one of
`prompt` or `promptFile` MUST be provided, enforced by
a `superRefine` validation.

The `prompt` max size (128 KiB) and `promptFile` max path
length (1024 chars) MUST be validated independently.

## REMOVED Requirements

(None)
