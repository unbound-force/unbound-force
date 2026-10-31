<!-- capability: unleash
     change: model-matrix-delegation
     originating_issue: 887 -->

Requirement numbering: this delta modifies the existing "Command
delegation tool" requirement (defined in
delegation-skill-enforcement and its predecessors) and adds one
requirement. The model-matrix selection semantics the tool
consumes live in specs/model-matrix/spec.md (MM-FR-001 through
MM-FR-008).

## MODIFIED Requirements

### Requirement: Command delegation tool

`uf_delegate_command` MUST accept a `task_classification`
parameter with the values `deterministic` and `judgment`,
a `skill` parameter naming a repo skill or script, and a
`skill_absent_reason` parameter recording why deterministic
work has no applicable skill or script.

Prompt-only delegation (no `command_path`) MUST declare
`task_classification`; command-path delegation MAY omit it.
When `task_classification` is `deterministic`, the call MUST
supply `skill` or `skill_absent_reason`; a call supplying
neither MUST fail closed with a `SKILL_REFERENCE_REQUIRED`
envelope that is non-retryable and names the remediation (pass
a skill reference, record the absence reason, or classify the
work as judgment). `skill` and `skill_absent_reason` MUST be
mutually exclusive, and each MUST be rejected on calls not
classified `deterministic` -- the skill directive asserts
deterministic work, so a judgment-classified or unclassified
command-path call MUST NOT supply either field. `skill`
MUST be at least 1 and at most 256 characters and
`skill_absent_reason` at least 1 and at most 512 characters;
over-length or empty-string values MUST fail input validation
(absence is expressed by omitting the field or recording a
reason, never by an empty string). A `task_classification`
value other than `deterministic` or `judgment` MUST fail
input validation.

When `skill` is supplied, the tool MUST resolve it before
spawning the child session: a value without `/` MUST resolve
to a skill directory under the repo's skill roots
(`.opencode/skills/<name>/SKILL.md`, then
`.opencode/skill/<name>/SKILL.md`); a value containing `/`
MUST resolve as a traversal-protected repo-relative path. An
unresolvable reference MUST fail with a `SKILL_NOT_FOUND`
envelope that is non-retryable and names the searched
locations. A resolved reference MUST be injected into the
composed prompt as a tool-rendered directive instructing the
child to load the skill or run the script for the task.

`uf_delegate_command` MUST additionally accept a `model`
parameter (with optional `variant`), a `profile` parameter
naming a matrix capability, and a `pack` parameter naming a UF
convention pack. Selection parameters MUST be mutually
independent of the classification parameters; their validation
and resolution MUST follow the model-matrix requirements
(MM-FR-001 through MM-FR-008): `model` is a provider/model
identity used verbatim; `profile` MUST name a profile present
in the loaded matrix (or fail closed when no matrix is
configured); `pack` MUST name a pack whose file exists under
`.opencode/uf/packs/` AND whose name is a `packs` key in the
loaded matrix — an invalid pack MUST fail closed with a
non-retryable error naming the remediation before any child
session is spawned, at the same pre-spawn validation stage as
skill resolution. The child model MUST then be resolved
through the precedence chain (MM-FR-003), evaluated before
host introspection is attempted. When the chain resolves a
matrix profile model, host introspection MUST be skipped
entirely and a delegation that would previously have failed
with `HOST_MODEL_UNAVAILABLE` MUST proceed.

Prompt composition remains deterministic with one added
section: command text (when present), then feature-dir
context, then the tool-rendered skill directive (when a skill
reference resolved), then the tool-rendered pack directive
(when a pack was explicitly supplied or classifier-selected),
then extra instructions. The pack directive MUST be
tool-rendered from a fixed template instructing the child to
load `.opencode/uf/packs/<pack>.md` before writing — the pack
content MUST NOT be inlined and the directive MUST NOT be
agent-reconstructed; its exact shape is golden-tested. A call
carrying both a resolved skill reference and a selected pack
renders both directives in that fixed order.

The tool MUST record `task_classification`, the resolved skill
reference, and any `skill_absent_reason` in the checkpoint data
and the structured success envelope, so delegations are
auditable without reconstructing prompts. Calls classified
`judgment` (and command-path calls without a classification)
MUST NOT carry a skill reference or absence reason.
`skill_absent_reason` values are recorded verbatim; the
checkpoint store's JSON serialization is the sanitization
boundary, and no additional content filtering is applied. A
recorded reason is vacuous when it does not state what was
searched for or what was found (`n/a` and `not needed` are
vacuous; a reason naming the searched locations and the
absence found is substantive). Vacuous reasons are flaggable
review findings against the deterministic-delegation rule.

The tool MUST additionally record the selection audit trail
(MM-FR-007) in the checkpoint data and the structured success
envelope: `selection_source`, the selected pack, the resolved
capability, the resolved model identity, and the classifier
confidence when a classifier participated. Calls omitting
every selection parameter on a repository without a matrix
MUST behave identically to the pre-change tool (host model,
`selection_source: "host"`) and their envelopes MUST remain
backward-compatible extensions of the pre-change shape.

The tool resolves the child model from the parent session's
current assistant message when the precedence chain falls
through to the host. When that message carries no valid
`variant` field (a provider that does not populate variants),
resolution MUST fall back to spawning the child without a
variant override rather than failing. When host-model
resolution fails for a structural reason (no resolvable model
or variant), the resulting error envelope MUST be classified
non-retryable and MUST name the remediation; only transient
causes MAY be classified retryable.

