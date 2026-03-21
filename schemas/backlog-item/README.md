# Backlog Item Schema

The backlog item payload captures Muti-Mind's product backlog
output: prioritized work items with acceptance criteria in
Given/When/Then format.

## Producer

**Muti-Mind** — the Product Owner hero.

## Consumers

- **Mx F** — tracks backlog health and sprint planning
- **Cobalt-Crush** — implements items from the backlog
- **Gaze** — validates acceptance criteria

## Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Backlog item identifier (e.g., BI-001) |
| `title` | string | Item title |
| `type` | string | Item type (feature, bug, chore, spike) |
| `priority` | string | Priority level (P1, P2, P3) |
| `status` | string | Current status (backlog, ready, in-progress, done) |
| `acceptance_criteria` | array | Criteria in Given/When/Then format |

## Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Detailed description |
| `sprint` | string | Assigned sprint |
| `effort_estimate` | string | Effort estimate |
| `dependencies` | array | Dependent backlog item IDs |
| `related_specs` | array | Related specification IDs |

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03-21 | Initial release |
