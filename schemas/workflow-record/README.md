# Workflow Record Schema

The workflow record payload captures the complete lifecycle history
of a feature workflow through the 6-stage hero pipeline: define,
implement, validate, review, accept, reflect.

## Producer

**Swarm Orchestration** — the workflow engine.

## Consumers

- **Mx F** — uses workflow data for metrics and coaching
- **Muti-Mind** — uses outcomes for backlog management

## Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `workflow_id` | string | Unique workflow instance identifier |
| `backlog_item_id` | string | Associated backlog item |
| `stages` | array | Stage records (name, hero, status, timestamps, artifacts) |
| `artifacts` | array | All artifact types produced during workflow |
| `decisions` | array | Decision points (review verdicts, acceptance decisions) |
| `total_elapsed_time` | string | Total workflow duration |
| `outcome` | string | Final outcome: shipped, rejected, or abandoned |

## Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `learning_feedback` | array | Cross-hero learning recommendations |

## Stage Fields (per stage object)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `stage_name` | string | Yes | Stage identifier (define, implement, validate, review, accept, reflect) |
| `hero` | string | Yes | Hero name assigned to this stage |
| `status` | string | Yes | Stage status (pending, active, completed, skipped, failed) |
| `execution_mode` | string | No | `"human"` or `"swarm"` — who drives this stage (v1.1.0+) |
| `artifacts_produced` | array | No | Artifact paths produced during this stage |
| `artifacts_consumed` | array | No | Artifact paths consumed from previous stages |
| `started_at` | string | No | ISO 8601 timestamp when the stage was activated |
| `completed_at` | string | No | ISO 8601 timestamp when the stage was completed |
| `skip_reason` | string | No | Why the stage was skipped (if applicable) |
| `error` | string | No | Error description (if failed) |

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.1.0 | 2026-03-22 | Added `execution_mode` field to stage objects, renamed `measure` stage to `reflect` |
| 1.0.0 | 2026-03-21 | Initial release |
