# Review Verdict Schema

The review verdict payload captures The Divisor's multi-persona
code review output: individual persona assessments, the council's
aggregate decision, and any unresolved findings.

## Producer

**The Divisor** — the PR Reviewer Council hero.

## Consumers

- **Mx F** — tracks review iterations and coaching patterns
- **Cobalt-Crush** — addresses findings and learns from patterns
- **Muti-Mind** — uses review status for acceptance decisions

## Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `persona_verdicts` | array | Individual persona assessments (persona, verdict, findings, summary) |
| `council_decision` | string | Overall decision: APPROVED, CHANGES_REQUESTED, or ESCALATED |
| `iteration_count` | integer | Number of review iterations |
| `pr_url` | string | Pull request URL |
| `convention_pack_used` | string | Convention pack ID used for review |

## Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `unresolved_findings` | array | Findings not yet addressed |

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03-21 | Initial release |
