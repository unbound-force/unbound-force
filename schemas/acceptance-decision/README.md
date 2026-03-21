# Acceptance Decision Schema

The acceptance decision payload captures Muti-Mind's verdict on
whether a completed backlog item meets its acceptance criteria.

## Producer

**Muti-Mind** — the Product Owner hero.

## Consumers

- **Mx F** — tracks acceptance rates and coaching patterns
- **Cobalt-Crush** — receives accept/reject feedback

## Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `item_id` | string | Backlog item being evaluated |
| `decision` | string | Decision: accept, reject, or conditional |
| `rationale` | string | Explanation for the decision |
| `criteria_met` | array | Acceptance criteria that passed |
| `criteria_failed` | array | Acceptance criteria that failed |
| `gaze_report_ref` | string | Path to the Gaze quality report used |
| `decided_at` | string | ISO 8601 timestamp of decision |

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03-21 | Initial release |
