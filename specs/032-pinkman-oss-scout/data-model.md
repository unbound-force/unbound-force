# Data Model: Pinkman OSS Scout

## Entities

### Scouted Project

An open source project discovered by Pinkman during a
scouting invocation.

| Field                  | Type        | Description                                            |
|------------------------|-------------|--------------------------------------------------------|
| name                   | string      | Project name (repository name)                         |
| repository_url         | string      | Full URL to the project repository                     |
| description            | string      | Project description (from repository)                  |
| primary_language       | string      | Primary programming language                           |
| license_spdx_id        | string      | SPDX identifier for the project's license              |
| license_verdict        | enum        | OSI approval status (see License Verdict below)        |
| license_explanation    | string      | Human-readable explanation of the verdict              |
| compatibility_tier     | enum        | License compatibility tier (see below)                 |
| compatibility_verdict  | enum        | Compatibility verdict vs Apache-2.0 (see below)        |
| stars                  | integer     | Total star count at time of scouting                   |
| forks                  | integer     | Total fork count at time of scouting                   |
| star_growth_rate       | float       | Stars gained in last 90 days as % of total             |
| release_velocity       | integer     | Number of releases in last 6 months                    |
| contributor_activity   | integer     | Unique contributors with commits in last 90 days       |
| last_commit_date       | date        | Date of most recent commit                             |
| direct_dependencies    | list        | List of Dependency Reference entities                  |
| has_manifest           | boolean     | Whether a dependency manifest was detected             |
| scouted_at             | datetime    | Timestamp when this project was scouted                |

**Identity**: `repository_url` is the unique identifier.

**Validation**:
- `repository_url` MUST be a valid URL
- `license_spdx_id` MUST be a valid SPDX identifier or
  "UNKNOWN" or "CUSTOM"
- `license_verdict` MUST be one of the defined enum
  values
- `star_growth_rate` is expressed as a decimal (0.15
  = 15% growth)

### License Verdict (Enum)

| Value               | Description                                              |
|---------------------|----------------------------------------------------------|
| `approved`          | License appears on the current OSI-approved list         |
| `not_approved`      | License is known but does not appear on the OSI list     |
| `unknown`           | No license file detected in the project                  |
| `manual_review`     | Non-standard or custom license requires human review     |
| `dual_approved`     | Dual-license project; at least one option is OSI-approved |

### Compatibility Tier (Enum)

Added by opsx/pinkman-license-compatibility.

| Value              | Description                                              |
|--------------------|----------------------------------------------------------|
| `permissive`       | No derivative work obligations (MIT, Apache-2.0, BSD, ISC, etc.) |
| `weak-copyleft`    | File-level or linking-exception copyleft (LGPL, MPL-2.0, etc.) |
| `strong-copyleft`  | Full copyleft — derivative works must use same license (GPL, AGPL) |
| `unknown`          | License not in tier table, SPDX AND/WITH expression, or unrecognized |

### Compatibility Verdict (Enum)

Added by opsx/pinkman-license-compatibility.

| Value           | Description                                              |
|-----------------|----------------------------------------------------------|
| `compatible`    | No conflict with Apache-2.0 (permissive tier)            |
| `caution`       | May be compatible depending on usage; requires legal review (weak-copyleft, unknown) |
| `incompatible`  | Derivative work obligations conflict with Apache-2.0 (strong-copyleft, not_approved) |

### Compatibility Gate

The compatibility verdict caps the recommendation:

| Compatibility | Maximum recommendation |
|---------------|-----------------------|
| `compatible`  | `adopt`               |
| `caution`     | `evaluate`            |
| `incompatible`| `avoid`               |

### Dependency Reference

A direct dependency of a scouted project, extracted from
its dependency manifest.

| Field       | Type    | Description                                         |
|-------------|---------|-----------------------------------------------------|
| name        | string  | Dependency name (module path or package name)       |
| version     | string  | Version constraint from the manifest                |
| registry    | string  | Package registry (go, npm, crates, pypi)            |

**Identity**: `name` + `registry` is the composite key.

### Dependency Overlap

A dependency that appears in two or more scouted
projects within a single result set. Computed as a
post-processing step after all projects are scouted.

| Field               | Type    | Description                                    |
|---------------------|---------|------------------------------------------------|
| dependency_name     | string  | Name of the shared dependency                  |
| registry            | string  | Package registry                               |
| project_count       | integer | Number of scouted projects that use it         |
| projects            | list    | List of (project_name, version) tuples         |
| version_conflict    | boolean | True if versions differ across projects        |
| versions            | list    | Distinct versions used across projects         |

