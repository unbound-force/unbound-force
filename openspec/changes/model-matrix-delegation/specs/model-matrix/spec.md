<!-- capability: model-matrix
     change: model-matrix-delegation
     originating_issue: 887 -->

Requirement numbering: MM-FR-001 through MM-FR-008 are allocated
for this change. MM-FR-009 through MM-FR-018 are reserved for
future model-matrix changes. The delegation-tool integration
requirements live in specs/unleash/spec.md; the scaffold asset
requirements live in specs/scaffold/spec.md.

## ADDED Requirements

### Requirement: Tracked Model Matrix Configuration

**ID: MM-FR-001** This requirement MUST be satisfied.

A tracked `.uf/model-matrix.yaml` file MUST declare the model
selection structure for delegated content-agent work. The file
MUST be a closed contract validated against the canonical JSON
Schema (MM-FR-008) with no unknown root or nested keys. The v1
surface MUST comprise:

- `version: 1`;
- `profiles` — capability-named model slots (1–32 entries). Each
  profile MAY carry `model` and `variant`; both are optional, and
  a profile without a model resolves to the host model (host
  fallback). Tracked profiles SHOULD carry no model strings —
  per-developer identities arrive via the override file
  (MM-FR-002);
- `packs` — a mapping from UF convention pack name to capability.
  Every key MUST reference a pack file that exists under
  `.opencode/uf/packs/` (a `packs` entry naming a nonexistent
  pack MUST fail validation), and every value MUST name a profile
  present in `profiles`. Cross-cutting packs (severity,
  determinism, review-voice, default) SHOULD NOT carry matrix
  entries — they are never the primary capability of a task;
- `defaults` — a closed key set mapping task kind to capability
  (`plan` and `tasks` MUST map to the spec-authoring capability;
  `implement` MAY map to a capability). Unknown task kinds MUST
  fail validation;
- `agents` — optional per-agent capability pins; every key MUST
  name a non-reviewer (non-`divisor-`) agent and every value MUST
  name a present profile;
- `classifier` — structure only, no model strings in tracked
  config: `enabled`, `agent`, `agent_timeout_ms`, `llm`
  (endpoint/model/apiKey structure — populated only via the
  override file), and `confidence_threshold`.

The starter capability set is `spec-authoring`, `code-writing`,
`content-writing`, and `mechanical`; additional capabilities MUST
be expressible by adding `profiles` and `packs` entries with no
code changes.

#### Scenario: Starter tracked matrix validates

- **GIVEN** the tracked matrix declares the four starter
  capabilities with empty profiles, a `packs` mapping whose keys
  all exist under `.opencode/uf/packs/`, `defaults` mapping
  `plan` and `tasks` to `spec-authoring` and `implement` to
  `code-writing`, and classifier structure only
- **WHEN** the matrix is loaded and validated
- **THEN** it parses successfully with no model strings required

#### Scenario: Packs entry naming a nonexistent pack fails

- **GIVEN** a matrix whose `packs` mapping contains
  `no-such-pack: code-writing` and no pack file exists at
  `.opencode/uf/packs/no-such-pack.md`
- **WHEN** the matrix is validated
- **THEN** validation fails closed naming the offending pack key

#### Scenario: Packs entry naming an absent profile fails

- **GIVEN** a matrix whose `packs` mapping contains
  `go: no-such-capability` and `profiles` declares no such
  capability
- **WHEN** the matrix is validated
- **THEN** validation fails closed naming the missing profile

#### Scenario: Unknown defaults key fails

- **GIVEN** a matrix whose `defaults` carries an unknown task
  kind
- **WHEN** the matrix is validated
- **THEN** validation fails closed naming the accepted key set

#### Scenario: New capability without code changes

- **GIVEN** a matrix that adds a `profiles` entry and `packs`
  mappings for a capability beyond the starter set
- **WHEN** the matrix is validated and a delegation selects the
  new capability
- **THEN** selection resolves through the new profile with no
  code changes

### Requirement: Gitignored Override File

