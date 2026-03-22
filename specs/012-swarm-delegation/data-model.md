# Data Model: Swarm Delegation Workflow

**Date**: 2026-03-22
**Branch**: `012-swarm-delegation`

## Entities

### WorkflowStage (modified)

Existing entity in `internal/orchestration/models.go`.
One new field added.

| Field | Type | Description | New? |
|-------|------|-------------|------|
| stage_name | string | Stage identifier (define, implement, validate, review, accept, reflect) | No (reflect replaces measure) |
| hero | string | Hero name assigned to this stage | No |
| status | string | Stage status (pending, active, completed, skipped, failed) | No |
| execution_mode | string | `"human"` or `"swarm"` -- who drives this stage | **Yes** |
| artifacts_produced | []string | Artifact paths produced during this stage | No |
| artifacts_consumed | []string | Artifact paths consumed from previous stages | No |
| started_at | *time | When the stage was activated | No |
| completed_at | *time | When the stage was completed | No |
| skip_reason | string | Why the stage was skipped (if applicable) | No |
| error | string | Error description (if failed) | No |

**Validation rules**:
- `execution_mode` MUST be `"human"`, `"swarm"`, or empty
  (empty treated as `"human"` for backward compatibility)
- JSON tag: `json:"execution_mode,omitempty"` to avoid
  breaking legacy serialization

### WorkflowInstance (modified)

Existing entity in `internal/orchestration/models.go`.
No new fields. Status value set expanded.

| Field | Existing Values | New Values |
|-------|----------------|------------|
| status | active, completed, failed, escalated | + **awaiting_human** |

*Note*: `pending` and `skipped` are stage-level statuses
only. Workflows are created with `active` status
immediately via `Start()`.

**State transitions** (workflow-level):

```text
active ──advance──▶ active           (human→human or swarm→swarm)
active ──advance──▶ awaiting_human   (swarm→human boundary)
awaiting_human ──advance──▶ active   (human resumes)
active ──advance──▶ completed        (last stage done)
active ──escalate──▶ escalated       (max iterations or contradiction)
active ──reject──▶ failed            (acceptance rejection)
```

### Stage Constants (modified)

| Old Constant | New Constant | String Value |
|-------------|-------------|-------------|
| StageMeasure | **StageReflect** | `"reflect"` (was `"measure"`) |

All other constants unchanged: `StageDefine`,
`StageImplement`, `StageValidate`, `StageReview`,
`StageAccept`.

### Execution Mode Constants (new)

| Constant | Value | Description |
|----------|-------|-------------|
| ModeHuman | `"human"` | Stage advanced by human operator |
| ModeSwarm | `"swarm"` | Stage advanced by swarm autonomously |

### Status Constants (modified)

| Constant | Value | New? |
|----------|-------|------|
| StatusAwaitingHuman | `"awaiting_human"` | **Yes** |

All existing status constants unchanged.

## Mappings

### StageOrder (modified)

```text
[define, implement, validate, review, accept, reflect]
```

Last element changed from `"measure"` to `"reflect"`.

### StageHeroMap (modified)

| Stage | Hero |
|-------|------|
| define | muti-mind |
| implement | cobalt-crush |
| validate | gaze |
| review | divisor |
| accept | muti-mind |
| reflect | mx-f |

Last entry changed from `measure: mx-f` to
`reflect: mx-f`.

### StageExecutionModeMap (new)

| Stage | Execution Mode |
|-------|---------------|
| define | human |
| implement | swarm |
| validate | swarm |
| review | swarm |
| accept | human |
| reflect | swarm |

### heroSpecs (modified)

The `mx-f` entry's `Role` field changes from
`StageMeasure` to `StageReflect`.

## State Transition Logic

### Advance() Checkpoint Detection

When `Advance()` finds the next non-skipped pending stage:

```text
IF current_stage.execution_mode is "swarm"
   AND next_stage.execution_mode is "human"
THEN
   set workflow.status = "awaiting_human"
   do NOT activate next_stage
   save and return
```

When `Advance()` is called on an `awaiting_human` workflow:

```text
IF workflow.status == "awaiting_human"
THEN
   find the next pending non-skipped stage
   activate it (set status=active, started_at=now)
   set workflow.status = "active"
   save and return
```

### Backward Compatibility

Empty `execution_mode` is treated as `"human"`. Since
human-to-human transitions never trigger the checkpoint
condition, legacy workflows behave identically to
pre-change behavior.

### Latest() Discovery

`WorkflowStore.Latest()` returns workflows matching either
`StatusActive` or `StatusAwaitingHuman` for the given
branch. This ensures paused workflows are discoverable via
`/workflow status` and `/workflow advance` without
requiring the operator to provide a workflow ID.
