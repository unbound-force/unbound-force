## ADDED Requirements

### Requirement: dispatch_agent_run Tool

The `review-dispatch` plugin MUST register a
`dispatch_agent_run` tool that deterministically resolves
model slugs from the review matrix and executes an agent
session via `executeAgentSession()`.

The tool MUST accept the following parameters:
- `agent` (required): Agent name matching
  `/^divisor-[a-z0-9-]{1,63}$/`
- `prompt` (optional): Inline prompt text, max 128 KiB
- `promptFile` (optional): File path to prompt, max 1024
  chars
- `tier` (optional): One of `"lightweight"`,
  `"standard"`, `"heavy"` — used to resolve the model
  from the review matrix
- `model` (optional): Explicit `provider/model-id@suffix`
  — bypasses tier resolution
- `variant` (optional): Runtime variant suffix
- `read_only` (optional): Defaults to `true`
- `timeout` (optional): Max execution time in ms

Exactly one of `prompt` or `promptFile` MUST be provided.
If `tier` is provided, `model` MUST NOT be provided (and
vice versa). If neither `tier` nor `model` is provided,
the tool MUST default to `"standard"` tier.

#### Scenario: Tier-based model resolution

- **GIVEN** a review matrix at `.uf/review-matrix.yaml`
  with `profiles.standard.model:
  "google-vertex-anthropic/claude-sonnet-4-20250514@default"`
- **WHEN** `dispatch_agent_run` is called with
  `{ agent: "divisor-architect", prompt: "...", tier: "standard" }`
- **THEN** the tool MUST resolve the model to
  `"google-vertex-anthropic/claude-sonnet-4-20250514@default"`
  and pass the full slug including `@default` to
  `executeAgentSession()`

#### Scenario: Explicit model bypasses tier resolution

- **GIVEN** a tool call with `model:
  "google-vertex-anthropic/claude-opus-4-6@default"`
- **WHEN** the tool processes the input
- **THEN** the tool MUST use the provided model slug
  directly without reading the review matrix for model
  resolution

#### Scenario: Default tier when neither specified

- **GIVEN** a tool call with only `agent` and `prompt`
- **WHEN** the tool processes the input
- **THEN** the tool MUST default to tier `"standard"`
  and resolve the model from the standard profile in the
  review matrix

#### Scenario: Mutual exclusivity of tier and model

- **GIVEN** a tool call with both `tier: "heavy"` and
  `model: "some/model@v1"`
- **WHEN** input validation runs
- **THEN** the tool MUST reject with a validation error
  stating that `tier` and `model` are mutually exclusive

### Requirement: Agent Manifest Validation

The `dispatch_agent_run` tool MUST validate the `agent`
parameter against the reviewer manifest at
`.uf/reviewer-capabilities.yaml` before invoking the
session. The tool MUST NOT rely solely on upstream
validation from `plan_review_dispatch`.

#### Scenario: Unknown agent rejected

- **GIVEN** an agent name `"divisor-unknown"` not present
  in the reviewer manifest
- **WHEN** `dispatch_agent_run` is called
- **THEN** the tool MUST return a failed result with an
  error identifying the unknown agent

#### Scenario: Valid agent accepted

- **GIVEN** an agent name `"divisor-architect"` present in
  the reviewer manifest
- **WHEN** `dispatch_agent_run` is called with valid
  parameters
- **THEN** the tool MUST proceed to model resolution and
  session execution

### Requirement: Review Matrix Loading

The `dispatch_agent_run` tool MUST load the review matrix
using existing `loadPolicies()` and resolve tier-based
models using `profilePair()`.

If the review matrix file is missing or unparseable, the
tool MUST return a failed result with a descriptive error.
It MUST NOT fall back to a hardcoded model.

#### Scenario: Missing review matrix

- **GIVEN** no file at `.uf/review-matrix.yaml`
- **WHEN** `dispatch_agent_run` is called with
  `tier: "standard"`
- **THEN** the tool MUST return a failed result with an
  error indicating the review matrix could not be loaded

### Requirement: Client Capture in review-dispatch

The `review-dispatch` plugin's `server()` method MUST
capture `input.client` (the full plugin client) in
addition to the existing `readText` and `parseYaml`
captures. The captured client MUST be passed to
`executeAgentSession()` when handling `dispatch_agent_run`
calls.

#### Scenario: Plugin initialization captures client

- **GIVEN** the review-dispatch plugin is loaded by
  OpenCode
- **WHEN** `server()` is called with `input`
- **THEN** `input.client` MUST be stored and available
  to the `dispatch_agent_run` tool handler

### Requirement: Provenance in dispatch_agent_run

The `dispatch_agent_run` tool MUST return an
`InvokeAgentResult` with full provenance metadata:
`requested_model`, `resolved_parent_model`,
`reported_child_model`, `model_mismatch`, `variant`,
`read_only`, `usage`, and `error`.

The `requested_model` MUST reflect the model slug as
resolved from the review matrix (for tier-based) or as
provided (for explicit model).

#### Scenario: Provenance tracks tier-resolved model

- **GIVEN** tier `"heavy"` resolves to model
  `"google-vertex-anthropic/claude-opus-4-6@default"`
- **WHEN** the agent session completes
- **THEN** `provenance.requested_model` MUST be
  `"google-vertex-anthropic/claude-opus-4-6@default"`

## MODIFIED Requirements

(None — dispatch_agent_run is a new tool)

## REMOVED Requirements

(None)
