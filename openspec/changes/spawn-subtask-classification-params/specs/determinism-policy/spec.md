## ADDED Requirements

### Requirement: Change record for the spawn-parameter update

This change MUST record the Swarm-path carrier update in
`CHANGELOG.md` per the AGENTS.md documentation gate, covering
the spawn call block parameterization, the scaffold mirror, the
DE-008 wording change with the pack version bump, the
determinism-first skill update, and the Guard checklist update.

#### Scenario: CHANGELOG entry present

- **GIVEN** the change lands
- **WHEN** `CHANGELOG.md` is read
- **THEN** an entry describes the Swarm-path spawn call moving
  from description-carrier prose to spawn-tool parameters across
  the template, mirror, pack, skill, and Guard layers

## MODIFIED Requirements

### Requirement: Pack MUST rule DE-008 — delegation skill enforcement

The determinism convention pack MUST define DE-008 [MUST]: work
delegated through a delegation tool and classified as
deterministic MUST carry a skill or script reference verified to
exist, or a substantive recorded absence reason (one that states
what was searched for or what was found); command, agent, and
skill instruction that delegates work MUST bind mechanical work
to repo tooling through the delegation tools' parameter
contracts (`task_classification`, `skill`, `skill_absent_reason`
on `uf_delegate_command` and on the forge subtask spawn tool),
not through prose-only skill instructions. Swarm-path parallel
delegations (`[P]` tasks spawned through the forge subtask tool)
MUST pass the task classification and, for deterministic tasks,
the verified skill or script reference or recorded absence
reason as spawn-tool parameters; the spawned subtask description
is not the contract's carrier. Both tools enforce the
requirement fail-closed (`SKILL_REFERENCE_REQUIRED`,
`SKILL_NOT_FOUND`) and record the classification, resolved
reference, and reason; that record, not prompt archaeology, is
the audit source. Delegating prose that names skills without
the contract is an interim mitigation, not the mechanism of
record. The pack's severity mapping MUST list DE-008 among the
MUST rules whose violations are CRITICAL findings.

Previously: DE-008 bound Swarm-path parallel delegations to
carry the classification and reference in the spawned subtask
description, because the forge subtask tool had no parameter
contract of its own; only `uf_delegate_command` enforced
fail-closed.

#### Scenario: Rule present with MUST marker

- **GIVEN** the determinism pack at
  `.opencode/uf/packs/determinism.md`
- **WHEN** its rules are read
- **THEN** DE-008 carries a `[MUST]` marker and names the
  parameter contracts on both delegation tools

#### Scenario: Severity mapping covers DE-008

- **GIVEN** the pack's severity-mapping paragraph
- **WHEN** the MUST rule list is enumerated
- **THEN** DE-008 appears alongside DE-001, DE-003, DE-005,
  and DE-007 as a CRITICAL-severity rule

### Requirement: Pack version bump for the delegation rule

The pack version MUST bump 1.3.0 → 1.4.0 (MINOR: DE-008's
Swarm-path clause changes the normative carrier from description
prose to tool parameters) when this change lands, and the
severity-mapping paragraph MUST be updated in the same edit if
its text changes. No rule, severity, or threshold MAY be
weakened or removed by the bump.

Previously: the pack version bumped 1.2.0 → 1.3.0 (MINOR: new
rule) when DE-008 landed.

#### Scenario: Version reads 1.4.0

- **GIVEN** the pack frontmatter
- **WHEN** the version field is read after the change lands
- **THEN** it reads `1.4.0`

### Requirement: Determinism-first skill delegation guidance

The determinism-first skill MUST instruct agents that delegated
work carries its classification on both delegation paths: when
delegating work classified as deterministic, the delegation MUST
declare `task_classification` and pass a verified skill or
script reference (or record why none exists) through the command
delegation tool's parameters, and Swarm-path `[P]` delegations
MUST pass the same fields as spawn-tool parameters. The skill's
Related Enforcement section MUST reference DE-008 and the
parameter contracts so the classification rubric and the
enforcement mechanisms point at each other.

Previously: the skill instructed that Swarm-path spawn prompts
carry the classification and reference in the subtask
description, covering the command delegation tool's parameters
only.