**ID: MM-FR-002** This requirement MUST be satisfied.

The loader MUST support a `.uf/model-matrix.override.yaml` file
that carries per-developer model identities. The file MUST be
gitignored and MUST NOT be tracked in version control. The
override surface MUST be constrained to a fail-closed whitelist:

- `profiles.<name>.model`
- `profiles.<name>.variant`
- `classifier.llm.<key>` (any nested key under `classifier.llm`)

Any other key or nested path in the override — including
structural keys such as `packs`, `defaults`, `agents`,
`classifier.enabled`, or `classifier.agent` — MUST cause a
validation failure whose error names the whitelist, preventing
silent governance changes. Whitelisted paths perform a
nested-path merge into the base matrix: override values replace
their counterparts, absent paths fall through to the base, and an
explicit `null` model propagates as host fallback. The merged
result MUST be re-validated against the identical schema as the
base file. An empty or comments-only override file is a no-op.

#### Scenario: Override supplies a profile model

- **GIVEN** a tracked matrix with `profiles.spec-authoring: {}`
- **AND** a gitignored override with
  `profiles.spec-authoring.model: some-provider/some-model`
- **WHEN** the matrix pair is loaded
- **THEN** the merged result carries the override model for that
  profile and passes schema validation

#### Scenario: Non-whitelisted override key fails closed

- **GIVEN** a valid tracked matrix
- **AND** an override containing `defaults: {implement:
  mechanical}` (a non-whitelisted key)
- **WHEN** the override is validated
- **THEN** validation fails with an error naming the whitelist
  (`profiles.*.model`, `profiles.*.variant`,
  `classifier.llm.*`)
- **AND** no merged configuration is produced

#### Scenario: Structural classifier key outside llm fails

- **GIVEN** a valid tracked matrix
- **AND** an override containing `classifier.enabled: false`
- **WHEN** the override is validated
- **THEN** validation fails closed naming the whitelist

#### Scenario: Empty override is a no-op

- **GIVEN** a valid tracked matrix
- **AND** an override file that exists but is empty or contains
  only YAML comments
- **WHEN** the matrix pair is loaded
- **THEN** the base matrix is used without modification

#### Scenario: Explicit null model propagates host fallback

- **GIVEN** a tracked matrix with
  `profiles.code-writing.model: provider/model`
- **AND** an override with `profiles.code-writing.model: null`
- **WHEN** the matrix pair is loaded
- **THEN** the merged profile carries no model and selections
  through it resolve to the host model

#### Scenario: Override referencing an absent profile fails

- **GIVEN** a valid tracked matrix
- **AND** an override setting `profiles.nonexistent.model`
- **WHEN** the merged result is validated
- **THEN** validation fails at the `profiles.nonexistent` path

### Requirement: Model Selection Precedence

**ID: MM-FR-003** This requirement MUST be satisfied.

The delegation tool MUST resolve the child model through this
fixed precedence chain, stopping at the first step that yields a
selection:

1. an explicit `model` parameter (with optional `variant`);
2. an explicit `profile` parameter naming a capability whose
   profile carries a model;
3. an explicit `pack` parameter validated per MM-FR-004 whose
   mapped profile carries a model;
4. an `agents:` pin for the delegated agent;
5. static `defaults:` by task kind (`checkpoint_step` `plan` or
   `tasks` resolves to the spec-authoring capability;
   `implement` and unknown kinds proceed to step 6);
6. the classifier cascade (MM-FR-005) for `implement` and
   unknown task kinds;
7. the host model — the resolved profile lacks a model, or no
   matrix is present.

Steps 2–6 resolve a capability first, then its profile; a
resolved profile without a model falls through to the host model
(step 7). When any matrix profile model resolves at steps 1–6,
host introspection MUST be skipped entirely — the delegation MUST
NOT fail with `HOST_MODEL_UNAVAILABLE` because the parent session
could not be introspected. When no matrix file is present, the
chain is a no-op and the host model is used exactly as before
this change (MM-FR-004 zero-config scenario). An explicit
parameter naming an unknown profile or an invalid pack MUST fail
closed with a non-retryable error naming the remediation.

