# Feedback Triage Schema

The feedback triage payload captures the author's response to PR
review feedback: per-item classification, assessment evidence,
author decisions, and execution outcomes.

## Producer

**Cobalt-Crush** — executing the `/uf.address-feedback` slash command.

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

New producers emit schema 1.1.0 (`v1.1.0.schema.json`); 1.0.0
(`v1.0.0.schema.json`) stays for historical reads.

## Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `items[].file` | string or null | File path (null for general PR comments) |
| `items[].line` | integer or null | Line number (null for general PR comments) |
| `items[].decision_reasoning` | string or null | Author's reasoning for reject/modify/ask decisions |
| `items[].commit_sha` | string or null | Commit SHA for accepted/modified code changes |
| `items[].filed_issue_number` | integer or null | Issue filed for a `file` decision (1.1.0) |
| `items[].filed_issue_url` | string or null | URL of the issue filed for a `file` decision (1.1.0) |
| `summary.filed` | integer | Items filed as issues (1.1.0) |
| `summary.replied` | integer | Items answered with a reply only (1.1.0) |

## FILE and REPLY Items

`/uf.address-feedback` files or answers MEDIUM and LOW findings
without changing code (feedback-loop-scope). Version 1.1.0 records
them directly:

- `decision: "file"`: filed as an issue and answered with a reply
  linking it; `filed_issue_number` and `filed_issue_url` name the
  issue, and the item counts under `summary.filed`
- `decision: "reply"`: answered with a reply only; the item counts
  under `summary.replied`

Both carry `recommendation: "author-decides"` and a null
`commit_sha`.

Version 1.0.0 has no such decisions. Artifacts written before 1.1.0
(PR #993) record these items as `decision: "reject"` with a
`decision_reasoning` prefix of `filed: <issue URL>` or `reply:`, and
count them under `rejected`. A reader of historical 1.0.0 artifacts
tells them apart from a genuine reject by that prefix.

## Version Compatibility

1.1.0 is a strict superset of 1.0.0: every 1.0.0 artifact validates
under 1.1.0, and the new fields are optional. A reader pinned to the
1.0.0 schema rejects the `file` and `reply` decisions rather than
misreading them; read the artifact's `schema_version` and validate
against the matching schema.

## Schema Evolution

Minor versions add optional fields and may relax
`additionalProperties` constraints. Major versions may remove
or rename fields. Consumers should handle unknown fields
gracefully when consuming artifacts from newer minor versions.

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.1.0 | 2026-11-08 | Adds the `file` and `reply` decisions, `filed_issue_number`, `filed_issue_url`, and the `filed` and `replied` summary counts (#999) |
| 1.0.0 | 2026-05-13 | Initial release |
