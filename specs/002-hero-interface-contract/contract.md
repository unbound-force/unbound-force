# Hero Interface Contract

**Version**: 1.0.0
**Date**: 2026-02-25
**Authority**: Unbound Force Org Constitution v1.0.0
**Related Specs**: Spec 001 (Org Constitution), Spec 003
(Speckit Framework), Spec 009 (Shared Data Model)

## Preamble

The Hero Interface Contract defines the minimum requirements
every repository MUST satisfy to be a member of the Unbound
Force swarm. It is the "plug shape" that ensures
interoperability — any repository that conforms to this
contract can produce and consume artifacts, expose OpenCode
agents, and be discovered by swarm tooling.

This contract derives its authority from the Unbound Force
Org Constitution (Spec 001, v1.0.0). All requirements in this
contract trace to one or more of the three constitutional
principles:

- **Principle I (Autonomous Collaboration)**: Heroes
  communicate through well-defined artifacts, not runtime
  coupling. This contract defines the artifact envelope format
  and the artifact type registry.
- **Principle II (Composability First)**: Heroes are
  independently installable and usable alone. This contract
  defines the repository structure and hero manifest that
  enable standalone operation.
- **Principle III (Observable Quality)**: Heroes produce
  machine-parseable output with provenance metadata. This
  contract requires JSON minimum output and defines the
  provenance fields in the artifact envelope.

### Scope

This contract governs all hero repositories in the
`unbound-force` GitHub organization, including:

- Hero repositories: Gaze, Muti-Mind, Cobalt-Crush, The
  Divisor, Mx F
- Supporting repositories: Website
- Future hero repositories added to the organization

The contract does NOT govern the meta repository
(`unbound-force/unbound-force`) itself, as it is not a hero.
However, the meta repository SHOULD comply with the structural
requirements where applicable.

### Relationship to Other Specs

- **Spec 001 (Org Constitution)**: The highest-authority
  document. This contract implements the constitutional
  principles as concrete, enforceable requirements.
- **Spec 003 (Speckit Framework)**: Defines the distribution
  mechanism for the speckit framework. This contract requires
  heroes to use speckit (Section 4) but defers the how to
  Spec 003.
- **Spec 009 (Shared Data Model)**: Defines the formal JSON
  Schemas for the artifact envelope and all artifact type
  payloads. This contract defines the envelope conceptually
  (Section 3) — Spec 009 provides the machine-readable
  schemas.

### Terminology

This contract uses RFC 2119 keywords (MUST, MUST NOT, SHOULD,
SHOULD NOT, MAY) to indicate requirement levels. The key words
are to be interpreted as described in RFC 2119.

---

## 1. Hero Lifecycle

Every hero repository follows a defined lifecycle from
creation to ongoing maintenance. Contract requirements apply
at specific lifecycle stages.

### Lifecycle Stages

1. **Bootstrap**: Create the repository with the required
   directory structure (Section 2). Create the hero manifest
   at `.unbound-force/hero.json` (Section 6). Install speckit
   from the canonical source (Section 4).

2. **Constitution**: Ratify the hero constitution at
   `.specify/memory/constitution.md`. The constitution MUST
   include a `parent_constitution` field referencing the org
   constitution version (see Section 2.1). Run
   `/constitution-check` to verify alignment with the org
   constitution.

3. **Specify**: Create the first feature specification using
   the speckit pipeline: constitution -> specify -> clarify ->
   plan -> tasks -> analyze -> checklist -> implement.

4. **Implement**: Build the hero's functionality. Heroes that
   produce output MUST support at minimum JSON format (FR-010).
   Heroes that produce artifacts for inter-hero consumption
   MUST use the artifact envelope format (Section 3).

5. **Deploy**: Release the hero for use. Follow semantic
   versioning (MAJOR.MINOR.PATCH). Update the hero manifest
   version field on each release.

6. **Maintain**: Update the hero manifest when adding new
   artifacts, agents, or commands. Re-validate contract
   compliance when the contract is updated. Review
   constitution alignment when the org constitution is amended.

---

## 2. Repository Structure

Every hero repository MUST contain the following elements.
Elements marked MUST are required for contract compliance.
Elements marked SHOULD are recommended but not blocking.

