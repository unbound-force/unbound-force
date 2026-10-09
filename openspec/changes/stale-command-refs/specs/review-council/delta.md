## MODIFIED Requirements

### Requirement: Agent Discovery Step

The `/uf.review-council` command MUST discover available reviewer
agents by reading the `.opencode/agents/` directory before
delegating to any reviewers.

(Previously: `/review-council`)

#### Scenario: All five reviewers present

- **GIVEN** `.opencode/agents/` contains `reviewer-adversary.md`,
  `reviewer-architect.md`, `reviewer-guard.md`,
  `reviewer-testing.md`, and `reviewer-sre.md`
- **WHEN** the `/uf.review-council` command runs
- **THEN** all five agents are discovered and invoked in parallel

#### Scenario: Subset of reviewers present

- **GIVEN** `.opencode/agents/` contains only
  `reviewer-adversary.md`, `reviewer-architect.md`,
  `reviewer-guard.md`, and `reviewer-sre.md`
  (no `reviewer-testing.md`)
- **WHEN** the `/uf.review-council` command runs
- **THEN** only the four discovered agents are invoked
- **AND** the final report notes that `reviewer-testing` was
  absent (informational, non-blocking)

#### Scenario: No reviewer agents found

- **GIVEN** `.opencode/agents/` contains no files matching
  `reviewer-*.md`
- **WHEN** the `/uf.review-council` command runs
- **THEN** the command reports that no reviewer agents were
  found and stops without attempting delegation

### Requirement: Unknown Reviewer Agent Discovery

- **GIVEN** `.opencode/agents/` contains a file named
  `reviewer-performance.md` that is not in the known roles table
- **WHEN** the `/uf.review-council` command runs
- **THEN** `reviewer-performance` is included in the invocation
  list and invoked with a generic review delegation prompt
