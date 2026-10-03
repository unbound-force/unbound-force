# Artifact Schema Registry

This directory is the machine-readable payload registry for the Hero
Interface Contract. Every registered artifact has a draft 2020-12
schema, samples, and producer and consumer documentation.

## Review Artifact Relationships

| Type | Producer | Consumers | Purpose |
| --- | --- | --- | --- |
| `review-dispatch` | Divisor | Divisor, Mx F | Provenance |
| `review-verdict` | Divisor | Mx F, Cobalt, Muti-Mind | Decision |

The Divisor emits both artifacts for advisor-backed workflows.
`review-dispatch` is additive observability and model provenance.
`review-verdict` remains the downstream decision artifact.

New `review-verdict` producers emit 2.0.0. Migrated consumers support
v1 historical reads and v2 current reads. A v1-only consumer rejects
v2 through the Hero Interface Contract major-version rule.
