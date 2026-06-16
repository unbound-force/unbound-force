# Issue Triage Schema

The issue triage payload captures the Divisor review panel's
assessment of a GitHub issue: per-agent classification, validity
determination, objectivity analysis, and actions taken.

## Producer

**The Divisor** (review panel) — executing the `/triage-issue` slash command.

## Consumers

- **Mx F** — tracks triage pattern trends, classification accuracy, tone consistency
- **Muti-Mind** — detects recurring issue categories for backlog health analysis
- **Cobalt-Crush** — learns from past triage decisions to improve issue quality

## Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `issue_number` | integer | GitHub issue number |
| `issue_url` | string | Full GitHub issue URL |
| `repo` | string | Repository in owner/repo format |
| `title` | string | Issue title |
| `author` | string | Issue author GitHub login |
| `category` | string (enum) | Classification: bug, feature, enhancement, question, opinion, duplicate, needs-clarification |
| `validity` | string (enum) | Panel verdict: valid, invalid, needs-clarification |
| `objectivity` | string (enum) | Assessment: objective, subjective |
| `assessments` | array | Per-agent assessment records (agent, verdict, category, objectivity, reasoning, split_recommendation) |
| `actions_taken` | object | Actions performed (labels_applied, comment_posted, child_issues_created, label_creation_failed) |
| `summary` | object | Aggregate statistics (agents_consulted, agents_available, consensus, dissenting_agents) |

## Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `duplicate_of` | integer or null | Issue number this duplicates (null if not duplicate) |
| `split_issues` | array | Child issues created from splitting (number, title, url) |
| `assessments[].split_recommendation` | array or null | Proposed child issues from agent (null if no split) |

## Schema Evolution

Minor versions add optional fields and may relax
`additionalProperties` constraints. Major versions may remove
or rename fields. Consumers should handle unknown fields
gracefully when consuming artifacts from newer minor versions.

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-06-16 | Initial release |
