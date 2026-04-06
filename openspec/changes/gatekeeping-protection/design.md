## Context

The codebase has gatekeeping values distributed across
multiple layers:

- **Go constants**: `MaxIterations = 3`,
  `minPatternOccurrences = 3`
- **CI config**: coverage thresholds (80%/90%), `-race`
  flag, pinned action SHAs, linter config
- **Convention packs**: `[MUST]` vs `[SHOULD]`
  classifications on 100+ rules
- **Severity definitions**: CRITICAL/HIGH/MEDIUM/LOW
  boundaries and auto-fix policy
- **Agent frontmatter**: temperature settings, tool
  access restrictions (`write: false`)
- **Constitution**: 42 MUST rules
- **Markdown commands**: review iteration limits,
  worker concurrency limits, spec review markers

No document currently says "don't change these." The
protection strategy is behavioral: add explicit
instructions to the documents agents read before
acting.

## Goals / Non-Goals

### Goals

- Make gate modification a constitution-level violation
  (CRITICAL severity, non-negotiable)
- Enumerate protected value categories in AGENTS.md so
  agents know exactly what they cannot change
- Add gate-change detection to the Guard (intent drift)
  and Adversary (security) review checklists
- Add a "stop and report" instruction to Cobalt-Crush
  so it knows what to do when it hits a gate it cannot
  meet
- Keep scaffold asset copies in sync

### Non-Goals

- No Go regression tests for specific values (future
  spec could add these for highest-risk gates)
- No `uf doctor` checks for gate drift
- No automated enforcement — behavioral rules only
- No changes to the values themselves — only protecting
  them from unauthorized modification

## Decisions

### D1: Constitution-level rule

Gatekeeping integrity is added to the constitution's
Development Workflow section (not as a new core
principle). This makes it a MUST rule governed by the
existing compliance review process. Violating it is
automatically CRITICAL severity per the existing
governance model.

**Rationale**: A gate that an agent can weaken is not a
gate. Making gate modification a constitutional
violation gives it the highest enforcement weight
available without adding programmatic checks.

### D2: Exhaustive category list in AGENTS.md

Rather than a vague "don't change important values,"
AGENTS.md explicitly lists the protected categories:

1. Coverage thresholds and CRAP scores
2. Severity definitions and auto-fix policies
3. Convention pack rule classifications (MUST/SHOULD)
4. CI flags and linter configuration
5. Agent temperature and tool-access settings
6. Constitution MUST rules
7. Review iteration limits and worker concurrency
8. Workflow gate markers

**Rationale**: Agents follow instructions literally.
A vague rule gets vague compliance. An exhaustive list
removes ambiguity about what is protected.

### D3: Guard owns intent drift, Adversary owns security

The Guard's domain is "intent drift detection" — an
agent changing a gate to make code pass is the
definition of intent drift. The Guard gets a new
checklist item in both Code Review and Spec Review
modes.

The Adversary's domain is "security and resilience" —
CI flag removal, linter disabling, and pinned SHA
replacement are security concerns. The Adversary gets a
narrower, security-focused checklist item.

This avoids overlap (per Spec 019 exclusive ownership
boundaries).

### D4: Cobalt-Crush gets a proactive constraint

The Guard and Adversary detect gate changes after the
fact (during review). Cobalt-Crush needs a proactive
instruction: "when you encounter a gate you cannot meet,
stop and report rather than modifying the gate."

This is the most impactful single addition because
Cobalt-Crush is the agent most likely to encounter
a gate during implementation.

### D5: No changes to gate values themselves

This change protects gates — it does not add, remove,
or adjust any actual gatekeeping value. The current
thresholds, severity definitions, and CI flags remain
as-is.

## Risks / Trade-offs

### Risk: Behavioral compliance is not enforcement

Agents can still modify gates if they don't follow
instructions. This is accepted — behavioral rules are
the lightest-weight protection and sufficient for the
current maturity level. Programmatic enforcement (tests,
doctor checks) can be added later for the highest-risk
gates.

### Risk: Instruction bloat

Adding text to 5 agent files and 2 governance documents
increases the context that agents must process. Kept
minimal (~5-10 lines per file) to avoid diluting
attention.

### Trade-off: Legitimate gate changes require
human-driven PRs

If a gate value genuinely needs changing (e.g., raising
a coverage threshold), the change must be made by a
human or explicitly requested by a human. Agents cannot
initiate gate changes even when it would be reasonable.
This is intentional — false negatives (missed
legitimate changes) are less harmful than false
positives (silent gate weakening).
