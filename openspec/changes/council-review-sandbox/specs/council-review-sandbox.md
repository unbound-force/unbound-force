# Council Review Sandbox Specification

## ADDED Requirements

### Requirement: Runtime permission config MUST deny dangerous tools

The `run-review.sh` script MUST inject an `OPENCODE_CONFIG_CONTENT`
env var containing a permission config that denies `edit`, `bash`,
`webfetch`, `websearch`, and `skill`. Read, glob, and grep MUST
remain allowed (OpenCode defaults).

#### Scenario: Model attempts bash command

- **GIVEN** `run-review.sh` sets `OPENCODE_CONFIG_CONTENT` with
  `"bash": "deny"`
- **WHEN** the model attempts to use the `bash` tool
- **THEN** OpenCode denies the tool call at the runtime level

#### Scenario: Model attempts web fetch

- **GIVEN** `run-review.sh` sets `OPENCODE_CONFIG_CONTENT` with
  `"webfetch": "deny"`
- **WHEN** the model attempts to use the `webfetch` tool
- **THEN** OpenCode denies the tool call at the runtime level

#### Scenario: Model reads diff file

- **GIVEN** `run-review.sh` sets `OPENCODE_CONFIG_CONTENT` with
  no denial for `read`
- **WHEN** the model uses the `read` tool on
  `pr-diff-annotated.patch`
- **THEN** the tool call succeeds and returns the file contents

### Requirement: Task tool MUST NOT be denied

The permission config MUST NOT include `"task": "deny"`. The
council review orchestrator MUST be able to invoke Divisor
subagents via the Task tool for multi-agent review.

#### Scenario: Orchestrator invokes Divisor subagent

- **GIVEN** `run-review.sh` sets `OPENCODE_CONFIG_CONTENT`
  without `"task": "deny"`
- **WHEN** the orchestrator invokes a Divisor subagent via Task
- **THEN** the subagent is available and review proceeds in
  multi-agent mode

#### Scenario: Task denied breaks multi-agent review

- **GIVEN** `OPENCODE_CONFIG_CONTENT` includes `"task": "deny"`
- **WHEN** the orchestrator attempts multi-agent review
- **THEN** no subagents are available, degrading to single-agent
  mode (THIS IS THE FAILURE CASE -- task MUST NOT be denied)

### Requirement: External plugins MUST be isolated

The `opencode run` invocation MUST include the `--pure` flag to
prevent external MCP plugins from loading during the review.

#### Scenario: Pure mode active

- **GIVEN** `run-review.sh` invokes `opencode run --pure`
- **WHEN** OpenCode initializes the review session
- **THEN** no external MCP plugins are loaded (including dewey
  and replicator defined in the project's `opencode.json`)

### Requirement: All ask-default permissions MUST be denied

The permission config MUST explicitly deny every permission that
defaults to `"ask"` (`external_directory`, `doom_loop`). CI has
no TTY for approval prompts. Blanket skip flags
(`--dangerously-skip-permissions`, `--auto`) MUST NOT be used.

#### Scenario: No TTY in CI

- **GIVEN** `OPENCODE_CONFIG_CONTENT` includes
  `"external_directory": "deny"` and `"doom_loop": "deny"`
- **WHEN** OpenCode encounters these permissions during review
- **THEN** the tool call is denied (no TTY prompt needed)

#### Scenario: No blanket skip flags

- **GIVEN** `run-review.sh` invokes `opencode run`
- **WHEN** the CLI arguments are inspected
- **THEN** neither `--dangerously-skip-permissions` nor `--auto`
  is present

### Requirement: Permission config MUST merge with Vertex config

The `OPENCODE_CONFIG_CONTENT` env var MUST contain both permission
denials and Vertex AI provider config (when using
`google-vertex-anthropic` provider). The two configs MUST be
merged into a single JSON object.

#### Scenario: Vertex AI provider with permissions

- **GIVEN** `MODEL` is `google-vertex-anthropic/claude-sonnet-4-6`
- **WHEN** `run-review.sh` builds `OPENCODE_CONFIG_CONTENT`
- **THEN** the JSON contains both `"permission"` and `"provider"`
  keys

#### Scenario: Non-Vertex provider with permissions

- **GIVEN** `MODEL` is `anthropic/claude-sonnet-4-6`
- **WHEN** `run-review.sh` builds `OPENCODE_CONFIG_CONTENT`
- **THEN** the JSON contains `"permission"` key only (no
  `"provider"` key needed)

### Requirement: Divisor agents MUST use `permission:` frontmatter

All 9 Divisor agent definitions MUST use the `permission:` syntax
(not deprecated `tools:`). Review agents MUST deny `edit`, `bash`,
and `webfetch`. Content agents MUST deny tools inappropriate for
their role.

#### Scenario: Review agent frontmatter (adversary, architect,
  guard, sre, testing)

- **GIVEN** a review-focused Divisor agent `.md` file
- **WHEN** the frontmatter is parsed
- **THEN** it contains `permission:` with `edit: deny`,
  `bash: deny`, `webfetch: deny`
- **AND** it does NOT contain a `tools:` key

#### Scenario: Content agent frontmatter (curator, envoy,
  herald, scribe)

- **GIVEN** a content-focused Divisor agent `.md` file
- **WHEN** the frontmatter is parsed
- **THEN** it contains `permission:` (not `tools:`)
- **AND** deny rules match the agent's interactive role

### Requirement: Sandboxing decision MUST be documented

The `council-review-action/docs/decisions.md` file MUST contain a
"Runtime sandbox" decision entry documenting:
- `OPENCODE_CONFIG_CONTENT` permission config mechanism
- `--pure` plugin isolation
- Why ask-default permissions are denied explicitly (no blanket skip)
- Why `task` is NOT denied
- Why `--agent plan` is rejected
- Agent frontmatter as defense-in-depth layer

### Requirement: Security risk register MUST reflect mitigations

The `council-review-action/docs/security-risks.md` MUST update
T2 (network exfiltration) and A3 (secrets in environment) to
reference `OPENCODE_CONFIG_CONTENT` permission config and `--pure`
as implemented mitigations.
