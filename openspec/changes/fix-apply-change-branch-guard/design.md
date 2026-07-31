## Context

The `openspec-apply-change` skill file places a CRITICAL branch-
safety rule ("NEVER switch branches or suggest archiving with
uncommitted changes") at line 212, inside the Guardrails section
at the end of a 219-line file. Agents processing the file
sequentially encounter Step 8's completion flow (which suggests
archiving) at lines 139-149 before reaching the constraint that
prohibits it with uncommitted changes.

This is a documented T2 weakness pattern: a MANDATORY rule placed
after the workflow it governs.

The proposal's constitution alignment confirmed this change is N/A
for all principles -- it modifies agent instructions only, with no
impact on artifact formats, composability, or testability.

## Goals / Non-Goals

### Goals
- Ensure agents encounter the branch-safety constraint BEFORE any
  workflow step that could trigger a violation
- Add a pre-condition block immediately after the "Steps" heading
  and before Step 1
- Retain the existing guardrail text for completeness (belt and
  suspenders)

### Non-Goals
- Rewriting or restructuring the entire SKILL.md file
- Addressing the same T2 pattern in other skill files (#355, #361)
- Adding runtime enforcement or tooling changes
- Modifying any Go source code

## Decisions

### D1: Pre-condition block placement

Insert a `**Pre-condition**` block between the `**Steps**` heading
(line 16) and Step 1 (line 18). This ensures the constraint is the
first thing an agent reads when entering the workflow.

**Rationale**: The issue specifies this exact placement. A pre-
condition is semantically correct -- it describes something that
MUST be verified before any step executes, not a step itself.

### D2: Keep existing Guardrails entry

The existing guardrail at line 212 remains unchanged. Having the
rule in both places (pre-condition and guardrails) provides
defense-in-depth: agents that skip to the guardrails section
still encounter it.

**Rationale**: Removing the guardrail entry would weaken coverage
for agents that process guardrails first (e.g., during planning
or review rather than execution).

### D3: Pre-condition text includes actionable verification

The pre-condition text includes the `git status --short` command
as the verification mechanism, matching the issue's suggested fix.

**Rationale**: A constraint without a verification method is
unenforceable. Including the command makes the rule actionable.

## Risks / Trade-offs

### Risk: Duplication of the rule in two locations

The same rule appears in both the pre-condition and guardrails.
Future edits might update one but not the other.

**Mitigation**: The pre-condition is short (2 lines) and
references the same concept. The guardrails entry is also brief.
Drift risk is low for such a focused rule.

### Trade-off: Minimal change vs. full restructure

A more thorough fix would audit the entire file for other T2
patterns. This change scopes narrowly to the reported issue.

**Accepted**: The sibling issues (#355, #361) track the same
pattern in other files. Each fix is intentionally scoped to
one file per issue.
