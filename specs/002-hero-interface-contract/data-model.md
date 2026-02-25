# Data Model: Hero Interface Contract

**Spec**: [spec.md](spec.md) | **Date**: 2026-02-25

## Entities

### Hero Repository Structure

The minimum required layout for a valid Unbound Force hero
repository.

| Element | Type | Required | Description |
|---------|------|----------|-------------|
| `.specify/memory/constitution.md` | file | yes | Hero constitution, ratified, with `parent_constitution` reference |
| `.specify/templates/` | dir | yes | Speckit templates (6 files from canonical source) |
| `.specify/scripts/bash/` | dir | yes | Speckit automation scripts (5 files from canonical source) |
| `.opencode/` | dir | yes | OpenCode configuration directory |
| `.opencode/command/` | dir | yes | OpenCode commands (speckit pipeline + hero-specific) |
| `specs/` | dir | yes | Feature specifications directory |
| `AGENTS.md` | file | yes | Agent context file with project overview, tech stack, and conventions |
| `LICENSE` | file | yes | License file (Apache 2.0 recommended) |
| `README.md` | file | yes | Project README following the hero README template |
| `.unbound-force/hero.json` | file | yes | Hero manifest (machine-readable metadata) |
| `.opencode/agents/` | dir | optional | OpenCode agent definitions (if hero provides agents) |
| `.github/workflows/` | dir | optional | CI/CD workflows |
| `.specify/config.yaml` | file | optional | Speckit project-specific configuration (per Spec 003) |

**Validation**: A repository is contract-compliant when all
required elements are present and the hero manifest validates
against the hero manifest schema.

### Hero Manifest

Machine-readable description of a hero's identity, capabilities,
and integration points. Located at `.unbound-force/hero.json`.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | yes | Hero name (e.g., "gaze", "the-divisor") |
| display_name | string | yes | Human-readable name (e.g., "Gaze", "The Divisor") |
| role | enum | yes | One of: tester, developer, reviewer, product-owner, manager |
| version | semver string | yes | Current hero version |
| description | string | yes | One-line description of the hero's purpose |
| repository | URL string | yes | GitHub repository URL |
| parent_constitution_version | semver string | yes | Org constitution version this hero aligns with |
| artifacts_produced | Artifact Reference[] | yes | Artifact types this hero produces (may be empty) |
| artifacts_consumed | Artifact Reference[] | yes | Artifact types this hero consumes (may be empty) |
| opencode_agents | Agent Reference[] | yes | OpenCode agents this hero provides (may be empty) |
| opencode_commands | Command Reference[] | yes | OpenCode commands this hero provides (may be empty) |
| dependencies | Dependency[] | yes | Other heroes this hero can integrate with (may be empty; all are optional per Principle II) |
| mcp_server | MCP Reference | optional | MCP server metadata if the hero exposes MCP capabilities |

**Validation**: The hero manifest MUST validate against the
hero manifest JSON Schema (`schemas/hero-manifest/v1.0.0.schema.json`).

**Lifecycle**: Created during hero bootstrapping. Updated when
the hero adds new artifacts, agents, commands, or bumps its
version.

### Artifact Reference

A reference to an artifact type produced or consumed by a hero.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type_id | string | yes | Artifact type identifier (e.g., "quality-report") |
| schema_version | semver string | yes | Schema version supported (e.g., "1.0.0") |
| description | string | optional | Brief description of what this artifact contains |

### Agent Reference

A reference to an OpenCode agent provided by a hero.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | yes | Agent file name without extension (e.g., "gaze-reporter") |
| description | string | yes | What the agent does |
| mode | enum | yes | One of: agent, subagent |
| tools | string[] | yes | Tool permissions (e.g., ["read", "bash"]) |

### Command Reference

A reference to an OpenCode command provided by a hero.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | yes | Command trigger (e.g., "/gaze", "/review-council") |
| description | string | yes | What the command does |
| agent | string | optional | Agent the command delegates to (if any) |

### Dependency

A declaration that this hero can integrate with another hero.
All dependencies are optional per Principle II — a hero MUST
function without any of its declared dependencies being present.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| hero_name | string | yes | Name of the hero this integrates with |
| integration_type | enum | yes | One of: artifact_consumer, artifact_producer, peer |
| artifact_types | string[] | optional | Which artifact types are exchanged |
| description | string | optional | How the integration works |

### MCP Reference

Metadata for a hero's MCP server, if it exposes one.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| server_name | string | yes | MCP server name (convention: `{hero-name}-mcp`) |
| version | semver string | yes | MCP server version |
| tools | string[] | yes | List of MCP tool names exposed |

### Artifact Envelope (Conceptual)

