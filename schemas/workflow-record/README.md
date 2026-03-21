# Workflow Record Schema

The workflow record payload captures the complete lifecycle history
of a feature workflow through the 6-stage hero pipeline: define,
implement, validate, review, accept, measure.

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

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03-21 | Initial release |
