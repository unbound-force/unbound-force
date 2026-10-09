# Review Verdict Schema

The review verdict payload is The Divisor's canonical downstream
decision artifact. It captures persona assessments, the council's
aggregate decision, and unresolved findings.

## Producer

**The Divisor** — the PR Reviewer Council hero.

## Consumers

- **Mx F** — tracks review iterations and coaching patterns
- **Cobalt-Crush** — addresses findings and learns from patterns
- **Muti-Mind** — uses review status for acceptance decisions

## Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `persona_verdicts` | array | Persona assessments |
| `council_decision` | string | Versioned overall decision |
| `iteration_count` | integer | Number of review iterations |
| `pr_url` | string | Pull request URL |
| `convention_pack_used` | string | Convention pack ID |

## Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `unresolved_findings` | array | Findings not yet addressed |

## Version 2 Migration

New producers MUST emit schema 2.0.0. Its decision values are
`APPROVED`, `CHANGES_REQUESTED`, `ESCALATED`, `INCONCLUSIVE`, and
`UNAVAILABLE`. Version 1.0.0 remains available for historical reads.

The projection from `review-dispatch` is:

| Dispatch verdict | Canonical decision |
| --- | --- |
| `APPROVE` | `APPROVED` |
| `APPROVE WITH ADVISORIES` | `ESCALATED` |
| `REQUEST CHANGES` | `CHANGES_REQUESTED` |
| `INCONCLUSIVE` | `INCONCLUSIVE` |
| `UNAVAILABLE` | `UNAVAILABLE` |

The Divisor emits `review-dispatch` additively for observability and
model provenance. Mx F, Cobalt-Crush, and Muti-Mind consume this
`review-verdict` artifact for decisions. They MUST block automated
progression for every value except `APPROVED` unless an existing human
gate resolves it.

Migrated readers support historical v1 and current v2 explicitly.
Unmigrated v1-only readers reject v2 because the major versions differ.

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 2.0.0 | 2026-10-01 | Adds native no-success decisions |
| 1.0.0 | 2026-03-21 | Initial release |