**Derivation**: Computed from the union of all
`direct_dependencies` across all Scouted Projects in a
result set. A Dependency Overlap entry is created when
`project_count >= 2`.

### Dependency Health Report

The result of auditing an existing dependency from a
local project's manifest file.

| Field                | Type    | Description                                      |
|----------------------|---------|--------------------------------------------------|
| dependency_name      | string  | Name of the dependency                           |
| current_version      | string  | Version currently used in the manifest           |
| latest_version       | string  | Latest available version from the registry       |
| update_available     | boolean | True if latest_version > current_version         |
| license_current      | string  | SPDX identifier of the current version's license |
| license_latest       | string  | SPDX identifier of the latest version's license  |
| license_changed      | boolean | True if license differs between versions         |
| license_still_osi    | boolean | True if latest version license is OSI-approved   |
| maintenance_risk     | enum    | Risk level (healthy, warning, critical)          |
| risk_indicators      | list    | Specific risk signals (see below)                |
| last_commit_date     | date    | Date of most recent commit to the dependency     |
| repository_archived  | boolean | True if the repository is archived               |
| owner_changed        | boolean | True if the repository owner has changed         |

**Maintenance Risk Enum**:
- `healthy`: Active commits within 6 months, responsive
  issue triage, regular releases
- `warning`: No commits in 6-12 months, or owner change
  detected, or issue backlog growing
- `critical`: No commits in 12+ months, repository
  archived, or license changed to non-OSI-approved

**Risk Indicators** (list of strings):
- `"no_commits_12m"`: No commits in 12+ months
- `"no_commits_6m"`: No commits in 6-12 months
- `"archived"`: Repository is archived
- `"owner_changed"`: Repository transferred to new owner
- `"license_changed"`: License changed between versions
- `"license_not_osi"`: New license is not OSI-approved
- `"issues_growing"`: Open issue count growing with no
  resolution trend

### Adoption Recommendation

A structured report for a specific project under
evaluation.

| Field                  | Type       | Description                                     |
|------------------------|------------|-------------------------------------------------|
| project                | ref        | Reference to Scouted Project entity             |
| license_analysis       | section    | License verdict with explanation                |
| community_health       | section    | Star/fork/contributor metrics with assessment   |
| maintenance_signals    | section    | Risk indicators and maintenance health          |
| trend_trajectory       | section    | Trend indicators with historical context        |
| dependency_list        | list       | Direct dependencies of the project              |
| dependency_overlap     | list       | Shared dependencies with other evaluated projects |
| existing_dep_relation  | section    | Relationship to current Unbound Force deps      |
| recommendation         | enum       | Overall recommendation (see below)              |
| recommendation_reason  | string     | Human-readable justification                    |

**Recommendation Enum**:
- `adopt`: OSI-approved, healthy, trending, no conflicts
- `evaluate`: OSI-approved but has concerns (maintenance
  risk, low trend signals)
- `defer`: OSI-approved but significant concerns
  (critical maintenance risk, license instability)
- `avoid`: Not OSI-approved, or critical supply chain
  risks

### Scouting Report (YAML Frontmatter Schema)

The YAML frontmatter for reports stored at
`.uf/pinkman/reports/`.

```yaml
---
producer: pinkman
version: "1.0.0"
timestamp: "2026-04-22T14:30:00Z"
query: "static analysis Go"
mode: "discover"          # discover | trend | audit | report
result_count: 12
compatible_count: 8
incompatible_count: 3
unknown_count: 1
overlap_count: 5
sources_consulted:
  - github.com
  - pkg.go.dev
sources_failed: []
fallback_license_list: false
---
```

## Relationships

```text
Scouted Project
  ├── 1:N → Dependency Reference (direct_dependencies)
  ├── 1:1 → License Verdict (enum field)
  └── 0:1 → Adoption Recommendation (optional)

Dependency Overlap
  └── N:M → Scouted Project (projects list)

Dependency Health Report
  └── 1:1 → local manifest entry (audit mode)
```

## State Transitions

Scouted Projects do not have lifecycle states -- they
are point-in-time snapshots. Each scouting invocation
produces a new set of Scouted Projects. Historical
awareness is achieved through Dewey integration (query
past scouting reports for the same project URL).

Dependency Health Reports are similarly stateless --
each audit produces a fresh report. Trend detection
across audits ("this dependency's risk level increased
since last month") is a future enhancement.
