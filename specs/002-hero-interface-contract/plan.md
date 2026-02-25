# Implementation Plan: Hero Interface Contract

**Branch**: `002-hero-interface-contract` | **Date**: 2026-02-25
**Spec**: [spec.md](spec.md)
**Input**: Feature specification from
`/specs/002-hero-interface-contract/spec.md`

## Summary

Define the standard structure, contracts, and integration
points that every hero repository must implement to be a member
of the Unbound Force swarm. This spec produces three primary
deliverables:

1. The Hero Interface Contract document — a governance
   specification defining required repository structure,
   artifact envelope format, OpenCode agent/command
   conventions, speckit integration requirements, and hero
   manifest format.
2. A hero manifest JSON Schema (`.unbound-force/hero.json`
   schema) for machine-readable hero metadata.
3. A bash validation script that checks any repository against
   the contract's structural requirements.

The artifact envelope is defined conceptually (fields, rules,
versioning semantics). The formal JSON Schema for the envelope
and all artifact type payload schemas are deferred to Spec 009
(Shared Data Model). The speckit distribution mechanism is
deferred to Spec 003 (Speckit Framework).

## Technical Context

**Language/Version**: Markdown (contract document) + JSON Schema
draft 2020-12 (hero manifest schema) + Bash (validation script)
**Primary Dependencies**: OpenCode (agent runtime), speckit
(pipeline integration)
**Storage**: Filesystem only — contract document in
`specs/002-hero-interface-contract/`, hero manifest schema as a
reference JSON Schema, validation script at repo root
**Testing**: Manual validation against SC-001 through SC-007.
Bash validation script tested against Gaze and Website repos.
Hero manifest schema validated with a sample `hero.json` for
Gaze.
**Target Platform**: Any Unbound Force hero repository
**Project Type**: Governance document + JSON Schema definition +
Bash validation script
**Performance Goals**: N/A (documents and schemas, not runtime
software)
**Constraints**: Must be compatible with existing Gaze and
Website repo structures. Must not mandate retroactive changes
that break current workflows — a migration path is required for
pre-existing repos.
**Scale/Scope**: Governs 5 hero repos + 1 website repo. Hero
manifest schema consumed by discovery tooling and the Swarm
plugin.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after
Phase 1 design.*

### Pre-Design Check

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Autonomous Collaboration | PASS | The contract defines the artifact envelope format — the mechanism for artifact-based communication (constitution line 72-74: "Inter-hero communication MUST use the artifact envelope format defined by the Hero Interface Contract"). It standardizes how heroes produce and consume artifacts without runtime coupling. The hero manifest enables discovery without synchronous coordination. |
| II. Composability First | PASS | The contract requires heroes to be independently installable (FR-001 repo structure). The hero manifest (FR-002) declares capabilities without mandating dependencies. Artifact consumption is optional — a hero MUST function standalone even if consumed artifact types are unavailable (US5 acceptance scenario 3). The validation script does not block hero operation, only reports compliance. |
| III. Observable Quality | PASS | The contract requires JSON minimum output (FR-010). The artifact envelope includes provenance metadata (hero, version, timestamp). The hero manifest is a machine-parseable JSON file. The validation script (FR-008) provides automated, reproducible evidence of contract compliance. The hero manifest JSON Schema enables machine validation of manifest files. |

**Gate result**: PASS — all three principles satisfied.
Proceeding to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/002-hero-interface-contract/
├── spec.md              # Feature specification (exists)
├── plan.md              # This file
├── research.md          # Phase 0: Research decisions
├── data-model.md        # Phase 1: Entity models
└── quickstart.md        # Phase 1: Usage guide
```

### Source Code (repository root)

```text
# Hero manifest schema (reference for hero repos)
schemas/
└── hero-manifest/
    └── v1.0.0.schema.json

# Validation script
scripts/
└── validate-hero-contract.sh
```

**Structure Decision**: No traditional `src/` or `tests/`
directories. This feature produces governance documents, a
JSON Schema definition, and a bash validation script. The
contract document lives in `specs/002-hero-interface-contract/`
per speckit convention. The hero manifest schema lives in
`schemas/hero-manifest/` as a reference schema that hero repos
will validate against. The validation script lives in
`scripts/` at repository root.

### Post-Design Re-evaluation

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Autonomous Collaboration | PASS | No regressions. The data model defines the artifact envelope conceptually with all required provenance fields. Hero manifests declare artifact production/consumption without requiring runtime handshakes. The validation script checks structural compliance without requiring other heroes to be present. |
| II. Composability First | PASS | No regressions. The hero manifest JSON Schema is independently usable — any tool can validate a manifest without installing other heroes. The validation script reports findings but does not block hero operation. Missing optional elements produce warnings, not errors. |
| III. Observable Quality | PASS | No regressions. The hero manifest schema is a formal JSON Schema (draft 2020-12) enabling automated validation. The validation script produces structured output (pass/fail per check). All quality claims about contract compliance are backed by the automated script. |

**Post-design gate result**: PASS — all three principles remain
satisfied after design phase. No regressions from pre-design.

## Complexity Tracking

No constitution violations to justify. All three principles
pass cleanly.
