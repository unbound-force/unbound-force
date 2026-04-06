## Why

AI agents implementing tasks can "move the goalposts"
— lowering a coverage threshold, weakening a MUST to
SHOULD, removing a CI flag, or inflating an agent's
temperature — to make their implementation pass rather
than fixing the actual issue. No current instruction
explicitly forbids this.

An audit of the codebase found 20+ unprotected
gatekeeping values across the constitution, convention
packs, CI config, agent frontmatter, and Markdown
commands. None have explicit "do not modify" protections
in the documents agents read before acting.

This change adds behavioral rules — explicit "do not
modify gatekeeping values" instructions — to the
documents that govern agent behavior: the constitution,
AGENTS.md, and the three agents most involved in
gate enforcement (Guard, Adversary) and gate encounter
(Cobalt-Crush).

## What Changes

### New Capabilities

- `Gatekeeping Integrity rule`: A new constitutional
  rule in the Development Workflow section that makes
  modifying a gatekeeping value a constitution-level
  violation (CRITICAL severity).
- `Gatekeeping Value Protection section`: A new
  subsection in AGENTS.md Behavioral Constraints that
  enumerates the protected value categories and provides
  a "what to do instead" instruction.
- `Guard gatekeeping check`: A new checklist item in
  divisor-guard.md for both Code Review and Spec Review
  modes that detects gate modifications.
- `Adversary gate tampering check`: A new checklist
  item in divisor-adversary.md Security Checks that
  detects CI/security gate weakening.
- `Cobalt-Crush gatekeeping constraint`: A new
  behavioral constraint in cobalt-crush-dev.md that
  instructs the developer agent to stop and report
  when it cannot meet a gate, rather than modifying it.

### Modified Capabilities

- `constitution.md`: Extended Development Workflow with
  Gatekeeping Integrity rule.
- `AGENTS.md`: Extended Behavioral Constraints with
  protected value categories.
- `divisor-guard.md`: Extended audit checklists in both
  review modes.
- `divisor-adversary.md`: Extended Security Checks
  section.
- `cobalt-crush-dev.md`: Extended behavioral constraints.

### Removed Capabilities

None.

## Impact

- 3 agent files modified + 3 scaffold asset copies
  synchronized
- 1 constitution file modified
- 1 AGENTS.md modified
- No Go code changes
- No test changes (behavioral rules only)
- Total: ~50 lines of Markdown additions across 8 files

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

Gatekeeping protection strengthens artifact integrity.
Agents continue to collaborate through artifacts — they
are now explicitly prevented from modifying the quality
standards those artifacts must meet.

### II. Composability First

**Assessment**: N/A

This change does not affect hero installation or
standalone functionality. It adds behavioral rules
to shared documents that all heroes read.

### III. Observable Quality

**Assessment**: PASS

By protecting severity definitions and quality
thresholds, this change ensures that Observable Quality
metrics remain meaningful. A CRAP score threshold of 30
only matters if agents cannot change it to 100.

### IV. Testability

**Assessment**: N/A

This change adds Markdown instructions, not code.
Testability is not directly affected. The behavioral
rules themselves are enforced through agent compliance,
not automated tests (a future spec could add regression
tests for the highest-risk gates).
