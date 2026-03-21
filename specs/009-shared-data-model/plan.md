# Implementation Plan: Shared Data Model

**Branch**: `009-shared-data-model` | **Date**: 2026-03-21 | **Spec**: [spec.md](spec.md)

## Summary

Spec 009 creates the shared data model for inter-hero communication:
JSON Schemas for the artifact envelope and all 7 artifact types,
generated from existing Go structs using `invopop/jsonschema`.
Includes a schema registry (`schemas/`), sample artifacts, READMEs,
convention pack structural validation, and CI validation via
`go test`.

## Technical Context

**Language/Version**: Go 1.24+ (schema generation + validation), JSON Schema draft 2020-12
**Primary Dependencies**: `github.com/invopop/jsonschema` (Go struct → JSON Schema generation), existing Go structs in `internal/artifacts`, `internal/metrics`, `internal/orchestration`, `internal/coaching`, `internal/impediment`
**Storage**: JSON Schema files at `schemas/{type}/v1.0.0.schema.json`, sample artifacts at `schemas/{type}/samples/`
**Testing**: `go test` — generate schemas, validate samples against them, drift detection
**Constraints**: Go structs are source of truth. Schemas derived, not authored manually.

## Constitution Check

### I. Autonomous Collaboration — PASS
Schemas are the formal contracts between heroes. Each hero produces
self-describing artifacts that any consumer can validate independently.

### II. Composability First — PASS
Schemas are independently usable. A hero can validate artifacts
without needing other heroes installed.

### III. Observable Quality — PASS
JSON Schema validation produces machine-parseable error messages.
All artifacts include provenance metadata (hero, version, timestamp).

### IV. Testability — PASS
Schema generation from Go structs is deterministic and testable.
Sample validation is automated via `go test`.

## Project Structure

```text
# Schema generation tool
internal/schemas/
├── generate.go        # Schema generator using invopop/jsonschema
├── generate_test.go   # Tests: generate + validate samples
├── validate.go        # Runtime validation helper
└── validate_test.go   # Validation round-trip tests

# Schema registry (generated output)
schemas/
├── envelope/
│   ├── v1.0.0.schema.json
│   ├── samples/sample-envelope.json
│   └── README.md
├── quality-report/
│   ├── v1.0.0.schema.json
│   ├── samples/sample-quality-report.json
│   └── README.md
├── review-verdict/        # ... same structure
├── backlog-item/
├── acceptance-decision/
├── metrics-snapshot/
├── coaching-record/
├── workflow-record/
└── convention-pack/
    ├── validator.go       # Markdown structural validator
    └── validator_test.go

# Convention pack validation
internal/schemas/packvalidator.go   # Validates Markdown pack structure
```

## Complexity Tracking

No constitution violations. Schema generation is mechanical —
the complexity is in mapping Go struct tags to JSON Schema draft
2020-12 constructs. `invopop/jsonschema` handles the heavy lifting.
<!-- scaffolded by unbound vdev -->
