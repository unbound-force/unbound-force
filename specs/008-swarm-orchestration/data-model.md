# Data Model: Swarm Orchestration

**Spec**: 008-swarm-orchestration
**Date**: 2026-03-20

## Entities

### Workflow Instance

A single execution of the hero lifecycle for a feature.
Persisted as JSON at `.unbound-force/workflows/{workflow_id}.json`.

| Attribute | Type | Notes |
|-----------|------|-------|
| workflow_id | string | UUID or `wf-{branch}-{timestamp}` |
| feature_branch | string | Git branch name |
| backlog_item_id | string | Muti-Mind backlog item ID (e.g., BI-042) |
| stages | []WorkflowStage | Ordered list of lifecycle stages |
| current_stage | int | Index into stages[] |
| started_at | ISO 8601 | When workflow was created |
| completed_at | ISO 8601 (opt) | When workflow finished |
| status | string | active/completed/failed/escalated |
| available_heroes | []string | Heroes detected at workflow start |
| iteration_count | int | Total review-fix iterations |

**State transitions**:
```text
          ┌──────────┐
          │  active   │
          └────┬──────┘
               │ all stages complete
               ├──────────────────────► completed
               │ stage failed + no fallback
               ├──────────────────────► failed
               │ max iterations or conflict
               └──────────────────────► escalated
```

### Workflow Stage

One step in the hero lifecycle.

| Attribute | Type | Notes |
|-----------|------|-------|
| stage_name | string | define/implement/validate/review/accept/measure |
| hero | string | muti-mind/cobalt-crush/gaze/divisor/mx-f |
| status | string | pending/active/completed/skipped/failed |
| artifacts_produced | []string | Paths to artifacts produced |
| artifacts_consumed | []string | Paths to artifacts consumed |
| started_at | ISO 8601 (opt) | When stage began |
| completed_at | ISO 8601 (opt) | When stage finished |
| skip_reason | string (opt) | Why stage was skipped (hero unavailable) |
| error | string (opt) | Error message if failed |

**Stage sequence**:
```text
define → implement → validate → review → accept → measure
  │          │           │          │         │         │
  │          │           │          │         │         └─ mx-f
  │          │           │          └─────────┘
  │          │           │          iteration loop
  │          │           └─ gaze (skippable)
  │          └─ cobalt-crush
  └─ muti-mind
```

### Artifact Context

Metadata added to artifact envelopes for workflow tracking.

| Attribute | Type | Notes |
|-----------|------|-------|
| branch | string | Git branch (e.g., `feat/health-check`) |
| commit | string (opt) | Git commit SHA |
| backlog_item_id | string (opt) | Originating backlog item |
| correlation_id | string (opt) | UUID for request tracing |
| workflow_id | string (opt) | Links artifact to workflow |

### Learning Feedback

A cross-hero recommendation from pattern analysis.

| Attribute | Type | Notes |
|-----------|------|-------|
| id | string | `LF-NNN` auto-incrementing |
| source_hero | string | Hero that observed the pattern |
| target_hero | string | Hero that should adapt |
| pattern_observed | string | What pattern was detected |
| recommendation | string | What change is proposed |
| supporting_data | map | Evidence (metrics, finding counts, etc.) |
| status | string | proposed/accepted/rejected/implemented |
| created_at | ISO 8601 | When feedback was generated |
| workflow_ids | []string | Workflows that contributed to the pattern |

### Workflow Record

Complete lifecycle history, produced as an artifact on
workflow completion.

| Attribute | Type | Notes |
|-----------|------|-------|
| workflow_id | string | Same as WorkflowInstance |
| backlog_item_id | string | Originating backlog item |
| stages | []WorkflowStage | All stages with timing |
| artifacts | []string | All artifact paths |
| decisions | []Decision | All acceptance/review decisions |
| total_elapsed_time | duration | Start to completion |
| outcome | string | shipped/rejected/abandoned |
| learning_feedback | []string | LF-NNN IDs produced |

### Decision

A decision point in the workflow.

| Attribute | Type | Notes |
|-----------|------|-------|
| type | string | review-verdict/acceptance-decision |
| hero | string | Which hero made it |
| result | string | approve/reject/conditional/request-changes |
| rationale | string | Why this decision |
| iteration | int | Which iteration |
| timestamp | ISO 8601 | When decided |

### Hero Status

Availability state of a hero, detected at workflow start.

| Attribute | Type | Notes |
|-----------|------|-------|
| name | string | Hero name (muti-mind, cobalt-crush, gaze, divisor, mx-f) |
| role | string | Hero role (define, implement, validate, review, accept, measure) |
| available | bool | Whether the hero is detected |
| agent_file | string | Expected agent file path (e.g., `muti-mind-po.md`) |
| detection_method | string | How availability was determined (file_exists, exec_lookpath) |

### Swarm Skills Package

Not a Go struct — a Markdown file at
`.opencode/skill/unbound-force-heroes/SKILL.md`.

Content describes:
- Hero names and roles
- Natural language routing patterns
- Workflow stage sequence
- Escalation rules

## Relationships

```text
Workflow Instance ──contains──► Workflow Stage[] (6 stages)
       │                              │
       │                              │ produces/consumes
       │                              ▼
       │                        Artifact (at .unbound-force/artifacts/)
       │                              │
       │                              │ includes
       │                              ▼
       │                        Artifact Context (branch, workflow_id)
       │
       │ on completion produces
       ▼
Workflow Record ──contains──► Decision[]
       │
       │ analyzed by
       ▼
Learning Feedback ──targets──► Hero (convention pack, priority, etc.)
```

## Validation Rules

1. `workflow_id` MUST be unique across all workflows.
2. `stages` MUST contain exactly 6 entries in order:
   define, implement, validate, review, accept, measure.
3. `current_stage` MUST be a valid index into `stages[]`.
4. A stage MUST NOT transition to `completed` without
   producing at least one artifact (or being `skipped`).
5. `skipped` stages MUST have a `skip_reason`.
6. `failed` stages MUST have an `error` message.
7. Workflow status `completed` requires all non-skipped
   stages to be `completed`.
8. Artifact context `branch` MUST match the workflow's
   `feature_branch` (prevents cross-contamination).
9. Learning feedback MUST reference at least one
   `workflow_id` as evidence.
10. `iteration_count` MUST NOT exceed 3 before escalation.
<!-- scaffolded by unbound vdev -->
