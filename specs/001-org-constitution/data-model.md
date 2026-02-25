# Data Model: Unbound Force Organization Constitution

**Spec**: [spec.md](spec.md) | **Date**: 2026-02-25

## Entities

### Org Constitution

The root governance document for the Unbound Force organization.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| title | string | yes | Document title (e.g., "Unbound Force Constitution") |
| version | semver string | yes | Constitution version (MAJOR.MINOR.PATCH) |
| ratification_date | ISO 8601 date | yes | Date the constitution was first adopted |
| last_amended_date | ISO 8601 date | yes | Date of the most recent amendment |
| principles | Core Principle[] | yes | Exactly three core principles (I, II, III) |
| hero_alignment_rules | string[] | yes | Rules governing hero constitution alignment |
| development_workflow | Workflow Rule[] | yes | Development process rules (branching, review, CI, etc.) |
| governance_rules | Governance Rule[] | yes | Amendment, versioning, supremacy, compliance rules |

**Lifecycle**: Template → Ratified → Amended (via PR) → Superseded (if MAJOR bump replaces principles)

**File location**: `.specify/memory/constitution.md`

### Core Principle

A named, numbered principle with MUST/SHOULD/MAY rules.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| number | roman numeral (I/II/III) | yes | Principle ordering |
| name | string | yes | Short name (e.g., "Autonomous Collaboration") |
| description | string | yes | Opening statement of the principle |
| must_rules | string[] | yes | Non-negotiable rules (minimum 3) |
| should_rules | string[] | yes | Recommended rules (minimum 1) |
| may_rules | string[] | no | Optional allowances |
| rationale | string | yes | Explanation of why this principle matters |

**Validation**: Each principle MUST have >= 3 MUST rules and >= 1 SHOULD rule (FR-002).

### Hero Constitution

A per-repository constitution that extends the org constitution.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| title | string | yes | Hero-specific title (e.g., "Gaze Constitution") |
| parent_constitution_version | semver string | yes | Which org constitution version this aligns with |
| hero_name | string | yes | Name of the hero (e.g., "Gaze") |
| version | semver string | yes | Hero constitution version |
| ratification_date | ISO 8601 date | yes | Date adopted |
| last_amended_date | ISO 8601 date | yes | Date of most recent amendment |
| principles | Core Principle[] | yes | Hero-specific principles (minimum 1) |
| governance_rules | Governance Rule[] | yes | Hero-specific governance |

**Constraint**: Hero principles MUST NOT contradict any org principle (FR-007).

**Constraint**: `parent_constitution_version` MUST reference a valid org constitution version.

**File location**: `.specify/memory/constitution.md` within each hero repository.

### Alignment Check

A validation report comparing a hero constitution against the org constitution. Produced by the OpenCode alignment agent.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| hero_name | string | yes | Name of the hero being checked |
| hero_constitution_version | semver string | yes | Version of the hero constitution |
| org_constitution_version | semver string | yes | Version of the org constitution checked against |
| checked_at | ISO 8601 datetime | yes | When the check was performed |
| findings | Alignment Finding[] | yes | One finding per org principle |
| parent_reference_status | enum | yes | PRESENT or MISSING |
| overall_status | enum | yes | ALIGNED or NON-ALIGNED |

**Status derivation**: NON-ALIGNED if any finding has status CONTRADICTION, or if parent_reference_status is MISSING.

### Alignment Finding

A single finding from the alignment check, mapping one org principle to zero or more hero principles.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| org_principle | string | yes | Org principle name (e.g., "I. Autonomous Collaboration") |
| hero_principles_mapped | string[] | yes | Hero principles that support this org principle (may be empty) |
| status | enum | yes | ALIGNED, GAP, or CONTRADICTION |
| rationale | string | yes | Explanation of the finding |

**Status values**:

- **ALIGNED**: At least one hero principle supports this org principle with no contradictions.
- **GAP**: No hero principle explicitly addresses this org principle, but no contradiction exists. The hero SHOULD consider adding coverage.
- **CONTRADICTION**: A hero principle directly contradicts a MUST rule from this org principle. This MUST be resolved before the hero is considered aligned.

## Relationships

```text
Org Constitution (1) ──────< Hero Constitution (many)
  │                              │
  │ defines                      │ references via
  │                              │ parent_constitution_version
  ▼                              ▼
Core Principle (3)         Core Principle (1+)
  │                              │
  │ checked against              │ mapped to
  ▼                              ▼
Alignment Finding (1 per org principle per check)
  │
  │ aggregated into
  ▼
Alignment Check (1 per hero per run)
```

## State Transitions

### Alignment Check Status

```text
[Start] ──> Run Agent ──> Collect Findings
                              │
                    ┌─────────┴─────────┐
                    │                     │
              All ALIGNED           Any CONTRADICTION
              + parent PRESENT      or parent MISSING
                    │                     │
                    ▼                     ▼
                ALIGNED              NON-ALIGNED
```

### Org Constitution Lifecycle

```text
[Template] ──> [Ratified v1.0.0]
                    │
                    ├──> [Amended v1.1.0] (MINOR: new guidance)
                    │         │
                    │         └──> [Amended v1.1.1] (PATCH: wording fix)
                    │
                    └──> [Amended v2.0.0] (MAJOR: principle change)
                               │
                               └──> All hero constitutions MUST
                                    be reviewed for alignment
```
