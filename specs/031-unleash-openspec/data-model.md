# Data Model: Unleash OpenSpec Support

**Branch**: `031-unleash-openspec` | **Date**: 2026-04-16

## Change Map

This feature modifies 2 files (1 live + 1 scaffold
copy). No new files are created. No Go code is changed.

### File: `.opencode/command/unleash.md`

| Section | Change Type | Lines (approx) | FR |
|---------|------------|----------------|-----|
| Description (frontmatter) | Edit | 1-8 | FR-001 |
| Description (body) | Edit | 12-21 | FR-001 |
| Step 1 -- Branch Safety Gate | Replace + Add | 46-82 | FR-001, FR-002, FR-003 |
| Step 2 -- Resumability Detection | Add conditional | 84-124 | FR-006 |
| Steps 1-3 -- Clarify/Plan/Tasks | Add skip logic | 126-228 | FR-004 |
| Step 4 -- Spec Review | Edit directory ref | 230-279 | FR-005, FR-007 |
| Steps 5-8 | Edit directory refs | 281-577 | FR-007 |
| Guardrails | Edit | 579-603 | FR-008 |

### File: `internal/scaffold/assets/opencode/command/unleash.md`

Exact copy of the live file above. Synced after all
edits are complete.

## Section-by-Section Change Detail

### 1. Frontmatter Description (lines 1-8)

**Before**:
```yaml
description: >
  Run the full Speckit pipeline autonomously...
```

**After**:
```yaml
description: >
  Run the full Speckit or OpenSpec pipeline
  autonomously...
```

### 2. Body Description (lines 12-21)

**Before**:
```
Autonomous Speckit pipeline execution. Takes a spec
from draft to demo-ready code...
```

**After**:
```
Autonomous pipeline execution for both Speckit
(strategic) and OpenSpec (tactical) changes. Takes a
spec from draft to demo-ready code...
```

### 3. Step 1 -- Branch Safety Gate (lines 46-82)

**Before** (3 branches):
1. `main` → STOP
2. `opsx/*` → STOP
3. Non-`NNN-*` → STOP
4. Run `check-prerequisites.sh`, extract `FEATURE_DIR`

**After** (4 branches):
1. `main` → STOP (unchanged)
2. `opsx/*` → OpenSpec mode:
   - Extract `<name>` from `opsx/<name>`
   - Set `FEATURE_DIR` = `openspec/changes/<name>/`
   - Check `FEATURE_DIR/tasks.md` exists → if not,
     STOP: "No tasks.md found for change `<name>`.
     Run `/opsx-propose` first."
   - Set `WORKFLOW_TIER` = `openspec`
   - Announce: "Detected OpenSpec change: `<name>`"
3. `NNN-*` → Speckit mode:
   - Run `check-prerequisites.sh` (unchanged)
   - Set `FEATURE_DIR` from JSON output
   - Set `WORKFLOW_TIER` = `speckit`
4. Other → STOP (unchanged)

### 4. Step 2 -- Resumability Detection (lines 84-124)

**Before**: Checks clarify/plan/tasks/spec-review/
implementation/code-review in sequence.

**After**: Same checks, but for OpenSpec mode:
- Clarify: always "done" (skip)
- Plan: always "done" (skip)
- Tasks: always "done" (skip — tasks.md existence
  was verified in Step 1)
- Spec review: check `<!-- spec-review: passed -->`
  in `FEATURE_DIR/tasks.md` (same as Speckit)
- Implementation: check all `[x]` in
  `FEATURE_DIR/tasks.md` (same as Speckit)
- Code review: check `<!-- code-review: passed -->`
  in `FEATURE_DIR/tasks.md` (same as Speckit)

### 5. Steps 1-3 -- Clarify/Plan/Tasks (lines 126-228)

**Before**: Always execute clarify, plan, tasks steps.

**After**: If `WORKFLOW_TIER` = `openspec`:
- Announce: "OpenSpec mode — artifacts from
  /opsx-propose, skipping clarify/plan/tasks"
- Skip directly to Step 4 (spec review)

If `WORKFLOW_TIER` = `speckit`: execute unchanged.

### 6. Step 4 -- Spec Review (lines 230-279)

**Before**: Hardcoded "Run in **Spec Review Mode** —
review the spec artifacts, not code."

**After**: Pass `FEATURE_DIR` to the review council.
The review council auto-detects the workflow tier from
the branch name, so no explicit mode override is needed
beyond "Run in **Spec Review Mode**." The feature
directory tells the council where to find artifacts.

### 7. Steps 5-8 (lines 281-577)

**Before**: References to "the feature directory" and
"spec.md" are implicit.

**After**: All references use `FEATURE_DIR`. For
OpenSpec, `spec.md` references in the Demo step (Step
8) are replaced with `proposal.md` when
`WORKFLOW_TIER` = `openspec`.

### 8. Guardrails (lines 579-603)

**Before**:
```
- **NEVER run on `main`** -- the command is for Speckit
  feature branches only
```

**After**:
```
- **NEVER run on `main`** -- the command is for Speckit
  (`NNN-*`) and OpenSpec (`opsx/*`) feature branches
```

## Variables Introduced

| Variable | Set In | Used In | Speckit Value | OpenSpec Value |
|----------|--------|---------|---------------|----------------|
| `FEATURE_DIR` | Step 1 | Steps 2-8 | `specs/NNN-name/` | `openspec/changes/<name>/` |
| `WORKFLOW_TIER` | Step 1 | Steps 1-3, 8 | `speckit` | `openspec` |

These are conceptual variables in the Markdown
instructions — they guide the agent's behavior, not
literal shell variables.

## Artifact Mapping

| Artifact | Speckit Path | OpenSpec Path |
|----------|-------------|---------------|
| Spec/Proposal | `FEATURE_DIR/spec.md` | `FEATURE_DIR/proposal.md` |
| Plan/Design | `FEATURE_DIR/plan.md` | `FEATURE_DIR/design.md` |
| Tasks | `FEATURE_DIR/tasks.md` | `FEATURE_DIR/tasks.md` |
| Quickstart | `FEATURE_DIR/quickstart.md` | N/A |
| Spec review marker | `FEATURE_DIR/tasks.md` | `FEATURE_DIR/tasks.md` |
| Code review marker | `FEATURE_DIR/tasks.md` | `FEATURE_DIR/tasks.md` |

## FR Traceability

| FR | Section(s) Modified | Verified By |
|----|---------------------|-------------|
| FR-001 | Step 1 (remove STOP, add detection) | SC-001 |
| FR-002 | Step 1 (extract name from branch) | SC-001 |
| FR-003 | Step 1 (check tasks.md exists) | SC-001 |
| FR-004 | Steps 1-3 (skip for OpenSpec) | SC-001 |
| FR-005 | Step 4 (pass FEATURE_DIR to council) | SC-002 |
| FR-006 | Step 2 (resumability markers) | SC-004 |
| FR-007 | Steps 4-8 (use FEATURE_DIR) | SC-001 |
| FR-008 | Guardrails + Step 1 (backward compat) | SC-003 |
| FR-009 | Scaffold asset sync | SC-005 |
| FR-010 | Existing tests pass | SC-005 |
