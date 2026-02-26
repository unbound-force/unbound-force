# Research: Hero Interface Contract

**Spec**: [spec.md](spec.md) | **Date**: 2026-02-25

## Research Task 1: Artifact Envelope Ownership

### Decision

Spec 002 defines the artifact envelope conceptually — the
required fields, their types, their purpose, and the versioning
semantics. Spec 009 (Shared Data Model) provides the formal
JSON Schema file. This split ensures the contract establishes
the "shape" while the data model provides the machine-readable
validation artifact.

### Rationale

The constitution (line 72-74) says: "Inter-hero communication
MUST use the artifact envelope format defined by the Hero
Interface Contract." This makes the Hero Interface Contract
the authoritative source for what the envelope contains. But
Spec 009's entire purpose is to produce JSON Schemas for all
inter-hero artifacts.

The split resolves this naturally:

- **Spec 002 owns the concept**: What fields exist, what they
  mean, what rules govern them (e.g., `schema_version` uses
  semver, `timestamp` uses ISO 8601, `payload` is type-specific).
- **Spec 009 owns the schema**: The `.schema.json` file that
  validators consume. This is a mechanical translation of the
  concept into JSON Schema draft 2020-12.

This avoids duplicate schema definitions while honoring the
constitution's language about the Hero Interface Contract
defining the format.

### Alternatives Considered

- **Full ownership in Spec 002**: Define both the concept and
  the JSON Schema. Rejected because it would duplicate Spec
  009's schema registry and create a maintenance burden — two
  places to update when the envelope changes.
- **Full ownership in Spec 009**: The Hero Interface Contract
  just says "use the artifact envelope" without defining its
  fields. Rejected because the constitution explicitly assigns
  ownership to the Hero Interface Contract, and downstream specs
  (004-007) need to reference the envelope fields during their
  design phase before Spec 009 is finalized.

## Research Task 2: Speckit Distribution Scope

### Decision

FR-007 is scoped to state the requirement that hero repos
install speckit from the canonical source. The distribution
mechanism itself (git submodule, npm package, CLI tool, etc.)
is defined by Spec 003 (Speckit Framework Centralization).

### Rationale

Spec 003 is entirely dedicated to speckit centralization and
distribution. It defines:

- The canonical repository (`unbound-force/speckit`)
- Installation and upgrade mechanisms
- Extension points for project-specific customization
- Drift detection

If Spec 002 also defined a distribution mechanism, the two
specs would conflict or duplicate. The Hero Interface Contract's
role is to say "heroes MUST use speckit from the canonical
source" — Spec 003 defines what that source is and how to
consume it.

### Alternatives Considered

- **Keep full scope in FR-007**: Spec 002 defines its own
  distribution mechanism. Rejected because this duplicates
  Spec 003 and creates a governance conflict about who owns
  the mechanism.
- **Remove FR-007 entirely**: Rejected because the contract
  must state that speckit is required — even if the how is
  deferred. Without FR-007, the contract has no speckit
  requirement at all.

## Research Task 3: MCP Server Interface Scope

### Decision

FR-014 is scoped to a SHOULD with minimal naming conventions
only: `{hero-name}-mcp` for the MCP server name, hero-prefixed
tool names. Full MCP interface requirements are deferred to
hero-specific architecture specs (004-007) since MCP adoption
is optional and hero-specific.

### Rationale

MCP (Model Context Protocol) support is optional for heroes.
Only heroes that expose programmatic capabilities to external
tools need MCP servers. Currently:

