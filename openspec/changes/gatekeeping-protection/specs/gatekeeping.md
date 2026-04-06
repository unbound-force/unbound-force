## ADDED Requirements

### Requirement: Constitutional Gatekeeping Integrity Rule

The constitution's Development Workflow section MUST
include a Gatekeeping Integrity rule stating that agents
MUST NOT modify values serving as quality or governance
gates. Modification of a gatekeeping value by an agent
without explicit human authorization MUST be treated as
a CRITICAL-severity constitution violation.

#### Scenario: Agent encounters a gate it cannot meet

- **GIVEN** an agent is implementing a task
- **WHEN** the implementation cannot meet a gatekeeping
  value (e.g., coverage threshold, CRAP score)
- **THEN** the agent MUST stop and report the conflict
  rather than modifying the gate

#### Scenario: Agent attempts to weaken a gate

- **GIVEN** a reviewer (Guard or Adversary) is reviewing
  a change
- **WHEN** the change modifies a gatekeeping value
  (threshold, severity, MUST→SHOULD, CI flag)
- **THEN** the reviewer MUST flag it as a finding and
  verify explicit human authorization exists

---

### Requirement: AGENTS.md Gatekeeping Value Protection

AGENTS.md Behavioral Constraints section MUST include a
Gatekeeping Value Protection subsection that enumerates
the categories of protected values:

1. Coverage thresholds and CRAP scores
2. Severity definitions and auto-fix policies
3. Convention pack rule classifications (MUST/SHOULD)
4. CI flags and linter configuration
5. Agent temperature and tool-access settings
6. Constitution MUST rules
7. Review iteration limits and worker concurrency
8. Workflow gate markers

The subsection MUST include a "what to do instead"
instruction directing agents to report the conflict
and stop.

#### Scenario: Agent reads AGENTS.md before implementing

- **GIVEN** an agent loads AGENTS.md for project context
- **WHEN** it encounters a gatekeeping value during
  implementation
- **THEN** it knows from the Gatekeeping Value Protection
  section that modifying the value is prohibited and what
  to do instead

---

### Requirement: Guard Gatekeeping Checklist Item

The divisor-guard.md agent MUST include a Gatekeeping
Integrity checklist item in both Code Review Mode and
Spec Review Mode audit sections. The item MUST check
whether the change modifies any gatekeeping value and
whether explicit human authorization exists.

#### Scenario: Guard detects threshold modification

- **GIVEN** a change lowers a coverage threshold from
  80% to 60% in CI config
- **WHEN** the Guard reviews the change
- **THEN** the Guard flags it as a finding with severity
  determined by the impact of the weakened gate

#### Scenario: Human-authorized gate change passes

- **GIVEN** a change modifies a gate value with
  documented human authorization (e.g., PR description
  says "intentionally lowering threshold per team
  decision")
- **WHEN** the Guard reviews the change
- **THEN** the Guard notes the authorization and does
  not flag it as a violation

---

### Requirement: Adversary Gate Tampering Check

The divisor-adversary.md agent MUST include a Gate
Tampering checklist item in the Security Checks section
that detects removal or weakening of CI security
controls: `-race` flag, `govulncheck`, linter rules,
pinned action SHAs, and coverage thresholds.

#### Scenario: Adversary detects CI flag removal

- **GIVEN** a change removes the `-race` flag from CI
  test configuration
- **WHEN** the Adversary reviews the change
- **THEN** the Adversary flags it as HIGH severity
  (security-relevant gate weakened)

---

### Requirement: Cobalt-Crush Gatekeeping Constraint

The cobalt-crush-dev.md agent MUST include a Gatekeeping
Integrity behavioral constraint instructing it to stop
and report when an implementation cannot meet a quality
gate, rather than modifying the gate.

#### Scenario: Developer agent hits CRAP threshold

- **GIVEN** Cobalt-Crush is implementing a function
- **WHEN** the function's complexity exceeds the CRAP
  score threshold
- **THEN** Cobalt-Crush refactors the function or
  reports the conflict — it MUST NOT change the
  threshold

## MODIFIED Requirements

### Requirement: Scaffold Asset Synchronization

All modified agent files (`divisor-guard.md`,
`divisor-adversary.md`, `cobalt-crush-dev.md`) MUST
have their scaffold asset copies synchronized after
editing.

Previously: Scaffold assets synced after agent file
changes (no change to the mechanism — just noting that
this change triggers syncs for 3 files).

## REMOVED Requirements

None.
