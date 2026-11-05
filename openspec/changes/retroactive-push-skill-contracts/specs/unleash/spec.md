<!-- capability: unleash
     change: retroactive-push-skill-contracts
     Archive ordering: this change MUST archive after
     delegation-skill-enforcement, model-matrix-delegation,
     finale-token-reduction, and port-uf-rebase-pr.
     Supersession strategy: this MODIFIED delta restates the
     "Command delegation tool" requirement (UF-FR-002) from
     model-matrix-delegation's text with the skill-resolution
     contract corrected to the shipped single-root + repo-root
     file fallback behavior (PR 948, #934). The older
     delegation-skill-enforcement delta carries a light inline
     annotation only: model-matrix-delegation already replaces
     its text wholesale at that change's archive. Design
     rationale: design.md D1 and D5 in this change. -->

Requirement numbering: this delta allocates no new requirement
ids. It restates UF-FR-002 ("Command delegation tool") with the
skill-resolution contract corrected; the UF-FR-002 allocation
made by model-matrix-delegation carries forward unchanged.
UF-FR-003 through UF-FR-018 remain reserved for future unleash
changes. UF-FR-001 (Model-selection guidance) lives in
model-matrix-delegation/specs/unleash/spec.md and is not
touched. Per OpenSpec replacement semantics a MODIFIED delta
replaces its requirement wholesale — prose AND scenarios — so
all twenty-four scenarios of the current-generation text
(model-matrix-delegation/specs/unleash/spec.md) carry forward
below, with the stale two-root collision scenario rewritten as
repo-root file fallback scenarios.

## MODIFIED Requirements

### Requirement: Command delegation tool

**ID: UF-FR-002** This requirement MUST be satisfied.

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
spawning the child session. A value without `/` MUST resolve,
in order, to a skill directory at
`.opencode/skills/<name>/SKILL.md`, then — when no such skill
directory exists — to a repo-root file named `<name>`: the
value names a file at the repository root, resolved as a
traversal-protected root-file reference, and an existing
directory is NOT a file reference (a directory named `<name>`
at the repo root does not resolve). A value containing `/`
MUST resolve as a traversal-protected repo-relative path. A
value containing whitespace MUST fail with a `SKILL_NOT_FOUND`
envelope whose message explains that commands and make targets
are not valid `skill` values (spaced file paths are not valid
skill references — accepted tradeoff). Any other unresolvable
reference MUST fail with a `SKILL_NOT_FOUND` envelope that is
non-retryable and names the searched locations. A resolved
reference MUST be injected into the composed prompt as a
tool-rendered directive instructing the child to load the
skill or run the script for the task.

`uf_delegate_command` MUST additionally accept a `model`
parameter (with optional `variant`), a `profile` parameter
naming a matrix capability, and a `pack` parameter naming a UF
convention pack. Selection parameters MUST be mutually
independent of the classification parameters; their validation
and resolution MUST follow the model-matrix requirements
(MM-FR-001 through MM-FR-008): `model` is a provider/model
identity used verbatim after non-empty shape validation;
`profile` and `pack` MUST each be at least 1 and at most 128
characters — empty-string or over-length values MUST fail
input validation at the same pre-spawn stage as skill
resolution; `profile` MUST name a profile present
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
renders both directives in that fixed order. An explicitly
supplied pack that passes validation MUST render its directive
even when a higher-precedence parameter (`model` or `profile`)
supplies the child model — the directive duty is independent of
the model-precedence outcome (MM-FR-003).

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

Previously: the requirement (as restated by
model-matrix-delegation) resolved a `skill` value without `/`
against two skill roots — `.opencode/skills/<name>/SKILL.md`,
then `.opencode/skill/<name>/SKILL.md` — with no repo-root
file fallback, and carried a "Skill-root collision resolves to
the first root" scenario exercising the two-root collision.
The legacy `.opencode/skill/` root is retired (the
`uf init` skill-directory migration moves it to
`.opencode/skills/`), and the shipped resolver (#934) probes
the single `.opencode/skills/` root, then falls back to a
repo-root file, and rejects whitespace-bearing values.

#### Scenario: Deterministic delegation carries a skill

- **GIVEN** a per-task delegation classified
  `task_classification: "deterministic"` with
  `skill: "pre-flight"` and the skill exists at
  `.opencode/skills/pre-flight/SKILL.md`
- **WHEN** `uf_delegate_command` executes
- **THEN** the composed prompt contains a tool-rendered
  directive naming the skill, the checkpoint data records the
  classification and resolved reference, and the run proceeds

#### Scenario: Deterministic delegation skips the skill

- **GIVEN** a per-task delegation classified
  `task_classification: "deterministic"` supplying neither
  `skill` nor `skill_absent_reason`
- **WHEN** `uf_delegate_command` validates the input
- **THEN** it returns a non-retryable `SKILL_REFERENCE_REQUIRED`
  envelope naming the remediation and spawns no child session

#### Scenario: Skill reference does not resolve

- **GIVEN** a delegation supplies `skill: "does-not-exist"`
- **WHEN** `uf_delegate_command` resolves the reference
- **THEN** it returns a non-retryable `SKILL_NOT_FOUND`
  envelope naming the searched locations and spawns no child
  session

#### Scenario: Repo-root file resolves as a skill reference

- **GIVEN** a delegation supplies `skill: "Makefile"` and no
  skill directory exists at `.opencode/skills/Makefile/` but a
  file named `Makefile` exists at the repository root
- **WHEN** `uf_delegate_command` resolves the reference
- **THEN** it resolves as a traversal-protected repo-root file
  reference, the prompt directive instructs the child to run
  the script at `Makefile`, and the resolved reference is
  recorded in the audit trail

#### Scenario: A directory is not a root-file reference

- **GIVEN** a delegation supplies `skill: "docs"` and
  `docs/` exists as a directory at the repository root (and no
  skill directory exists at `.opencode/skills/docs/`)
- **WHEN** `uf_delegate_command` resolves the reference
- **THEN** it returns a non-retryable `SKILL_NOT_FOUND`
  envelope naming the searched locations — an existing
  directory is not a file reference — and spawns no child
  session

#### Scenario: Whitespace-bearing skill value is rejected

- **GIVEN** a delegation supplies a `skill` value containing
  whitespace (for example `skill: "make plugin-test"`)
- **WHEN** `uf_delegate_command` resolves the reference
- **THEN** it returns a non-retryable `SKILL_NOT_FOUND`
  envelope explaining that commands and make targets are not
  valid `skill` values and spawns no child session

#### Scenario: No applicable skill, reason recorded

- **GIVEN** a deterministic delegation supplies
  `skill_absent_reason: "no repo script exists for this
  one-off rename"`
- **WHEN** `uf_delegate_command` executes
- **THEN** the run proceeds and the checkpoint data and success
  envelope record the classification and the reason

#### Scenario: Judgment work passes through

- **GIVEN** a delegation classified `task_classification:
  "judgment"` with no `skill` or `skill_absent_reason`
- **WHEN** `uf_delegate_command` validates the input
- **THEN** the run proceeds with no skill requirement

#### Scenario: Judgment-classified call supplying a skill is
rejected

- **GIVEN** a delegation classified `task_classification:
  "judgment"` supplies `skill: "pre-flight"`
- **WHEN** `uf_delegate_command` validates the input
- **THEN** it returns an input-validation failure stating
  that skill references apply to deterministic work only,
  and spawns no child session

#### Scenario: Prompt-only delegation without classification

- **GIVEN** a delegation supplies `prompt` but no
  `command_path` and no `task_classification`
- **WHEN** `uf_delegate_command` validates the input
- **THEN** it returns an input-validation failure instructing
  the caller to classify the work as deterministic or judgment
  before delegating

#### Scenario: Over-length values are rejected

- **GIVEN** a delegation supplies `skill` longer than 256
  characters or `skill_absent_reason` longer than 512 characters
- **WHEN** `uf_delegate_command` validates the input
- **THEN** it returns an input-validation failure and spawns no
  child session

#### Scenario: Empty-string values are rejected

- **GIVEN** a delegation supplies `skill: ""` or
  `skill_absent_reason: ""`
- **WHEN** `uf_delegate_command` validates the input
- **THEN** it returns an input-validation failure instructing
  the caller that absence is expressed by omitting the field or
  recording a reason, and spawns no child session

#### Scenario: Non-enum classification is rejected

- **GIVEN** a delegation supplies `task_classification` with a
  value other than `deterministic` or `judgment` (wrong case,
  an unknown string, or a non-string value)
- **WHEN** `uf_delegate_command` validates the input
- **THEN** it returns an input-validation failure naming the
  accepted values and spawns no child session

#### Scenario: Quick-tier whole-change delegation is classified

- **GIVEN** a quick-tier `/uf.unleash` run delegates the whole
  routed change through prompt-only `uf_delegate_command`
- **WHEN** the implement delegation is composed
- **THEN** it declares `task_classification` derived from the
  nature of the work -- `deterministic` with `skill` or
  `skill_absent_reason` only when a repo skill or script
  performs the change end-to-end, `judgment` otherwise (docs
  authorship and prose-finding application are judgment work)
  -- and an
  unclassified quick-tier delegation fails closed like any
  other prompt-only call

#### Scenario: Skill reference is a script path

- **GIVEN** a delegation supplies
  `skill: ".specify/scripts/bash/check-prerequisites.sh"` and
  the file exists
- **WHEN** `uf_delegate_command` resolves the reference
- **THEN** it is accepted as a traversal-protected
  repo-relative path, the prompt directive instructs the child
  to run the script, and no skill-directory lookup occurs

#### Scenario: Command-path delegation is unaffected

- **GIVEN** a Step 2 plan delegation supplies `command_path`
  and no `task_classification`, `skill`, or
  `skill_absent_reason`
- **WHEN** `uf_delegate_command` validates the input
- **THEN** the run proceeds without any classification
  requirement — command-path delegation MAY omit the
  classification parameters exactly as before this change

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
- **WHEN** a delegation supplies `pack: "review-voice"`
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

## REMOVED Requirements

None.