- Gaze does not have an MCP server (it's a CLI tool).
- The Divisor, Muti-Mind, Cobalt-Crush, and Mx F are spec-only
  and have not yet decided on MCP integration.

Defining detailed MCP interface requirements in the contract
would be premature. The contract should establish naming
conventions to prevent collisions (consistent with the agent
naming convention in FR-005) and defer implementation details
to each hero's architecture spec.

### Alternatives Considered

- **Full MCP specification**: Define request/response schemas,
  capability negotiation, authentication, etc. Rejected as
  premature — MCP standards are still evolving and no hero
  currently uses MCP.
- **Remove FR-014**: Rejected because MCP is a likely future
  integration point for heroes, and establishing naming
  conventions early prevents future collisions.

## Research Task 4: Validation Tooling Approach

### Decision

A bash script (`scripts/validate-hero-contract.sh`) that
checks directory structure, required file existence, constitution
presence, and manifest schema validity. The script outputs
structured results (pass/fail per check) and an overall
pass/fail status.

### Rationale

The validation script must work in any hero repo without
requiring additional dependencies. A bash script satisfies
this constraint:

- **No dependencies**: Bash is available on macOS and Linux
  (the two target platforms for Unbound Force development).
- **Structural checks**: File existence, directory existence,
  and basic content validation (e.g., does the constitution
  contain `parent_constitution`) are well-suited to bash.
- **JSON Schema validation**: For the hero manifest, the script
  can optionally call `ajv` (Node.js) or `jsonschema` (Python)
  if available, or fall back to basic JSON syntax checking
  with `python3 -m json.tool` or `jq`.

The script categorizes checks as:

- **Required**: Must pass for the repo to be contract-compliant.
  Missing required elements are errors.
- **Optional**: Recommended but not blocking. Missing optional
  elements are warnings.

### Alternatives Considered

- **Markdown checklist only**: A manual checklist document.
  Rejected because manual validation is not reproducible and
  violates Principle III (Observable Quality — quality claims
  must be backed by automated evidence).
- **Go CLI tool**: A Go binary for validation. Rejected because
  it introduces a build dependency and Go requirement for repos
  that may use other languages.
- **Both script and checklist**: Has merit but adds maintenance
  burden. The script output serves as the checklist — each
  check is a line item with pass/fail status.

## Research Task 5: Hero Manifest Location

### Decision

The hero manifest lives at `.unbound-force/hero.json` — a
dedicated directory for Unbound Force swarm metadata that is
separate from speckit (`.specify/`) and OpenCode (`.opencode/`).

### Rationale

Three location candidates were evaluated:

| Location | Pros | Cons |
|----------|------|------|
| `.unbound-force/hero.json` | Clear ownership, separate from other tools, extensible for future swarm metadata | New directory, slightly more nesting |
| `hero-manifest.json` (root) | Visible, easy to find | Clutters root, no namespace for future swarm files |
| `.specify/hero.json` | Reuses existing directory | Conflates speckit (generic tool) with Unbound Force (org-specific metadata) |

`.unbound-force/hero.json` was selected because:

1. It creates a clear namespace for Unbound Force-specific
   metadata that may grow beyond the manifest (e.g., swarm
   configuration, hero-to-hero integration settings).
2. It separates concerns: `.specify/` is for speckit (usable
   in non-hero repos), `.opencode/` is for OpenCode, and
   `.unbound-force/` is for swarm membership metadata.
3. It follows the hidden-directory convention used by other
   tools (`.github/`, `.vscode/`, `.opencode/`).

### Alternatives Considered

See table above. Root-level `hero-manifest.json` was the
closest alternative but rejected to avoid root clutter and
to provide a namespace for future files.

## Research Task 6: Gaze Output Format Gap Analysis

### Decision

Gaze's current JSON output schemas (analysis-report and
quality-report) do not use the artifact envelope format. This
is a known gap, not a blocker. The contract defines the target
format; Gaze will adopt the envelope in a future spec when
Spec 009 provides the formal schemas.

### Rationale

Gaze currently produces two JSON output formats:

1. **Analysis report**
   (`gaze analyze --format=json`): Top-level fields are
   `version` and `results[]`. No envelope metadata (hero,
   timestamp, artifact_type, schema_version).
2. **Quality report**
   (`gaze quality --format=json`): Top-level fields are
   `quality_reports[]` and `quality_summary`. No envelope.

Both schemas are embedded as Go string constants in
`internal/report/schema.go` and use JSON Schema draft 2020-12
with IDs under `https://github.com/unbound-force/gaze/`.

The gap is clear: Gaze's output would need to be wrapped in
the artifact envelope to be consumable by other heroes per the
contract. However, this change:

- Requires the formal JSON Schema from Spec 009 to exist first.
- Is a breaking change to Gaze's output format (MAJOR version
  bump).
- Should be planned as a Gaze-specific spec (e.g.,
  `006-envelope-adoption`), not mandated by the contract
  without a migration path.

The contract establishes the requirement; Gaze's adoption is
tracked as a downstream remediation item.

### Alternatives Considered

- **Require immediate adoption**: The contract mandates that
  Gaze wrap its output in the envelope now. Rejected because
  Spec 009 (which defines the envelope schema) is not yet
  finalized — premature dependency.
- **Exempt existing output**: Gaze's current format is
  grandfathered and never needs to change. Rejected because
  this would permanently prevent inter-hero artifact exchange
  for Gaze's primary output.