### Required Elements (MUST)

| Element | Type | Purpose |
|---------|------|---------|
| `.specify/memory/constitution.md` | file | Hero constitution, ratified, with `parent_constitution` reference |
| `.specify/templates/` | dir | Speckit templates (6 files from canonical source) |
| `.specify/scripts/bash/` | dir | Speckit automation scripts (5 files from canonical source) |
| `.opencode/` | dir | OpenCode configuration directory |
| `.opencode/command/` | dir | OpenCode commands (speckit pipeline + hero-specific) |
| `specs/` | dir | Feature specifications directory |
| `AGENTS.md` | file | Agent context file with project overview, tech stack, and conventions |
| `LICENSE` | file | License file (Apache 2.0 recommended; see Section 2.2) |
| `README.md` | file | Project README (see Section 2.3 for template) |
| `.unbound-force/hero.json` | file | Hero manifest (see Section 6) |

### Optional Elements (SHOULD)

| Element | Type | Purpose |
|---------|------|---------|
| `.opencode/agents/` | dir | OpenCode agent definitions (if hero provides agents) |
| `.github/workflows/` | dir | CI/CD workflows |
| `.specify/config.yaml` | file | Speckit project-specific configuration (per Spec 003) |

### Validation

A repository is contract-compliant when all required elements
are present and the hero manifest validates against the hero
manifest JSON Schema. The validation script
(`scripts/validate-hero-contract.sh`) automates this check.

### 2.1 Constitution Requirements

Hero constitutions MUST satisfy these requirements:

- The constitution MUST include a `parent_constitution`
  field referencing the org constitution version the hero
  aligns with. Format: `**Parent Constitution**: Unbound
  Force v{MAJOR}.{MINOR}.{PATCH}`
- The constitution MUST NOT contradict any org-level MUST
  rule from the org constitution.
- The constitution MUST be ratified (not a draft template)
  before the hero is considered contract-compliant.
- Hero constitutions MAY add principles beyond the three
  org principles, provided the additional principles do not
  contradict any org-level MUST rule.
- Hero constitutions that predate the org constitution MUST
  be reviewed for alignment. Contradictions MUST be resolved
  by amending the hero constitution.

### 2.2 License Requirements

- Apache 2.0 is RECOMMENDED for all hero repositories.
- Alternative licenses MAY be used with explicit justification
  documented in the hero's constitution.
- The `LICENSE` file MUST be present at the repository root
  regardless of the license chosen.

### 2.3 README Template

Hero repositories SHOULD follow this README structure:

```markdown
# {Hero Display Name}

{One-line description of the hero's purpose}

## Role

{Hero role}: {tester|developer|reviewer|product-owner|manager}

## Installation

{How to install the hero}

## Quick Start

{Minimal steps to use the hero}

## Integration with Other Heroes

{How this hero interacts with other heroes in the swarm}

## Links

- [Unbound Force Organization](https://github.com/unbound-force)
- [Org Constitution]({link to org constitution})
- [Hero Interface Contract]({link to this contract})
```

---

## 3. Artifact Envelope

Heroes exchange information through well-defined artifacts
rather than runtime coupling (Principle I). Every artifact
produced for inter-hero consumption MUST be wrapped in the
standard artifact envelope.

### Envelope Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `hero` | string | MUST | Name of the producing hero (e.g., "gaze") |
| `version` | semver string | MUST | Version of the producing hero |
| `timestamp` | ISO 8601 string | MUST | When the artifact was produced |
| `artifact_type` | string | MUST | Registered artifact type (e.g., "quality-report") |
| `schema_version` | semver string | MUST | Version of the payload schema |
| `context` | object | MUST | Execution context metadata (see below) |
| `payload` | object | MUST | Type-specific content |

### Context Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `branch` | string | MUST | Git branch the artifact was produced against |
| `commit` | string | SHOULD | Git commit SHA |
| `backlog_item_id` | string | MAY | Originating backlog item (from Muti-Mind) |
| `correlation_id` | UUID string | MAY | Links related artifacts across workflow stages |

### Example

```json
{
  "hero": "gaze",
  "version": "1.2.0",
  "timestamp": "2026-02-25T14:30:00Z",
  "artifact_type": "quality-report",
  "schema_version": "1.0.0",
  "context": {
    "branch": "main",
    "commit": "abc123def"
  },
  "payload": { }
}
```

