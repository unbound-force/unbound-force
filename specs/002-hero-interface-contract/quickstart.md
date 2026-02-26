# Quickstart: Hero Interface Contract

**Spec**: [spec.md](spec.md) | **Date**: 2026-02-25

## What is the Hero Interface Contract?

The Hero Interface Contract defines the minimum requirements
every hero repository must satisfy to be a member of the
Unbound Force swarm. It is the "plug shape" that ensures
interoperability — any repository that conforms to the contract
can produce and consume artifacts, expose OpenCode agents, and
be discovered by swarm tooling.

The contract covers five areas:

1. **Repository structure** — Required files and directories
2. **Artifact envelope** — Standard format for inter-hero
   communication
3. **OpenCode conventions** — Agent and command naming
4. **Speckit integration** — Shared development framework
5. **Hero manifest** — Machine-readable metadata for discovery

## Bootstrapping a New Hero Repository

### 1. Create the repository structure

```text
my-hero/
├── .specify/
│   ├── memory/
│   │   └── constitution.md     # Hero constitution
│   ├── templates/              # Speckit templates (from
│   │                             canonical source)
│   └── scripts/bash/           # Speckit scripts (from
│                                 canonical source)
├── .opencode/
│   └── command/                # Speckit pipeline commands
├── .unbound-force/
│   └── hero.json               # Hero manifest
├── specs/                      # Feature specifications
├── AGENTS.md                   # Agent context file
├── LICENSE                     # Apache 2.0
└── README.md                   # Project README
```

### 2. Install speckit from the canonical source

Per Spec 003, install speckit templates, scripts, and commands
from the `unbound-force/speckit` canonical repository. Do not
copy files from other hero repos.

### 3. Ratify the hero constitution

Run `/speckit.constitution` to create the hero constitution at
`.specify/memory/constitution.md`. Include the
`parent_constitution` reference:

```markdown
**Parent Constitution**: Unbound Force v1.0.0
```

### 4. Create the hero manifest

Create `.unbound-force/hero.json`:

```json
{
  "name": "my-hero",
  "display_name": "My Hero",
  "role": "tester",
  "version": "0.1.0",
  "description": "Brief description of the hero",
  "repository": "https://github.com/unbound-force/my-hero",
  "parent_constitution_version": "1.0.0",
  "artifacts_produced": [],
  "artifacts_consumed": [],
  "opencode_agents": [],
  "opencode_commands": [],
  "dependencies": []
}
```

### 5. Run the contract validation

```bash
bash scripts/validate-hero-contract.sh /path/to/my-hero
```

The script checks all required elements and reports pass/fail
for each.

## Running Contract Validation

The validation script checks a repository against the Hero
Interface Contract:

```bash
# From the unbound-force meta repo:
bash scripts/validate-hero-contract.sh /path/to/hero-repo

# Expected output:
# Hero Interface Contract Validation
# ===================================
# Repository: /path/to/hero-repo
# Contract version: 1.0.0
#
# Required Checks:
#   [PASS] .specify/memory/constitution.md exists
#   [PASS] .specify/templates/ exists and populated
#   [PASS] .specify/scripts/bash/ exists and populated
#   [PASS] .opencode/ exists
#   [PASS] .opencode/command/ exists
#   [PASS] specs/ exists
#   [PASS] AGENTS.md exists
#   [PASS] LICENSE exists
#   [PASS] README.md exists
#   [PASS] .unbound-force/hero.json exists
#   [PASS] hero.json is valid JSON
#   [PASS] hero.json contains required fields
#   [PASS] constitution contains parent_constitution ref
#
# Optional Checks:
#   [WARN] .opencode/agents/ not found (no agents provided)
#   [PASS] .github/workflows/ exists
#
# Overall: PASS (13/13 required, 1/2 optional)
```

## Producing Artifacts in Envelope Format

When a hero produces an artifact for consumption by other
heroes, it wraps the payload in the standard envelope:

