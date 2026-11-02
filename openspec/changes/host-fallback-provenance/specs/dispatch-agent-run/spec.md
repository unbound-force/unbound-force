<!-- capability: dispatch-agent-run
     change: host-fallback-provenance
     originating_issue: 901
     Archive ordering: this change MUST archive after
     shared-agent-executor — the MODIFIED 'dispatch_agent_run
     Tool' entry below narrows that change's ADDED tier-default
     clause for source "host" per requirement name and archive
     order; the MODIFY applies against the materialized
     predecessor text. Strategy and rationale: design.md (D6). -->

## ADDED Requirements

None.

## MODIFIED Requirements

### Requirement: dispatch_agent_run Tool

Previously: the tier-default clause read "If neither `tier`
nor `model` is provided, the tool MUST default to
`"standard"` tier" — unconditional across all sources. The
replacement text below is verbatim from the
shared-agent-executor delta except that the default is scoped
to entries whose `source` is not `"host"`. Supersession
rationale: design.md (D6); archive ordering: this change MUST
archive after shared-agent-executor, per the header.

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
vice versa). If neither `tier` nor `model` is provided, the
tool MUST default to `"standard"` tier — except for a
`source: "host"` entry, which resolves the parent session's
active model per the review-dispatch capability's RD-FR-037
(Host-Source Model Resolution) and never consults the
`"standard"` profile.

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

#### Scenario: Host source bypasses the tier default

- **GIVEN** a tool call with `agent`, `prompt`, and
  `source: "host"` (neither `tier` nor `model`)
- **WHEN** the tool processes the input
- **THEN** it resolves the parent session's active model per
  RD-FR-037 and does not consult the `"standard"` profile

## REMOVED Requirements

None.
