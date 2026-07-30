## Why

The `openspec-archive-change` skill (`.opencode/skills/openspec-archive-change/SKILL.md`) has two structural gaps that allow agents to perform irreversible actions without explicit user confirmation:

1. **Gap A -- Commit guard is prose only (lines 85-89)**: The instruction "CRITICAL: Do NOT move to step 6 (archive) or step 7 (branch switch) with uncommitted changes" is prose only. No `AskUserQuestion` gate blocks progression to the archive step when changes are uncommitted. Agents with compressed session context may skip prose-only warnings (see issue #346).

2. **Gap B -- Unguarded `git checkout main` (line 119)**: Step 7 executes `git checkout main` with no `AskUserQuestion` gate. Branch switches change working directory state and are irreversible in session context.

These gaps were identified in the parent audit (issue #346) root cause analysis, which found that agents bypass prose-only guards when session context is compressed. The same Gap B pattern exists in `opsx-archive.md` (issue #356).

Fixes: #360

## What Changes

Add two `AskUserQuestion` gates to the `openspec-archive-change` skill:

1. Before step 6 (archive): An explicit confirmation gate that verifies all changes are committed and pushed before proceeding to the archive operation.
2. Before `git checkout main` in step 7: An explicit confirmation gate before the irreversible branch switch.

## Capabilities

### New Capabilities

- None (this is a hardening fix, not a feature)

### Modified Capabilities

- `openspec-archive-change skill`: Step 5 gains an `AskUserQuestion` confirmation gate between commit verification and archive execution. Step 7 gains an `AskUserQuestion` gate before `git checkout main`.

### Removed Capabilities

- None

## Impact

- **File**: `.opencode/skills/openspec-archive-change/SKILL.md`
- **Behavioral**: Agents will now require explicit user confirmation before archiving (after commit check) and before switching to main branch
- **Scope**: Single file change, no code or test changes needed
- **Risk**: Low -- adds safety gates without changing existing logic flow

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies a skill file (agent instructions), not inter-hero artifact communication. No artifacts or data exchange formats are affected.

### II. Composability First

**Assessment**: N/A

This change is internal to the meta repository's agent tooling. It does not affect hero installability or dependencies.

### III. Observable Quality

**Assessment**: N/A

This change adds user confirmation gates to an agent skill. No machine-parseable output or provenance metadata is affected.

### IV. Testability

**Assessment**: N/A

Skill files are declarative instructions, not executable code. The gates will be verified through manual testing of the archive workflow. No automated test targets apply.

### V. Security by Default

**Assessment**: PASS

This change directly strengthens security posture by converting prose-only safety warnings into structural `AskUserQuestion` gates. It prevents agents from performing irreversible operations (archiving uncommitted work, switching branches) without explicit user consent -- aligning with the principle that security is a structural property, not a review-time afterthought.