```json
{
  "hero": "gaze",
  "version": "1.2.0",
  "timestamp": "2026-02-25T14:30:00Z",
  "artifact_type": "quality-report",
  "schema_version": "1.0.0",
  "context": {
    "branch": "main",
    "commit": "abc123def",
    "backlog_item_id": "BLI-042",
    "correlation_id": "550e8400-e29b-41d4-a716-446655440000"
  },
  "payload": {
    "summary": {
      "total_functions": 42,
      "avg_coverage": 85.3,
      "avg_crap": 4.2
    },
    "functions": [],
    "recommendations": []
  }
}
```

The envelope fields are defined by Spec 002 (this contract).
The payload schema is defined by Spec 009 (Shared Data Model).
The `context.correlation_id` links related artifacts across
workflow stages (e.g., a quality report and the review verdict
for the same PR).

## Consuming Artifacts

When a hero encounters an artifact:

1. **Check `artifact_type`**: If unrecognized, ignore
   gracefully (optional warning log).
2. **Check `schema_version`**: Compare MAJOR version against
   what the consumer supports.
   - Same MAJOR: proceed (minor/patch are backward compatible).
   - Different MAJOR: reject with version mismatch warning.
3. **Parse `payload`**: Use the type-specific schema to
   validate and extract data.
4. **Read `context`**: Use branch, commit, and correlation_id
   for traceability.

## OpenCode Agent and Command Naming

### Agent naming convention

```text
{hero-name}-{agent-function}.md

Examples:
  gaze-reporter.md         # Gaze's reporting agent
  divisor-guard.md         # The Divisor's Guard persona
  divisor-architect.md     # The Divisor's Architect persona
  divisor-adversary.md     # The Divisor's Adversary persona
  muti-mind-po.md          # Muti-Mind's Product Owner agent
```

### Command naming convention

```text
/{hero-name}                    # Primary command
/{hero-name}-{subfunction}      # Secondary commands

Examples:
  /gaze                         # Gaze's main command
  /classify-docs                # Gaze's doc classification
  /review-council               # The Divisor's review
  /muti-mind                    # Muti-Mind's main command
```

### MCP server naming convention (if applicable)

```text
{hero-name}-mcp                 # MCP server name
{hero-name}_{tool_name}         # MCP tool names

Examples:
  gaze-mcp                      # Gaze's MCP server
  gaze_analyze                  # Gaze analyze tool
  gaze_quality                  # Gaze quality tool
```

## Existing Repository Compliance

### Gaze (as of 2026-02-25)

| Check | Status | Notes |
|-------|--------|-------|
| `.specify/memory/constitution.md` | PASS | v1.0.0, ratified |
| `.specify/templates/` | PASS | 6 templates present |
| `.specify/scripts/bash/` | PASS | 5 scripts present |
| `.opencode/` | PASS | Configured |
| `.opencode/command/` | PASS | 12 commands (9 speckit + 3 hero) |
| `specs/` | PASS | 5 specs |
| `AGENTS.md` | PASS | Populated |
| `LICENSE` | PASS | Apache 2.0 |
| `README.md` | PASS | Populated |
| `.unbound-force/hero.json` | FAIL | Does not exist yet |
| `parent_constitution` ref | FAIL | Constitution predates org constitution |
| Agent naming | WARN | `doc-classifier.md` does not follow `gaze-` prefix convention |

**Remediation**: Create `.unbound-force/hero.json`, add
`parent_constitution` reference to constitution, rename
`doc-classifier.md` to `gaze-doc-classifier.md`.

### Website (as of 2026-02-25)

| Check | Status | Notes |
|-------|--------|-------|
| `.specify/memory/constitution.md` | PASS | v1.0.0, ratified |
| `.specify/templates/` | PASS | 6 templates present |
| `.specify/scripts/bash/` | PASS | 5 scripts present |
| `.opencode/` | PASS | Configured |
| `.opencode/command/` | PASS | 10 commands (9 speckit + 1 hero) |
| `specs/` | PASS | 2 specs |
| `AGENTS.md` | PASS | Populated |
| `LICENSE` | PASS | Apache 2.0 |
| `README.md` | PASS | Populated |
| `.unbound-force/hero.json` | FAIL | Does not exist yet |
| `parent_constitution` ref | FAIL | Constitution predates org constitution |
| Agent naming | PASS | Only reviewer agents (divisor-prefixed pattern) |

**Remediation**: Create `.unbound-force/hero.json`, add
`parent_constitution` reference to constitution.
