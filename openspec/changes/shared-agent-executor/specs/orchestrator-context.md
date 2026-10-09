## ADDED Requirements

### Requirement: Agent Definition Reinforcement in Child Prompts

Every orchestrator command that invokes Divisor agents via the
review-dispatch pipeline MUST instruct each child agent to
read its own agent definition file at
`.opencode/agents/{agent}.md` as Step 0 before conducting its
review, triage, or feedback assessment.

The instruction MUST appear in the mandatory child prompt
ingredients, ensuring the child agent actively processes:

- **Step 0: Prior Learnings** — Dewey knowledge graph queries
  defined in the agent definition
- **Source Documents** — mandatory reading list including
  `severity.md` per Spec 019 FR-006
- **Convention Pack markers** — `[PACK]` sections with
  graceful degradation
- **Domain-specific checks** — agent-specific verification
  steps (e.g., skill env-var reachability, repo detection)

The orchestrator MUST NOT inline agent definition content
into the child prompt. The child agent MUST read the file
itself using its available file-reading tools.

#### Scenario: Child agent reads its own definition

- **GIVEN** an orchestrator invokes `divisor-adversary` via
  `invoke_agent`
- **WHEN** the child prompt is constructed
- **THEN** the prompt MUST include an instruction to read
  `.opencode/agents/divisor-adversary.md` as Step 0 before
  conducting the review

#### Scenario: Agent definition sections are executed

- **GIVEN** a child agent has read its definition file
- **WHEN** the definition contains a Step 0: Prior Learnings
  section with Dewey query instructions
- **THEN** the child agent SHOULD execute those queries
  before beginning its review analysis

#### Scenario: Source Documents are loaded

- **GIVEN** a child agent has read its definition file
- **WHEN** the definition contains a Source Documents section
  listing `.opencode/uf/packs/severity.md`
- **THEN** the child agent SHOULD load severity.md per
  Spec 019 FR-006, satisfying the existing mandatory prompt
  ingredient for severity

### Requirement: All Review-Dispatch Orchestrators Updated

Every command that uses `plan_review_dispatch` and invokes
child agents via `invoke_agent` MUST include the agent
definition reinforcement instruction. The affected commands
are:

- `uf.review-council` — code and spec review council
- `uf.review-pr` — single-agent PR review
- `uf.triage-issue` — issue triage
- `uf.address-feedback` — reviewer feedback assessment

#### Scenario: Consistent reinforcement across commands

- **GIVEN** four orchestrator commands invoke Divisor agents
- **WHEN** each command constructs child prompts
- **THEN** all four MUST include the agent definition
  reinforcement instruction with equivalent wording

## MODIFIED Requirements

(None — this adds to existing child prompt ingredients)

## REMOVED Requirements

(None)
