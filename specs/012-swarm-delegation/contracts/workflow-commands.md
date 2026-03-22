# Contract: Workflow Commands

**Date**: 2026-03-22
**Branch**: `012-swarm-delegation`

This document defines the updated interface contract for
the `/workflow` commands after swarm delegation and
reflect stage changes.

## /workflow start

### Output Format (updated)

```
Workflow started: wf-feat-login-20260322T143000

Available heroes:
  ✓ Muti-Mind (define) [human]
  ✓ Cobalt-Crush (implement) [swarm]
  ✓ Gaze (validate) [swarm]
  ✓ The Divisor (review) [swarm]
  ✓ Muti-Mind (accept) [human]
  ✓ Mx F (reflect) [swarm]

Stage 1/6: define (Muti-Mind) [human]
  Next: Run /speckit.specify to create the feature spec.
```

### Changes from previous contract

- Stage 6 renamed from `measure` to `reflect`
- Each stage line includes `[human]` or `[swarm]` mode
  indicator
- Hero availability list includes mode per stage

## /workflow status

### Output Format (updated)

#### Active workflow

```
Workflow: wf-feat-login-20260322T143000
Branch: feat/login
Backlog Item: BI-042
Status: active
Started: 2026-03-22T14:30:00Z
Iterations: 1

Stages:
  ✓ define      (muti-mind)     completed  15m   [human]
  ◉ implement   (cobalt-crush)  active     2h    [swarm]
  ○ validate    (gaze)          pending          [swarm]
  ○ review      (divisor)       pending          [swarm]
  ○ accept      (muti-mind)     pending          [human]
  ○ reflect     (mx-f)          pending          [swarm]
```

#### Awaiting human

```
Workflow: wf-feat-login-20260322T143000
Branch: feat/login
Status: awaiting_human

Stages:
  ✓ define      (muti-mind)     completed  15m   [human]
  ✓ implement   (cobalt-crush)  completed  2h    [swarm]
  ✓ validate    (gaze)          completed  5m    [swarm]
  ✓ review      (divisor)       completed  20m   [swarm]
  ⏸ accept      (muti-mind)     pending          [human]  ← your turn
  ○ reflect     (mx-f)          pending          [swarm]

Run /workflow advance to resume.
```

### Changes from previous contract

- Stage 6 renamed from `measure` to `reflect`
- Each stage row includes `[human]` or `[swarm]` mode
- New `⏸` indicator for the next human-mode stage when
  workflow is `awaiting_human`
- `← your turn` annotation on the pending human stage
- Resume instruction shown when status is `awaiting_human`

## /workflow advance

### Output Formats (updated)

#### Stage Advanced (within same mode)

```
Stage advanced: define → implement

Workflow: wf-feat-login-20260322T143000
Stage 2/6: implement (Cobalt-Crush) [swarm]
  Next: Swarm will implement from tasks.md.
```

#### Checkpoint Reached (swarm → human boundary)

```
Workflow paused: wf-feat-login-20260322T143000

Swarm completed: implement → validate → review
Awaiting human action at: accept (Muti-Mind) [human]

Run /workflow advance to resume and accept the increment.
```

#### Resumed from Checkpoint

```
Workflow resumed: wf-feat-login-20260322T143000

Stage 5/6: accept (Muti-Mind) [human]
  Next: Review the increment and make an acceptance decision.
```

#### Workflow Completed

```
Workflow completed: wf-feat-login-20260322T143000

All stages:
  ✓ define      (muti-mind)     15m   [human]
  ✓ implement   (cobalt-crush)  2h    [swarm]
  ✓ validate    (gaze)          5m    [swarm]
  ✓ review      (divisor)       20m   [swarm]
  ✓ accept      (muti-mind)     10m   [human]
  ✓ reflect     (mx-f)          5m    [swarm]

Total time: 3h15m
Outcome: shipped
```

#### Escalation (unchanged)

```
Workflow escalated: wf-feat-login-20260322T143000
Reason: Maximum review iterations (3) reached.

Unresolved findings require manual review.
```

### Changes from previous contract

- Stage 6 renamed from `measure` to `reflect`
- New "Checkpoint Reached" output format when swarm
  completes and workflow pauses at human boundary
- New "Resumed from Checkpoint" output format when
  human advances from `awaiting_human` status
- All stage rows include `[human]` or `[swarm]` mode

## Workflow JSON Schema (updated)

### WorkflowStage object (new field)

```json
{
  "stage_name": "implement",
  "hero": "cobalt-crush",
  "status": "active",
  "execution_mode": "swarm",
  "artifacts_produced": [],
  "artifacts_consumed": [],
  "started_at": "2026-03-22T14:31:00Z"
}
```

### WorkflowInstance status values

Valid values: `"active"`, `"completed"`, `"failed"`,
`"escalated"`, `"awaiting_human"`

### Backward compatibility

Legacy JSON without `execution_mode` fields is valid.
Missing fields are treated as `"human"` mode.