Previously: the tool accepted `agent`, `command_path`,
`prompt`, `feature_dir`, `expected_artifact`, `checkpoint_step`,
`phase`, `timeout`, `task_classification`, `skill`, and
`skill_absent_reason`, resolved the child model exclusively
from the parent session's current assistant message, composed
the prompt as command text, feature-dir context, skill
directive, then extra prompt, and recorded classification and
skill audit fields — with no model, profile, or pack
parameters, no model matrix, no classifier, and no selection
provenance.

#### Scenario: Explicit model parameter is used verbatim

- **GIVEN** any repository state
- **WHEN** a delegation supplies
  `model: provider/explicit-model`
- **THEN** the child session runs on that model,
  `selection_source` records `explicit-model`, and host
  introspection is not attempted

#### Scenario: Pack parameter selects model and renders directive

- **GIVEN** a loaded matrix whose `packs` maps `go` to a
  `code-writing` profile carrying a model via the override
- **WHEN** a delegation supplies `pack: go`
- **THEN** the child runs on the `code-writing` profile model
- **AND** the composed prompt contains the tool-rendered pack
  directive instructing the child to load
  `.opencode/uf/packs/go.md`, positioned after the feature-dir
  context and after any skill directive, before the extra
  prompt
- **AND** `selection_source` records `explicit-pack`

#### Scenario: Pack parameter naming a missing pack file fails

- **GIVEN** a loaded matrix whose `packs` maps a name whose
  file does not exist under `.opencode/uf/packs/`
- **WHEN** a delegation supplies that pack
- **THEN** matrix validation fails closed (MM-FR-001)

#### Scenario: Pack parameter without a matrix entry fails

- **GIVEN** a loaded matrix and a pack file that exists at
  `.opencode/uf/packs/review-voice.md` but has no `packs`
  entry in the matrix
- **WHEN** a delegation supplies `pack: review-voice`
- **THEN** the call fails with a non-retryable error naming
  the missing matrix entry and no child session is spawned

#### Scenario: Skill and pack directives coexist in fixed order

- **GIVEN** a deterministic-classified delegation supplying
  both `skill: "pre-flight"` and `pack: "go"`
- **WHEN** the prompt is composed
- **THEN** the skill directive appears before the pack
  directive, both before the extra prompt

#### Scenario: Selection provenance is auditable

- **GIVEN** a delegation whose model was classifier-selected
- **WHEN** the run succeeds
- **THEN** the success envelope and checkpoint data record
  `selection_source: "classifier-agent"`, the selected pack,
  the resolved capability, the resolved model, and the
  classifier confidence (MM-FR-007)

#### Scenario: Zero-config call shape is unchanged

- **GIVEN** a repository without `.uf/model-matrix.yaml`
- **WHEN** a delegation supplies none of `model`, `profile`,
  `pack`
- **THEN** the run proceeds on the host model with
  `selection_source: "host"` and the pre-change envelope
  shape extended only additively

#### Scenario: Matrix model resolves without host introspection

- **GIVEN** a loaded matrix where the precedence chain
  resolves a profile model
- **AND** parent-session introspection would fail structurally
- **WHEN** the delegation executes
- **THEN** the run proceeds on the matrix model and no
  `HOST_MODEL_UNAVAILABLE` envelope is returned

#### Scenario: Provider populates no variant field

- **GIVEN** the parent session's current assistant message
  carries model identity but no `variant` field
- **WHEN** the precedence chain falls through to the host model
- **THEN** the child session spawns without a variant override
  and the delegation proceeds

#### Scenario: Structural resolution failure

- **GIVEN** the precedence chain falls through to the host
  model and resolution fails because no model can be resolved
  from the parent session at all
- **WHEN** the error envelope is returned
- **THEN** it is classified non-retryable and includes
  the remediation (configure or pass an explicit model)

## ADDED Requirements

### Requirement: Model-selection guidance for implement delegation

**ID: UF-FR-001** This requirement MUST be satisfied.

The `/uf.unleash` command guidance for implement-phase
delegation MUST recommend passing `pack` on
`uf_delegate_command` calls when the task clearly names a
convention pack (a Go task passes `pack: "go"`, a TypeScript
task `pack: "typescript"`, spec-authoring work `pack:
"meta-prompt"`); the classifier covers the remaining calls.
The guidance MUST document the selection parameter contract
(`model`, `profile`, `pack`) and the precedence order, and MUST
NOT instruct the orchestrator to reconstruct or inline pack
content — the pack directive is tool-rendered. The guidance is
documentation-only: the tool enforces the contract; the
classifier covers unmarked tasks.

#### Scenario: Implement delegation names its pack

- **GIVEN** an implement-phase task that clearly describes Go
  implementation work and the repo ships the `go` pack
- **WHEN** `/uf.unleash` composes the delegation
- **THEN** it passes `pack: "go"` per the command guidance

#### Scenario: Unmarked task relies on the classifier

- **GIVEN** an implement-phase task with no clearly named
  convention pack
- **WHEN** `/uf.unleash` composes the delegation
- **THEN** it omits `pack` and the classifier cascade
  (MM-FR-005) derives the selection

## REMOVED Requirements

None.