### Formal Schema

The formal JSON Schema for the artifact envelope is defined
by Spec 009 (Shared Data Model). This contract establishes
the conceptual definition — which fields exist, what they
mean, and what rules govern them. Spec 009 translates this
into a machine-readable JSON Schema (draft 2020-12).

### 3.1 Artifact Type Registry

The following artifact types are registered for inter-hero
communication. Each type has a designated producer and one
or more consumers. JSON Schemas for each type's payload are
defined by Spec 009 (Shared Data Model).

| Type ID | Producer | Consumers | Description |
|---------|----------|-----------|-------------|
| `quality-report` | Gaze | Mx F, Muti-Mind, Cobalt-Crush | Code quality analysis with coverage, complexity, and side-effect metrics |
| `review-verdict` | The Divisor | Mx F, Cobalt-Crush, Muti-Mind | PR review decision from the three-persona council |
| `backlog-item` | Muti-Mind | Mx F, Cobalt-Crush | Prioritized work item with acceptance criteria |
| `acceptance-decision` | Muti-Mind | Mx F, Cobalt-Crush | Accept/reject decision on completed work |
| `metrics-snapshot` | Mx F | Muti-Mind | Sprint and project health metrics |
| `coaching-record` | Mx F | All heroes | Retrospective patterns and coaching interactions |
| `workflow-record` | Swarm Orchestration | Mx F, Muti-Mind | End-to-end feature lifecycle record |

To register a new artifact type, a hero MUST submit a pull
request to the shared data model (Spec 009) with:

- A unique `type_id` (no duplicates allowed)
- A JSON Schema for the payload (draft 2020-12)
- At least one sample artifact
- Documentation of producer and consumer heroes

### 3.2 Artifact Versioning

