---
tag: schema-drift
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:18:14Z
identity: schema-drift-20261102T051814-jay-flowers
tier: draft
---

Schema-writer drift pattern: in schema-first Go projects, a runtime struct field tagged `json:"field,omitempty"` combined with a JSON Schema (generated from a separate mirror struct without omitempty) that lists the field in `required` produces artifacts that silently fail the project's own schema validation. Sample-based schema tests do not catch this because samples are hand-authored with the field present. The detecting test is a round-trip: write an artifact through the production writer function, then validate the written file against the canonical schema. Observed in unbound-force/unbound-force#557 (envelope `context` field: `internal/artifacts/artifacts.go` omits it via omitempty + nil-context convenience wrapper, while `schemas/envelope/v1.0.0.schema.json` requires it and `internal/schemas/types.go` EnvelopeSchema mirror declares it required).