#### Scenario: Explicit model wins outright

- **GIVEN** a loaded matrix whose `agents` pins and `defaults`
  would select other capabilities
- **WHEN** a delegation supplies `model:
  provider/explicit-model`
- **THEN** the child runs on the explicit model and
  `selection_source` records `explicit-model`

#### Scenario: Explicit profile beats pack and pins

- **GIVEN** a loaded matrix with an `agents` pin for the
  delegated agent and a `packs` mapping for a supplied pack
- **WHEN** a delegation supplies `profile: code-writing`
  and `pack: go`
- **THEN** the `profile` parameter wins, the `code-writing`
  profile's model is used, and `selection_source` records
  `explicit-profile`

#### Scenario: plan and tasks resolve statically

- **GIVEN** a loaded matrix whose `defaults` map `plan` and
  `tasks` to the spec-authoring capability, whose profile carries
  a model via the override
- **WHEN** delegations carry `checkpoint_step: "plan"` and
  `checkpoint_step: "tasks"`
- **THEN** both resolve to the spec-authoring profile's model
- **AND** `selection_source` records `default-kind`
- **AND** the classifier is not invoked

#### Scenario: Agent pin wins over defaults

- **GIVEN** a loaded matrix with `agents` pinning the delegated
  agent to a capability whose profile carries a model
- **WHEN** a delegation carries `checkpoint_step: "implement"`
- **THEN** the agent pin's profile model is used and
  `selection_source` records `agent-pin`

#### Scenario: Resolved profile without a model falls to host

- **GIVEN** a loaded matrix where the selected capability's
  profile carries no model (no override supplied)
- **WHEN** the delegation resolves the capability
- **THEN** the child runs on the host model and
  `selection_source` records `host`

#### Scenario: Matrix model resolution skips host introspection

- **GIVEN** a loaded matrix where selection resolves a profile
  model
- **AND** parent-session introspection would fail (no resolvable
  host model)
- **WHEN** the delegation executes
- **THEN** the run proceeds on the matrix model without
  attempting host introspection and without a
  `HOST_MODEL_UNAVAILABLE` failure

#### Scenario: Unknown profile fails closed

- **GIVEN** a loaded matrix
- **WHEN** a delegation supplies `profile: no-such-capability`
- **THEN** the call fails with a non-retryable error naming the
  unknown profile and no child session is spawned

### Requirement: Zero-Config No-Op

**ID: MM-FR-004** This requirement MUST be satisfied.

When `.uf/model-matrix.yaml` is absent, delegation MUST behave
identically to the pre-change tool: the child model resolves from
the parent session via host introspection, no matrix-related
errors or warnings are raised, no override file is read, and the
classifier is not invoked. A present-but-empty matrix file
(comments only) MUST be treated as a valid no-op matrix: the
precedence chain falls through every step to the host model. An
explicit `model`, `profile`, or `pack` parameter on a call with
no matrix present MUST still validate: `model` proceeds (it needs
no matrix); `profile` and `pack` fail closed with a non-retryable
error stating that no matrix is configured.

#### Scenario: Missing matrix is exact today-behavior

- **GIVEN** no `.uf/model-matrix.yaml` in the repository
- **WHEN** a delegation executes without selection parameters
- **THEN** the child model resolves from the parent session
  exactly as before this change, with no matrix errors

#### Scenario: Comments-only matrix is a no-op

- **GIVEN** a `.uf/model-matrix.yaml` containing only comments
- **WHEN** a delegation executes without selection parameters
- **THEN** the host model is used with no errors

#### Scenario: Explicit profile without a matrix fails closed

- **GIVEN** no `.uf/model-matrix.yaml` in the repository
- **WHEN** a delegation supplies `profile: code-writing`
- **THEN** the call fails with a non-retryable error stating no
  matrix is configured and no child session is spawned

### Requirement: Classifier Cascade

**ID: MM-FR-005** This requirement MUST be satisfied.