Standard wrapper for all inter-hero artifacts. This is the
conceptual definition; the formal JSON Schema is produced by
Spec 009 (Shared Data Model).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| hero | string | yes | Name of the producing hero (e.g., "gaze") |
| version | semver string | yes | Version of the producing hero |
| timestamp | ISO 8601 string | yes | When the artifact was produced |
| artifact_type | string | yes | Registered artifact type (e.g., "quality-report") |
| schema_version | semver string | yes | Version of the payload schema |
| context | Context object | yes | Execution context metadata |
| payload | object | yes | Type-specific content (validated by the type schema) |

### Context Object

Execution context included in every artifact envelope.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| branch | string | yes | Git branch the artifact was produced against |
| commit | string | optional | Git commit SHA |
| backlog_item_id | string | optional | Originating backlog item (from Muti-Mind) |
| correlation_id | UUID string | optional | Links related artifacts across workflow stages |

### Artifact Type Registration

An entry in the shared artifact type registry, documenting
which heroes produce and consume each type.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type_id | string | yes | Unique artifact type identifier |
| description | string | yes | What this artifact type represents |
| producing_heroes | string[] | yes | Heroes that produce this type |
| consuming_heroes | string[] | yes | Heroes that consume this type |
| current_schema_version | semver string | yes | Latest schema version |

**Registry** (initial registration):

| type_id | Producer | Consumers |
|---------|----------|-----------|
| quality-report | Gaze | Mx F, Muti-Mind, Cobalt-Crush |
| review-verdict | The Divisor | Mx F, Cobalt-Crush, Muti-Mind |
| backlog-item | Muti-Mind | Mx F, Cobalt-Crush |
| acceptance-decision | Muti-Mind | Mx F, Cobalt-Crush |
| metrics-snapshot | Mx F | Muti-Mind |
| coaching-record | Mx F | All heroes |
| workflow-record | Swarm Orchestration | Mx F, Muti-Mind |

### Validation Result

Output of the contract compliance validation script.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| hero_name | string | yes | Repository being validated |
| contract_version | semver string | yes | Version of the contract being checked against |
| checked_at | ISO 8601 string | yes | When the validation ran |
| required_checks | Check Result[] | yes | Results for required elements |
| optional_checks | Check Result[] | yes | Results for optional elements |
| overall_status | enum | yes | PASS (all required pass) or FAIL |

### Check Result

A single validation check outcome.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| check_id | string | yes | Identifier (e.g., "constitution-exists") |
| description | string | yes | What was checked |
| status | enum | yes | PASS, FAIL, or WARN |
| detail | string | optional | Additional context (e.g., missing file path) |

## Relationships

```text
Hero Repository (1)
  │
  ├── contains ──> Hero Manifest (1)
  │                  │
  │                  ├── declares ──> Artifact Reference (0..n)
  │                  │                  │
  │                  │                  └── references ──>
  │                  │                     Artifact Type Registration
  │                  │
  │                  ├── declares ──> Agent Reference (0..n)
  │                  ├── declares ──> Command Reference (0..n)
  │                  ├── declares ──> Dependency (0..n)
  │                  └── optionally ──> MCP Reference (0..1)
  │
  ├── produces ──> Artifact Envelope (0..n)
  │                  │
  │                  ├── wraps ──> Payload (typed by
  │                  │             Artifact Type Registration)
  │                  │
  │                  └── includes ──> Context Object (1)
  │
  └── validated by ──> Validation Result (1 per run)
                         │
                         └── contains ──> Check Result (1..n)
```

## State Transitions

### Hero Repository Lifecycle

```text
[Empty Repo]
    │
    ▼
[Bootstrap] ──> Create directory structure (FR-001)
    │           Create .unbound-force/hero.json (FR-002)
    │
    ▼
[Constitution] ──> Ratify hero constitution (FR-009)
    │               Add parent_constitution reference
    │
    ▼
[Specify] ──> Create first spec via speckit pipeline
    │
    ▼
[Implement] ──> Build hero functionality
    │
    ▼
[Deploy] ──> Release hero for use
    │
    ▼
[Maintain] ──> Update manifest on version bumps
               Re-validate on contract updates
               Adopt new artifact envelope when
               Spec 009 is finalized
```

### Validation Result Status

```text
[Start] ──> Run Script ──> Check Each Element
                              │
                    ┌─────────┴─────────┐
                    │                     │
              All Required           Any Required
              Checks PASS            Check FAILS
                    │                     │
                    ▼                     ▼
                  PASS                  FAIL
            (optional warns       (report missing
             still allowed)        elements)
```

### Artifact Version Compatibility

```text
Consumer expects v1.x.x
    │
    ├── Receives v1.0.0 ──> COMPATIBLE (exact match)
    │
    ├── Receives v1.1.0 ──> COMPATIBLE (minor bump,
    │                        new optional fields ignored)
    │
    ├── Receives v1.1.1 ──> COMPATIBLE (patch, no
    │                        functional change)
    │
    └── Receives v2.0.0 ──> INCOMPATIBLE (major bump,
                             consumer MUST detect via
                             schema_version and degrade
                             gracefully or reject)
```
