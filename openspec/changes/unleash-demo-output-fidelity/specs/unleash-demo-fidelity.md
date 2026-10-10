## ADDED Requirements

### FR-001: JIT re-read instruction in Step 10

The `/uf.unleash` Step 10 Demo section MUST begin with an
instruction requiring the agent to re-read the Step 10
section of the template file before composing any demo
output. The instruction MUST explicitly state that the
template file is the sole authority for the prescribed
output format after compression.

#### Scenario: JIT re-read instruction present in template

- **GIVEN** the `/uf.unleash` template
- **WHEN** the Step 10 Demo section is read
- **THEN** the first instruction MUST be a re-read
  directive telling the agent to re-read the Step 10
  Demo section from this template file before composing
  any output

#### Scenario: Demo Next Steps matches template

- **GIVEN** the agent has re-read the Step 10 section
- **WHEN** the agent composes the "Next Steps" output
- **THEN** the output MUST contain exactly two options:
  `/uf.finale` and a WORKFLOW_TIER-specific refinement
  option (`/speckit.clarify` for speckit; artifact
  update path for openspec)
- **AND** the output MUST NOT contain `/uf.review-council`,
  manual git commands, or any other improvised steps

### FR-002: Step 8 equivalence note

The Step 10 Demo section MUST include a note stating that
the pre-PR `/uf.review-council` requirement (AGENTS.md) is
already satisfied by Step 8 (Code Review). The note MUST
explicitly prohibit suggesting `/uf.review-council` as a
next step in the demo output.

#### Scenario: Equivalence note present in template

- **GIVEN** the `/uf.unleash` template Step 10 section
- **WHEN** the Next Steps prose is read
- **THEN** it MUST contain a note stating that Step 8
  (Code Review) already satisfies the pre-PR
  `/uf.review-council` requirement
- **AND** the format block MUST NOT contain
  `/uf.review-council`

#### Scenario: Agent does not re-suggest review council

- **GIVEN** the agent has reached Step 10 Demo (Steps
  1-9 complete)
- **WHEN** the agent composes Step 10 Next Steps
- **THEN** the output MUST NOT include `/uf.review-council`
  as a suggested next step
- **AND** the output MUST NOT include hand-rolled git
  commands (commit, push, PR creation)

### FR-003: Anti-improvisation guardrail

The Guardrails section of `/uf.unleash` MUST include a
rule prohibiting improvised Demo exit text. The guardrail
MUST reference the Step 10 format block as the
authoritative source for the "Next Steps" output.

#### Scenario: Guardrail present in template

- **GIVEN** the `/uf.unleash` template
- **WHEN** the Guardrails section is read
- **THEN** it MUST contain a rule starting with
  "NEVER improvise Demo exit text"

### FR-004: Command Template Fidelity rule

The `always-on-guidance/SKILL.md` MUST include a
"Command Template Fidelity" section with rules requiring
agents to re-read prescribed output sections from command
templates before emitting terminal output. The rule MUST
explicitly state that after context compression, the
template file is the sole authority for prescribed output.

#### Scenario: Fidelity rule present in guidance

- **GIVEN** the `always-on-guidance/SKILL.md` file
- **WHEN** the file contents are read
- **THEN** it MUST contain a section titled
  "Command Template Fidelity"
- **AND** the section MUST contain a rule about re-reading
  prescribed output from the template file

### FR-005: always-on-guidance scaffold propagation

The `always-on-guidance/SKILL.md` MUST be embedded in the
scaffold assets at `internal/scaffold/assets/opencode/
skills/always-on-guidance/SKILL.md`. The file MUST be
listed in `expectedAssetPaths` in `scaffold_test.go`. The
file MUST NOT appear in `knownNonEmbeddedFiles`.

#### Scenario: Drift test enforces parity

- **GIVEN** the canonical `always-on-guidance/SKILL.md` is
  modified
- **WHEN** `go test ./internal/scaffold/...` is run
- **THEN** the test MUST fail if the scaffold copy does not
  match the canonical copy

#### Scenario: uf init deploys the skill

- **GIVEN** a hero repo runs `uf init`
- **WHEN** the scaffold assets are deployed
- **THEN** the `always-on-guidance/SKILL.md` file MUST be
  present in the hero repo's `.opencode/skills/` directory

### FR-006: Reverse-drift test covers `.opencode/skills`

The `TestCanonicalSources_AreEmbedded` reverse-drift test
MUST walk `.opencode/skills/` in addition to the existing
directories (commands, agents, uf/packs). This ensures
that any canonical skill file added without a scaffold
entry or exclusion is detected.

#### Scenario: New skill file triggers reverse-drift failure

- **GIVEN** a new file exists in `.opencode/skills/`
- **AND** it is not listed in `expectedAssetPaths`
- **AND** it is not listed in `knownNonEmbeddedFiles`
- **WHEN** `go test ./internal/scaffold/...` is run
- **THEN** `TestCanonicalSources_AreEmbedded` MUST fail

## MODIFIED Requirements

### Requirement: Next Steps divergence reconciliation

The Step 10 Demo section MUST have a single canonical
representation of the "Next Steps" output. The format
block (fenced code block) is the authoritative source.
The prose description MUST reference the format block
rather than paraphrasing it independently.

Previously: Two divergent blocks described the same
output — prose (L681-685) and format block (L707-710) —
with slightly different wording.

### Requirement: AGENTS.md review-council rule

The review-council behavioral rule MUST include a
parenthetical clarifying that `/uf.unleash` Step 8
already executes the review council. The wording MUST
use "already executed as" to make clear the requirement
is met, not waived.

Previously: "MUST run `/uf.review-council` before PR
submission. Resolve all REQUEST CHANGES. No code changes
between APPROVE and PR."

New: "MUST run `/uf.review-council` before PR submission
(already executed as `/uf.unleash` Step 8; do not
re-run). Resolve all REQUEST CHANGES. No code changes
between APPROVE and PR."

### Requirement: PR Review Commands table

The "When" column for `/uf.review-council` MUST include a
note indicating it runs as part of `/uf.unleash` Step 8.

Previously: "Pre-PR (local)"
New: "Pre-PR (local; runs in `/uf.unleash` Step 8)"

## REMOVED Requirements

None.
<!-- scaffolded by uf vdev -->
