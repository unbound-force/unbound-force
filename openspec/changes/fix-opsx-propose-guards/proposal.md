## Why

Issue #353 identifies two structural gaps in the `opsx-propose`
command and its corresponding `openspec-propose` skill that can
cause agent misbehavior under context compression:

**Gap A — Dirty-tree guard is prose only**: The dirty working
tree check (Step 3a) describes what to do in narrative prose
but never specifies using `AskUserQuestion` with concrete
options. Under context compression, agents may skip the guard
entirely and silently switch branches with uncommitted work,
risking changes applied to the wrong branch.

**Gap B — STOP HERE rule placement**: The "STOP HERE. Do NOT
proceed to implementation." directive appears at line 131/160
after the full artifact-creation workflow. The rule that
governs when to stop appears after the actions it should gate.
Under context compression, agents may never reach this rule
and proceed to implementation.

Both gaps were surfaced during the root cause analysis of
issue #346 (review-pr command skipping confirmation gates
under session compression).

## What Changes

Two files are modified:

1. **`.opencode/commands/opsx-propose.md`** — The slash command
   consumed by OpenCode agents.
2. **`.opencode/skills/openspec-propose/SKILL.md`** — The skill
   file loaded when the openspec-propose skill is invoked.

### Gap A fix (both files):

Replace the prose-only dirty-tree guard with an explicit
`AskUserQuestion` call that presents concrete options:

```
Use AskUserQuestion with options:
  ["Stash changes and continue", "Abort — keep changes as-is"]
```

This makes the confirmation gate machine-enforceable rather
than relying on the agent to interpret prose instructions.

### Gap B fix (both files):

Add a bolded preamble at the top of the Steps section (before
Step 1) that states the STOP HERE rule up front:

```
**PREAMBLE — ARTIFACTS ONLY**: This command creates spec
artifacts (proposal, design, specs, tasks). It MUST NOT
implement code, commit, push, create PRs, or run /unleash,
/opsx-apply, or /cobalt-crush. After artifacts are complete,
STOP and prompt the user.
```

The existing STOP HERE block after Step 6 is retained as
reinforcement.

## Capabilities

### New Capabilities

- None (this is a fix, not a feature)

### Modified Capabilities

- `opsx-propose dirty-tree guard`: Adds explicit
  AskUserQuestion tool call with structured options instead
  of prose-only instructions
- `opsx-propose STOP HERE placement`: Adds preamble before
  the workflow steps to front-load the implementation
  prohibition

### Removed Capabilities

- None

## Impact

- **Files affected**: 2 agent instruction files (command +
  skill)
- **Behavioral change**: Agents will now present a structured
  confirmation dialog when dirty working trees are detected,
  instead of relying on prose interpretation
- **Risk**: Low — both files are agent instructions (Markdown),
  not source code; no build/test impact
- **Related**: Issue #350 tracks the same Gap A fix for the
  skill file specifically; this change addresses both files
  together

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.2.0).

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies agent instruction files, not inter-hero
artifact formats or communication protocols. No impact on
artifact-based collaboration.

### II. Composability First

**Assessment**: N/A

No dependencies are added or modified. The change is
contained within two instruction files in this repository.

### III. Observable Quality

**Assessment**: PASS

The fix makes agent behavior more deterministic by replacing
prose-only guards with structured tool calls. This improves
the observability of agent decision points — the
AskUserQuestion call creates an auditable interaction record.

### IV. Testability

**Assessment**: N/A

No source code is added or modified. The change affects
Markdown instruction files that are not subject to automated
testing. The behavioral improvement can be verified by
running `/opsx-propose` with a dirty working tree and
confirming the structured prompt appears.

### V. Security by Default

**Assessment**: PASS

The dirty-tree guard is a security-adjacent concern: silently
switching branches with uncommitted work can cause changes to
be applied to unintended branches, potentially exposing
in-progress work in the wrong context. Making the guard
machine-enforceable via AskUserQuestion strengthens the
least-privilege principle by ensuring explicit user consent
before branch operations.
