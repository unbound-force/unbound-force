# Feedback Triage Schema

The feedback triage payload captures the author's response to PR
review feedback: per-item classification, assessment evidence,
author decisions, and execution outcomes.

## Producer

**Cobalt-Crush** — executing the `/address-feedback` slash command.

## Consumers

- **Mx F** — tracks accept/reject ratios, review pattern trends, coaching insights
- **Cobalt-Crush** — learns from past triage decisions to proactively avoid common feedback
- **Muti-Mind** — detects recurring feedback gaps for backlog item creation

## Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `pr_number` | integer | Pull request number |
| `pr_url` | string | Pull request URL |
| `branch` | string | PR source branch name |
| `round` | integer | Triage round number (increments per invocation) |
| `items` | array | Per-item triage records (thread_id, reviewer, reviewer_role, classification, tier, evidence, recommendation, decision, decision_reasoning, commit_sha, divisor_agents_used, tier2_unavailable, conflict_flag) |
| `summary` | object | Aggregate statistics (total_items, accepted, modified, rejected, asked, tier1_count, tier2_count, divisor_agents_invoked) |

## Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `items[].file` | string or null | File path (null for general PR comments) |
| `items[].line` | integer or null | Line number (null for general PR comments) |
| `items[].decision_reasoning` | string or null | Author's reasoning for reject/modify/ask decisions |
| `items[].commit_sha` | string or null | Commit SHA for accepted/modified code changes |

## Schema Evolution

Minor versions add optional fields and may relax
`additionalProperties` constraints. Major versions may remove
or rename fields. Consumers should handle unknown fields
gracefully when consuming artifacts from newer minor versions.

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-05-13 | Initial release |
