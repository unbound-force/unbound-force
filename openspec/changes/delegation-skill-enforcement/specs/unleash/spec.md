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
(Superseded by `retroactive-push-skill-contracts`,
openspec/changes/retroactive-push-skill-contracts/: the skill
root set is `.opencode/skills/<name>/SKILL.md` only — the
legacy `.opencode/skill/` root is retired and migrated by
`uf init` — and a `/`-free value that names an existing
repo-root file resolves as a root-file reference; an existing
directory is not a file reference. That change MUST archive
after this change.)

The tool MUST record `task_classification`, the resolved skill
reference, and any `skill_absent_reason` in the checkpoint data
and the structured success envelope, so delegations are
auditable without reconstructing prompts. Calls classified
`judgment` (and command-path calls without a classification)
MUST NOT carry a skill reference or absence reason.
`skill_absent_reason`
values are recorded verbatim; the checkpoint store's JSON
serialization is the sanitization boundary, and no additional
content filtering is applied. A recorded reason is vacuous
when it does not state what was searched for or what was found
(`n/a` and `not needed` are vacuous; a reason naming the
searched locations and the absence found is substantive).
Vacuous reasons are flaggable review findings against the
deterministic-delegation rule.

The tool resolves the child model from the parent session's
current assistant message. When that message carries no valid
`variant` field (a provider that does not populate variants),
resolution MUST fall back to spawning the child without a
variant override rather than failing. When host-model
resolution fails for a structural reason (no resolvable model
or variant), the resulting error envelope MUST be classified
non-retryable and MUST name the remediation; only transient
causes MAY be classified retryable.

Previously: the tool read the command file, composed the
prompt, spawned the child session, verified the expected
artifact, and checkpointed -- spawning without a variant
override when the parent assistant message carried no variant
field, and classifying structural host-model resolution
failures non-retryable with remediation -- with no
classification, skill, or justification parameters; binding
delegated mechanical work to repo tooling was prose
instruction in the command template only.

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

#### Scenario: Skill-root collision resolves to the first root

- **GIVEN** the same skill name exists at both
  `.opencode/skills/<name>/SKILL.md` and
  `.opencode/skill/<name>/SKILL.md`
- **WHEN** `uf_delegate_command` resolves the reference
- **THEN** the `.opencode/skills/` entry is used and is the
  resolved reference recorded in the audit trail

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
- **THEN** the run proceeds exactly as before this change

#### Scenario: Provider populates no variant field

- **GIVEN** the parent session's current assistant message
  carries model identity but no `variant` field
- **WHEN** `uf_delegate_command` resolves the host model
- **THEN** the child session spawns without a variant override
  and the delegation proceeds

#### Scenario: Structural resolution failure

- **GIVEN** host-model resolution fails because no model can
  be resolved from the parent session at all
- **WHEN** the error envelope is returned
- **THEN** it is classified non-retryable and includes the
  remediation (configure or pass an explicit model)

### Requirement: Mechanical task tool references

The implement phase MUST classify each delegated task as
deterministic or judgment before delegation and pass the
classification through
`uf_delegate_command`'s `task_classification` parameter.
Deterministic tasks MUST pass a verified `skill` reference or
a `skill_absent_reason`; the tool enforces this contract. The
command template's implement guidance MUST present a
self-describing classification step and the parameter contract
(not prose-only skill instructions, and no citation of
repo-local guidance artifacts -- the template is
scaffold-embedded), the code-fix delegation calls MUST declare
`task_classification` derived from the nature of the fix
(deterministic with a verified skill or script reference when
a repo tool performs the fix end-to-end; judgment for
prose-finding application), the quick-tier whole-change
delegation MUST declare the classification derived from the
nature of the work, and Swarm-path parallel delegations (`[P]`
tasks spawned through the forge subtask tool) MUST carry the
task classification and, for deterministic tasks, the verified
skill or script reference or recorded absence reason in the
spawned subtask description.

Previously: the template instructed that task delegation
prompts name the repo skill or script for mechanical
procedures and pass the reference through to the child --
enforcement was agent compliance with the prose only.

#### Scenario: Mechanical task with repo skill present

- **GIVEN** a task requires executing the repo's CI-parity
  pre-flight procedure and the repo ships the pre-flight skill
- **WHEN** the task is delegated
- **THEN** the delegation is classified deterministic, names
  the skill, and the child runs the skill's full procedure

#### Scenario: Mechanical task with no repo skill

- **GIVEN** a task describes a mechanical procedure for which
  no repo skill or script exists
- **WHEN** the task is delegated
- **THEN** the delegation is classified deterministic with a
  `skill_absent_reason`, and the recorded reason is auditable
  in the checkpoint data

#### Scenario: Swarm-path parallel task carries the reference

- **GIVEN** a deterministic `[P]` task is delegated through the
  Swarm parallel path (forge subtask spawn)
- **WHEN** the spawned subtask description is composed
- **THEN** it names the task classification and the verified
  skill or script reference (or the recorded absence reason),
  and the Guard flags a spawn prompt that omits them

#### Scenario: Judgment task delegated without skill

- **GIVEN** a task requires evaluating two design approaches
- **WHEN** the task is delegated
- **THEN** the delegation is classified judgment and proceeds
  without a skill reference and without a failure