For `implement` and unknown task kinds that reach step 6 of the
precedence chain, selection MUST run a classifier cascade of at
most two engines, in order: the agent engine, then the LLM
engine. The first result that is valid JSON naming a
matrix-valid pack (a `packs` key whose pack file exists) with
confidence greater than or equal to the configured
`confidence_threshold` MUST win. A classifier-selected pack
performs the same double duty as an explicit pack parameter:
its mapped profile model is used AND the pack directive is
rendered in the child prompt.

Classifier failure MUST NEVER block delegation. A timed-out,
errored, malformed, below-threshold, or matrix-invalid classifier
result MUST fall back to the static default for the task kind
(the `defaults` entry when present, else the host model), and the
audit trail MUST record the fallback. Classifier results MUST be
cached keyed by the SHA-256 of the feature dir, prompt, and agent
name, with a bounded cache size; a cache hit MUST NOT re-invoke
either engine.

The agent engine MUST be a hidden subagent
(`.opencode/agents/model-matrix-classifier.md`, mode `subagent`,
hidden, described as internal to the model-matrix selection
layer) that investigates the feature dir's spec/plan/tasks and
existing code with tools, then responds with ONLY a JSON object
`{"pack": "<name>", "confidence": <0.0-1.0>, "reasoning":
"<optional>"}` — no markdown fences, no prose, never questions.
The agent engine run MUST be bounded by the configured
`agent_timeout_ms`. The child session inherits the session model
hierarchy (no endpoint configuration).

#### Scenario: Agent engine selects an implement pack

- **GIVEN** a loaded matrix with `packs` mapping `go` to
  `code-writing`, and an `implement` delegation whose prompt
  clearly describes Go work
- **WHEN** the classifier cascade runs and the agent engine
  returns `{"pack": "go", "confidence": 0.9}`
- **THEN** the `go` profile's model is used, the pack directive
  is rendered, and `selection_source` records `classifier-agent`

#### Scenario: Below-threshold result falls back statically

- **GIVEN** a configured `confidence_threshold` of 0.6
- **WHEN** a classifier result carries confidence 0.4
- **THEN** the result is discarded, the static default for the
  task kind is used, and the audit trail records the fallback

#### Scenario: Agent engine timeout falls through to LLM engine

- **GIVEN** the agent engine exceeds `agent_timeout_ms`
- **WHEN** the cascade proceeds
- **THEN** the LLM engine is consulted next, and if it also
  fails, the static default is used — delegation is never
  blocked

#### Scenario: Classifier names a matrix-invalid pack

- **GIVEN** a classifier result naming a pack that is not a
  `packs` key in the loaded matrix
- **WHEN** the result is validated
- **THEN** the result is discarded and the cascade continues or
  falls back statically

#### Scenario: Cache hit skips the engines

- **GIVEN** a prior classification cached for the same
  feature dir, prompt, and agent
- **WHEN** a subsequent delegation presents the same inputs
- **THEN** the cached pack is used without invoking either
  classifier engine

#### Scenario: Classifier never blocks delegation

- **GIVEN** both classifier engines fail (timeout, network
  error, malformed response)
- **WHEN** the delegation executes
- **THEN** the run proceeds on the static default with
  `selection_source` recording the fallback, and no error is
  surfaced to the caller

### Requirement: Classifier LLM Engine Contract

**ID: MM-FR-006** This requirement MUST be satisfied.

The classifier LLM engine MUST issue a chat-completions POST with
`temperature: 0` and `response_format: {type: "json_object"}`,
retrying once without the `response_format` field when the
endpoint rejects it with HTTP 400. The request MUST be bounded by
a timeout enforced via AbortSignal. The engine MUST interpolate
the apiKey from an `{env:VAR}` reference (never a literal key)
and MUST reject a configured literal apiKey. All LLM endpoint
configuration (endpoint URL, model, apiKey reference) MUST live
only in the gitignored override file under `classifier.llm.*`;
the tracked matrix carries `classifier.llm` structure only. When
`classifier.llm` carries no endpoint configuration, the LLM
engine MUST be skipped (agent engine result or static fallback
applies).

