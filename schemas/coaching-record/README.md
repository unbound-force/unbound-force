# Coaching Record Schema

The coaching record payload captures Mx F's coaching output:
retrospective sessions with identified patterns and action items,
or coaching interactions with questions and insights.

## Producer

**Mx F** — the Manager hero.

## Consumers

- **All heroes** — coaching records inform team-wide improvements
- **Muti-Mind** — uses patterns for backlog prioritization

## Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `record_type` | string | Record type: retrospective or coaching |

## Conditional Fields (based on record_type)

### When record_type = "retrospective"

| Field | Type | Description |
|-------|------|-------------|
| `retrospective.date` | string | Session date (YYYY-MM-DD) |
| `retrospective.patterns` | array | Identified patterns |
| `retrospective.action_items` | array | Committed improvements |

### When record_type = "coaching"

| Field | Type | Description |
|-------|------|-------------|
| `coaching_interaction.topic` | string | Coaching topic |
| `coaching_interaction.questions` | array | Questions asked |
| `coaching_interaction.insights` | array | Insights surfaced |
| `coaching_interaction.outcome` | string | Session outcome |

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03-21 | Initial release |