Artifact schemas follow semantic versioning
(MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking change to the payload structure (field
  removed, field type changed, required field added).
  Consumers expecting the previous MAJOR version MUST NOT
  be expected to parse the new version without migration.
  A migration guide MUST be provided.
- **MINOR**: Backward-compatible addition (new optional
  field). Consumers expecting the previous MINOR version
  MUST still parse the artifact successfully (new fields
  are ignored).
- **PATCH**: Documentation fix or non-functional change. No
  impact on parsing.

### 3.3 Producing and Consuming Artifacts

**Producer obligations**:

- Producers MUST conform to the registered schema for their
  artifact type.
- Producers MUST include all required envelope fields.
- Producers MUST use registered `artifact_type` values. Using
  unregistered types violates the contract.
- Producers MUST include provenance metadata (hero, version,
  timestamp) for traceability (Principle III).

**Consumer obligations**:

- Consumers MUST handle valid instances of artifact types
  they are registered to consume.
- Consumers MUST ignore unrecognized `artifact_type` values
  gracefully — no errors, no crashes. A warning log entry
  SHOULD be emitted.
- Consumers MUST check the `schema_version` field before
  parsing the payload:
  - Same MAJOR version: proceed (minor/patch differences are
    backward compatible).
  - Different MAJOR version: degrade gracefully or reject
    with a version mismatch warning. Consumers MUST NOT
    crash on version mismatches.

---

## 4. Speckit Integration

Every hero repository MUST use the speckit framework for
specification-driven development. The speckit framework
provides templates, scripts, and OpenCode commands that
standardize the development workflow across all heroes.

### Canonical Source

The canonical source for speckit files is defined by Spec 003
(Speckit Framework Centralization). Hero repositories MUST
install speckit from this canonical source, NOT by copy-pasting
files from other hero repositories.

The canonical file inventory:

- **Templates** (6 files in `.specify/templates/`):
  `spec-template.md`, `plan-template.md`,
  `tasks-template.md`, `checklist-template.md`,
  `constitution-template.md`, `agent-file-template.md`
- **Scripts** (5 files in `.specify/scripts/bash/`):
  `common.sh`, `check-prerequisites.sh`, `setup-plan.sh`,
  `create-new-feature.sh`, `update-agent-context.sh`
- **Commands** (9+ files in `.opencode/command/`):
  `speckit.constitution.md`, `speckit.specify.md`,
  `speckit.clarify.md`, `speckit.plan.md`,
  `speckit.tasks.md`, `speckit.analyze.md`,
  `speckit.checklist.md`, `speckit.implement.md`,
  `speckit.taskstoissues.md`

Where drift exists between hero repos and the org repo, the
`unbound-force` (org repo) versions are canonical.

### 4.1 Extension Points

Project-specific customizations MUST be handled via
`.specify/config.yaml` configuration (per Spec 003), NOT by
modifying canonical speckit files.

Heroes MUST NOT modify canonical speckit files to add
project-specific behavior. Modifications to canonical files
create drift and prevent clean upgrades.

Known drift points that MUST be resolved via configuration:

- `speckit.specify.md`: Integration pattern language
  (org repo uses "project-appropriate patterns"; Gaze uses
  "RESTful APIs") — resolved via `integration_patterns`
  config field.
- `speckit.plan.md`: Interface contract generation (org repo
  uses "interface contracts"; Gaze uses "API contracts") —
  resolved via `project_type` config field.
- `speckit.tasks.md`: Terminology (org repo uses "interfaces";
  Gaze uses "endpoints") — resolved via `project_type`
  config field.

### 4.2 Update Protocol

When the speckit framework is updated:

- The update tool MUST detect local modifications to canonical
  files (via content checksum comparison).
- Modified files MUST be skipped with a warning listing the
  conflicting files.
- New files (added in the newer speckit version) MUST be
  installed even when other files are skipped.
- The update tool MUST NOT overwrite modified files without
  explicit `--force` flag.
- After update, `.specify/speckit.version` MUST be updated
  to reflect the new version.

---

## 5. OpenCode Agent and Command Standards

Heroes that provide OpenCode agents or commands MUST follow
the naming conventions and format standards defined in this
section to prevent collisions and ensure consistency across
the swarm.

### Agent Naming Convention

Agent files MUST be named using the pattern:

```text
{hero-name}-{agent-function}.md
```

- `{hero-name}`: The hero's name in lowercase kebab-case
  (e.g., `gaze`, `divisor`, `muti-mind`, `cobalt-crush`,
  `mx-f`)
- `{agent-function}`: A descriptive function name in
  lowercase kebab-case (e.g., `reporter`, `guard`,
  `architect`, `po`)

Examples:

```text
gaze-reporter.md         # Gaze's reporting agent
divisor-guard.md         # The Divisor's Guard persona
divisor-architect.md     # The Divisor's Architect persona
divisor-adversary.md     # The Divisor's Adversary persona
muti-mind-po.md          # Muti-Mind's Product Owner agent
cobalt-crush-dev.md      # Cobalt-Crush's developer agent
mx-f-coach.md            # Mx F's coaching agent
```

The hero prefix MUST be present to prevent collisions when
multiple heroes install agents into the same project's
`.opencode/agents/` directory.

### Agent File Format

Every agent file MUST include these sections:

1. **Descriptive header**: Agent name, purpose, and hero
   that provides it.
2. **Tool permissions**: Explicit list of allowed tools
   (e.g., `read`, `edit`, `bash`, `glob`, `grep`).
3. **Model specification**: Which LLM model the agent
   should use (if agent-specific).
4. **Behavioral constraints**: Rules and boundaries for the
   agent's behavior.

### Command Naming Convention

Command files MUST be named using the pattern:

```text
{hero-name}.md           # Primary command
{hero-name}-{sub}.md     # Secondary commands
```

Command triggers follow the pattern:

```text
/{hero-name}              # Primary command trigger
/{hero-name}-{sub}        # Secondary command triggers
```

Examples:

```text
/gaze                     # Gaze's main command
/classify-docs            # Gaze's doc classification
/review-council           # The Divisor's review council
/muti-mind                # Muti-Mind's main command
```

### Command File Format

Every command file MUST define:

1. **Trigger syntax**: The command name and argument format.
2. **Argument parsing**: How arguments are extracted and
   validated.
3. **Agent delegation**: Which agent (if any) the command
   delegates to.
4. **Error handling**: How errors are reported to the user.

### 5.1 Init Command Convention

Heroes SHOULD provide an `init` command that scaffolds the
hero's OpenCode integration files into a target project:

```text
{hero-name} init [options]
```

The init command SHOULD:

- Create agent files in `.opencode/agents/`
- Create command files in `.opencode/command/`
- Detect existing files and warn on collision
- Support a `--force` flag to overwrite existing files
- Report which files were created, skipped, or overwritten

If a collision occurs during init (an agent or command file
with the same name already exists), the init command MUST
warn the user and skip the file unless `--force` is used.

### 5.2 MCP Server Conventions

Heroes that expose capabilities via MCP (Model Context
Protocol) SHOULD follow these naming conventions:

- **Server name**: `{hero-name}-mcp` (e.g., `gaze-mcp`,
  `divisor-mcp`)
- **Tool names**: `{hero-name}_{tool_name}` (e.g.,
  `gaze_analyze`, `gaze_quality`)

Full MCP interface requirements (request/response schemas,
capability negotiation, authentication) are deferred to
hero-specific architecture specs (Specs 004-007).

---

## 6. Hero Manifest

Every hero repository MUST include a machine-readable
manifest at `.unbound-force/hero.json` that describes the
hero's identity, capabilities, and integration points. The
manifest enables discovery tools (Swarm plugin, Mx F
metrics platform) to automatically detect and configure
heroes.

### Manifest Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | MUST | Hero name in lowercase kebab-case (e.g., "gaze") |
| `display_name` | string | MUST | Human-readable name (e.g., "Gaze") |
| `role` | enum | MUST | One of: `tester`, `developer`, `reviewer`, `product-owner`, `manager` |
| `version` | semver string | MUST | Current hero version (e.g., "1.2.0") |
| `description` | string | MUST | One-line description of the hero's purpose |
| `repository` | URL string | MUST | GitHub repository URL |
| `parent_constitution_version` | semver string | MUST | Org constitution version this hero aligns with |
| `artifacts_produced` | array | MUST | Artifact types this hero produces (may be empty `[]`) |
| `artifacts_consumed` | array | MUST | Artifact types this hero consumes (may be empty `[]`) |
| `opencode_agents` | array | MUST | OpenCode agents this hero provides (may be empty `[]`) |
| `opencode_commands` | array | MUST | OpenCode commands this hero provides (may be empty `[]`) |
| `dependencies` | array | MUST | Other heroes this hero integrates with (may be empty `[]`; all are optional per Principle II) |
| `mcp_server` | object | MAY | MCP server metadata if the hero exposes MCP capabilities |

### Sub-Entity: Artifact Reference

Each entry in `artifacts_produced` or `artifacts_consumed`:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type_id` | string | MUST | Artifact type ID (e.g., "quality-report") |
| `schema_version` | semver string | MUST | Schema version supported |
| `description` | string | MAY | Brief description |

### Sub-Entity: Agent Reference

Each entry in `opencode_agents`:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | MUST | Agent file name without `.md` extension |
| `description` | string | MUST | What the agent does |
| `mode` | enum | MUST | One of: `agent`, `subagent` |
| `tools` | array of strings | MUST | Tool permissions (e.g., `["read", "bash"]`) |

### Sub-Entity: Command Reference

Each entry in `opencode_commands`:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | MUST | Command trigger (e.g., "/gaze") |
| `description` | string | MUST | What the command does |
| `agent` | string | MAY | Agent the command delegates to |

### Sub-Entity: Dependency

Each entry in `dependencies`. All dependencies are optional
per Principle II — a hero MUST function without any of its
declared dependencies being present.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `hero_name` | string | MUST | Name of the hero this integrates with |
| `integration_type` | enum | MUST | One of: `artifact_consumer`, `artifact_producer`, `peer` |
| `artifact_types` | array of strings | MAY | Which artifact types are exchanged |
| `description` | string | MAY | How the integration works |

### Sub-Entity: MCP Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `server_name` | string | MUST | MCP server name (convention: `{hero-name}-mcp`) |
| `version` | semver string | MUST | MCP server version |
| `tools` | array of strings | MUST | List of MCP tool names exposed |

### Manifest Schema

The hero manifest MUST validate against the JSON Schema at
`schemas/hero-manifest/v1.0.0.schema.json`. This schema
enforces all required fields, type constraints, and enum
values defined above.

### Example

```json
{
  "name": "gaze",
  "display_name": "Gaze",
  "role": "tester",
  "version": "1.2.0",
  "description": "Static analysis tool for Go",
  "repository": "https://github.com/unbound-force/gaze",
  "parent_constitution_version": "1.0.0",
  "artifacts_produced": [
    {
      "type_id": "quality-report",
      "schema_version": "1.0.0",
      "description": "Code quality analysis with coverage"
    }
  ],
  "artifacts_consumed": [],
  "opencode_agents": [
    {
      "name": "gaze-reporter",
      "description": "Generates quality reports",
      "mode": "agent",
      "tools": ["read", "bash"]
    }
  ],
  "opencode_commands": [
    {
      "name": "/gaze",
      "description": "Run Gaze analysis",
      "agent": "gaze-reporter"
    }
  ],
  "dependencies": []
}
```

---

## 7. Version Compatibility

Heroes MUST handle version incompatibilities gracefully when
consuming artifacts from other heroes. Hard failures on
version mismatches violate Principle II (Composability First).

### Rules

1. **Same MAJOR version**: Compatible. Minor and patch
   differences are backward compatible. Consumers MUST
   proceed without error.
2. **Different MAJOR version**: Incompatible. Consumers
   MUST detect the mismatch via the `schema_version` field
   and either:
   - Apply a migration if one is available, or
   - Reject the artifact with a clear warning message
     ("incompatible schema version: expected v1.x.x,
     got v2.0.0").
   - Consumers MUST NOT crash or produce undefined behavior
     on version mismatches.
3. **Unknown artifact type**: If a hero encounters an
   artifact with an `artifact_type` it does not recognize,
   it MUST ignore it gracefully. A warning log entry SHOULD
   be emitted but the hero MUST NOT error or halt.

### Contract Versioning

This contract itself follows semantic versioning:

- **MAJOR**: Requirement removal or incompatible redefinition
  of a MUST rule.
- **MINOR**: New requirement added or existing SHOULD
  elevated to MUST.
- **PATCH**: Clarifications, examples, or non-functional
  wording changes.

---

## 8. Validation

Contract compliance is verified by the validation script at
`scripts/validate-hero-contract.sh`. The script accepts a
repository path and checks all required elements.

### Running Validation

```bash
bash scripts/validate-hero-contract.sh /path/to/hero-repo
```

### Check Categories

- **Required checks** (MUST pass): File and directory
  existence, constitution `parent_constitution` reference,
  hero manifest existence and validity.
- **Optional checks** (warnings only): CI workflows,
  OpenCode agents directory, config.yaml.

### Output Format

The script outputs structured results:

```text
Hero Interface Contract Validation
===================================
Repository: /path/to/hero-repo
Contract version: 1.0.0

Required Checks:
  [PASS] .specify/memory/constitution.md exists
  [FAIL] .unbound-force/hero.json exists

Optional Checks:
  [WARN] .opencode/agents/ not found

Overall: FAIL (12/13 required, 1/2 optional)
```

### Exit Codes

- `0`: All required checks pass (PASS).
- `1`: One or more required checks fail (FAIL).

---

## 9. Migration Guide

Hero repositories that predate this contract MUST be updated
to comply. The following migration steps are recommended:

1. **Create `.unbound-force/hero.json`**: Add the hero
   manifest with all required fields populated.
2. **Add `parent_constitution` reference**: Update the hero
   constitution to include the parent reference to the org
   constitution version.
3. **Rename non-compliant agents**: Rename agent files that
   do not follow the `{hero-name}-{agent-function}.md`
   convention.
4. **Run validation**: Execute the validation script and
   address any remaining failures.
5. **Adopt artifact envelope**: When Spec 009 provides the
   formal schemas, wrap hero output in the standard envelope
   format. This is a separate migration tracked per hero.

### Known Remediation Items

**Gaze**:
- Create `.unbound-force/hero.json`
- Add `parent_constitution` reference to constitution
- Rename `doc-classifier.md` to `gaze-doc-classifier.md`
- Adopt artifact envelope (deferred to Spec 009)

**Website**:
- Create `.unbound-force/hero.json`
- Add `parent_constitution` reference to constitution
