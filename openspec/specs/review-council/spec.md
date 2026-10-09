# Review Council Command Spec

Behavioral requirements for the `/uf.review-council` command.

Synced from change: `dynamic-reviewer-discovery` (2026-03-16)

---

## Requirements

### Requirement: Agent Discovery Step

The `/uf.review-council` command MUST discover available reviewer
agents by reading the `.opencode/agents/` directory before
delegating to any reviewers.

- The command MUST list files in `.opencode/agents/` matching
  the pattern `reviewer-*.md`.
- The command MUST extract agent names by stripping the `.md`
  extension from discovered filenames.
- The command MUST only invoke agents that were discovered.
- The command MUST NOT hardcode a fixed list of agents for
  invocation.

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

### Requirement: Discovery Summary in Final Report

The review council final report MUST include a discovery summary
section in both Code Review Mode and Spec Review Mode.

- The summary MUST list all discovered reviewer agents that
  were invoked.
- The summary MUST list any known reviewer roles that were
  absent (from the reference set: adversary, architect, guard,
  testing, sre).
- Absent reviewers MUST be reported as informational notes,
  not as blocking findings.

#### Scenario: Partial council with discovery summary

- **GIVEN** three of five reviewer agents are present
- **WHEN** the review council completes its review
- **THEN** the final report includes a "Discovery Summary"
  listing the three invoked reviewers and two absent reviewers
- **AND** the absent reviewers are marked as informational

### Requirement: Known Reviewer Role Descriptions

The command MUST maintain a reference table of known reviewer
roles with their descriptions. This table serves as documentation
and context for delegation, but MUST NOT be used as the
invocation list.

- The reference table SHOULD include role name, persona, and
  focus area for each known reviewer.
- When a discovered agent matches a known role, the command
  SHOULD use the role description to provide context to the
  agent during delegation.
- When a discovered agent does not match any known role, the
  command SHOULD still invoke it with a generic review prompt.

#### Scenario: Unknown reviewer agent discovered

- **GIVEN** `.opencode/agents/` contains a file named
  `reviewer-performance.md` that is not in the known roles table
- **WHEN** the `/uf.review-council` command runs
- **THEN** `reviewer-performance` is included in the invocation
  list and invoked with a generic review delegation prompt

### Requirement: Council Verdict Policy

The council verdict MUST be based on discovered reviewers only.

The council MUST return APPROVE only when all *discovered*
reviewers return APPROVE. Any single REQUEST CHANGES from a
discovered reviewer means the council verdict is REQUEST CHANGES.
Absent reviewers MUST NOT affect the verdict.

#### Scenario: Verdict with partial council

- **GIVEN** only three reviewer agents are discovered
- **AND** all three return APPROVE
- **WHEN** the council determines its verdict
- **THEN** the verdict is APPROVE
- **AND** an informational note lists the two absent reviewers

### Requirement: Iterative Fix Loop Scope

The iterative fix loop MUST re-run only discovered reviewers.

Re-run all *discovered* reviewers to verify fixes. The iteration
limit (3) and escalation policy (ask user) remain unchanged.