#### Scenario: Env-interpolated apiKey

- **GIVEN** an override with `classifier.llm.apiKey:
  "{env:MY_PROVIDER_KEY}"` and the environment variable set
- **WHEN** the LLM engine issues its request
- **THEN** the Authorization header carries the interpolated
  value and no literal key appears in any tracked file

#### Scenario: Literal apiKey is rejected

- **GIVEN** configuration carrying a literal apiKey string
- **WHEN** the LLM engine validates its configuration
- **THEN** it fails closed with an error requiring an
  `{env:VAR}` reference

#### Scenario: response_format rejection retries once

- **GIVEN** an endpoint that rejects `response_format` with
  HTTP 400
- **WHEN** the LLM engine issues its request
- **THEN** it retries once without the field and parses the
  response

#### Scenario: Unconfigured LLM engine is skipped

- **GIVEN** a tracked matrix whose `classifier.llm` is empty and
  no override populates it
- **WHEN** the agent engine fails
- **THEN** the cascade falls back statically without attempting
  any network call

### Requirement: Selection Audit Trail

**ID: MM-FR-007** This requirement MUST be satisfied.

Every delegation MUST record its model selection in the
structured success envelope and the checkpoint data: the
`selection_source` (exactly one of `explicit-model`,
`explicit-profile`, `explicit-pack`, `agent-pin`,
`default-kind`, `classifier-agent`, `classifier-llm`, `host`),
the selected pack (when one was selected or classified), the
resolved capability, the resolved model identity, and the
classifier confidence when a classifier participated. Calls
omitting selection parameters on a matrix-less repository MUST
record `selection_source: "host"` with the host model — the
pre-change envelope shape extended, never removed. Classifier
fallbacks MUST be recorded (the attempted engine and the fallback
reason) so low-confidence or failed classifications are
auditable.

#### Scenario: Audit fields accompany every selection

- **GIVEN** any delegation executing with a matrix present
- **WHEN** the run succeeds
- **THEN** the success envelope and checkpoint data carry
  `selection_source`, the selected pack or capability, and the
  resolved model

#### Scenario: Classifier confidence is recorded

- **GIVEN** a classifier-selected pack with confidence 0.85
- **WHEN** the delegation succeeds
- **THEN** the audit fields record `selection_source:
  "classifier-agent"` and the confidence value

#### Scenario: Fallback is auditable

- **GIVEN** a classifier that failed and fell back statically
- **WHEN** the delegation succeeds
- **THEN** the audit fields record the static fallback source
  and the classifier failure reason

### Requirement: Canonical JSON Schema Registry

**ID: MM-FR-008** This requirement MUST be satisfied.

The tracked matrix contract MUST be codified in a canonical JSON
Schema registry at `schemas/model-matrix/` containing
`v1.schema.json` (closed input, `additionalProperties: false`
throughout), a `README.md` documenting the surface and the
override whitelist, and `samples/` fixtures covering valid and
invalid documents (including whitelist-violating override
samples). Following the established registry convention
(review-matrix, review-dispatch), the deployed
`.uf/schemas/model-matrix/` copy and the embedded scaffold
mirror under `internal/scaffold/assets/uf/schemas/model-matrix/`
MUST be byte-identical to the root registry (drift-test
enforced). The plugin's runtime validation MUST enforce the same
surface as the schema.

#### Scenario: Closed schema rejects unknown fields

- **GIVEN** a matrix document with an unknown root or nested
  field
- **WHEN** it is validated against `v1.schema.json`
- **THEN** validation fails naming the unknown field

#### Scenario: Invalid samples fail, valid samples pass

- **GIVEN** the registry's valid and invalid sample fixtures
- **WHEN** each is validated against `v1.schema.json`
- **THEN** every valid sample passes and every invalid sample
  fails for its documented reason

## REMOVED Requirements

None. The model-matrix capability is new; no existing requirement
is removed.