#### Scenario: Body guidance references the contract

- **GIVEN** `.opencode/skills/determinism-first/SKILL.md`
- **WHEN** the body is read
- **THEN** delegation guidance names `task_classification`,
  `skill`, and `skill_absent_reason` and covers the Swarm-path
  spawn parameters alongside the command delegation tool

#### Scenario: Related Enforcement lists DE-008

- **GIVEN** the skill's Related Enforcement section
- **WHEN** its entries are read
- **THEN** DE-008 and the enforcement parameter contracts of
  both delegation tools are listed alongside the existing
  Principle VI and AGENTS.md references

### Requirement: Guard reviewer enforcement of DE-008

The Guard reviewer agent's Code Review Mode checklist MUST
include a deterministic-delegation check enforcing DE-008's
rule: flag new or modified delegation instruction that binds
mechanical work to skills via prose only (instead of the
parameter contracts), and flag deterministic-classified
delegations -- `uf_delegate_command` call sites and Swarm-path
spawn calls alike -- that carry no skill reference and no
recorded absence reason, and flag vacuous recorded reasons (a
reason that does not state what was searched for or what was
found). The Swarm-path check MUST audit the spawn call's
`task_classification`, `skill`, and `skill_absent_reason`
parameters, not the spawned subtask's prompt prose, and MUST
NOT describe the spawn tool as lacking a parameter contract.
The check MUST be self-describing (it names the parameter
contracts and cites no repo-local pack rule, because the agent
definition is scaffold-embedded) and MUST appear in the
checklist the review prompt assembly reads from the agent
definition, requiring no plugin changes.

Previously: the Swarm-path check audited spawn prompts for the
classification and reference carried in the subtask
description, and described the forge subtask tool as having no
parameter contract of its own.

#### Scenario: Checklist section present

- **GIVEN** `.opencode/agents/divisor-guard.md` in Code
  Review Mode
- **WHEN** the review checklist is enumerated
- **THEN** a deterministic-delegation section exists, is
  self-describing, and names the parameter contracts on both
  delegation tools

#### Scenario: Check reaches review prompts

- **GIVEN** `build_review_prompts` assembles a Guard review
  prompt from the agent definition
- **WHEN** the prompt is built
- **THEN** the deterministic-delegation check is included with
  no plugin code changes required

#### Scenario: Prose-only skill binding is flagged

- **GIVEN** a diff adds delegation instruction that requires
  skill usage in prompt prose without the parameter contract
- **WHEN** the Guard reviews the change
- **THEN** it produces a finding naming the
  deterministic-delegation violation

#### Scenario: Swarm-path spawn call omitting parameters is
flagged

- **GIVEN** a diff's Swarm-path spawn call for a deterministic
  `[P]` task passes no `task_classification` or passes neither
  `skill` nor `skill_absent_reason`
- **WHEN** the Guard reviews the change
- **THEN** it produces a finding naming the omitted parameter,
  not the prompt prose

#### Scenario: Vacuous absence reason is flagged

- **GIVEN** a deterministic delegation records
  `skill_absent_reason: "n/a"` in its checkpoint data
- **WHEN** the Guard audits the delegation
- **THEN** it produces a finding naming the vacuous absence
  reason as a deterministic-delegation violation

#### Scenario: Substantive absence reason passes

- **GIVEN** a deterministic delegation records
  `skill_absent_reason: "no repo skill or script ships
  tooling for this procedure; searched .opencode/skills/,
  .opencode/skill/, and .specify/scripts/bash/"`
- **WHEN** the Guard audits the delegation
- **THEN** no deterministic-delegation finding is raised for
  the recorded reason
  (The illustrative example's `.opencode/skill/` root is
  retired — superseded by `retroactive-push-skill-contracts`,
  openspec/changes/retroactive-push-skill-contracts/: the
  resolver searches `.opencode/skills/<name>/SKILL.md` and
  repo-root files only; the legacy root is migrated away by
  `uf init`. The scenario's intent — a substantive absence
  reason passes — is unchanged. That change MUST archive after
  this change.)
