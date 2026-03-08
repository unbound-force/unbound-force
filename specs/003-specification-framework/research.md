# Research: Specification Framework

**Spec**: 003-specification-framework
**Date**: 2026-03-08

## R1: Distribution Mechanism

**Decision**: Go binary with `embed.FS` scaffold pattern,
distributed via Homebrew cask and `go install`. Follows
the exact same architecture as the Gaze project
(`unbound-force/gaze`).

**Rationale**: Gaze already established a proven scaffold
pattern: Go's `embed.FS` compiles all distributable files
directly into the binary at build time. `unbound init`
extracts embedded files into the target repo's directory
structure. This provides a self-contained, cross-platform
distribution vehicle with zero runtime dependencies beyond
the binary itself. GoReleaser handles cross-compilation,
release artifacts, and Homebrew cask publishing. The
existing `unbound-force/homebrew-tap` infrastructure
supports this immediately.

**Alternatives considered**:

| Alternative | Rejected Because |
|-------------|-----------------|
| Bash copy script | Extra runtime dependency on curl/fetch; no binary distribution; manual checksum tracking; not standardized with other Unbound Force tools |
| Git submodule | Cannot install to multiple directories; confusing for contributors |
| npm package | Requires Node.js dependency for Bash/Markdown files; architecturally inappropriate for Go projects |
| Separate repo (`unbound-force/speckit`) | This repo is the canonical source; no separate repo needed |

**Architecture (matching Gaze pattern)**:

```text
cmd/unbound/
+-- main.go                    # Cobra CLI entry point

internal/scaffold/
+-- scaffold.go                # Core scaffold logic
+-- scaffold_test.go           # Tests (incl. drift detection)
+-- assets/                    # Embedded files (go:embed)
    +-- specify/
    |   +-- templates/         # 6 Speckit templates
    |   +-- scripts/bash/      # 5 Speckit scripts
    +-- opencode/
    |   +-- command/           # 10 OpenCode commands
    |   +-- agents/            # 1 agent file
    +-- openspec/
        +-- schemas/
        |   +-- unbound-force/ # Custom OpenSpec schema
        +-- config.yaml        # Default OpenSpec config
```

**File ownership model (from Gaze)**:

| Category | Ownership | On re-run behavior |
|----------|-----------|-------------------|
| Templates | User-owned | Skip if exists |
| Scripts | User-owned | Skip if exists |
| Commands (speckit.*) | Tool-owned | Overwrite if content differs |
| Agents | User-owned | Skip if exists |
| OpenSpec schema | Tool-owned | Overwrite if content differs |
| OpenSpec config | User-owned | Skip if exists |

**Version marker**: Each scaffolded file gets an HTML
comment marker inserted after YAML frontmatter:
```html
<!-- scaffolded by unbound v1.0.0 -->
```

**Distribution channels**:
- `brew install unbound-force/tap/unbound`
- `go install github.com/unbound-force/unbound-force/cmd/unbound@latest`
- Build from source: `go build ./cmd/unbound`

**Release pipeline**: GoReleaser v2 with:
- Cross-platform builds (darwin/amd64, darwin/arm64,
  linux/amd64, linux/arm64)
- CGO_ENABLED=0 for static binaries
- ldflags for version/commit/date injection
- Auto-publish Homebrew cask to `unbound-force/homebrew-tap`
- Optional macOS code signing and notarization

## R2: OpenSpec Custom Schema Approach

**Decision**: Fork the built-in `spec-driven` schema to
create an `unbound-force` custom schema, then customize
templates and add constitution context injection.

**Rationale**: Forking preserves the proven `spec-driven`
DAG (proposal -> specs + design -> tasks -> apply) while
allowing customization of templates and instructions. This
avoids building a schema from scratch and benefits from
OpenSpec's built-in schema validation.

**Alternatives considered**:

| Alternative | Rejected Because |
|-------------|-----------------|
| Config-only (no custom schema) | Cannot customize templates to include constitution alignment section; rules can suggest but not structure the output |
| Schema from scratch | Unnecessary work; the `spec-driven` DAG is already well-designed for our workflow |
| Adding a `constitution-check` artifact to the DAG | Over-engineering; constitution alignment is better as a required section within the proposal template than a separate artifact |

**Schema structure**:

```text
openspec/schemas/unbound-force/
+-- schema.yaml                # Forked from spec-driven
+-- templates/
    +-- proposal.md            # Adds Constitution Alignment section
    +-- spec.md                # Adds RFC 2119 language guidance
    +-- design.md              # Adds constitution compliance note
    +-- tasks.md               # Standard task template
```

**Constitution injection mechanism**:

1. **Template level**: The `proposal.md` template includes a
   mandatory "Constitution Alignment" section with three
   subsections (one per principle) requiring PASS/N/A
   assessment.

2. **Config level**: `openspec/config.yaml` `context` field
   includes the constitution text (or a summary of the three
   principles with reference to the full document). The
   `rules.proposal` field reinforces the alignment
   requirement.

3. **Runtime**: OpenSpec injects `context` and `rules` into
   every artifact prompt. The agent generating the proposal
   sees both the template structure and the constitution
   content.

## R3: OpenSpec Integration with OpenCode

**Decision**: Use `openspec init --tools opencode` to
generate OpenCode-compatible skill and command files. The
core profile provides 4 commands: `/opsx:propose`,
`/opsx:explore`, `/opsx:apply`, `/opsx:archive`.

**Rationale**: OpenSpec has native OpenCode support. The
generated files use the `.opencode/skills/` and
`.opencode/commands/` (plural) directories, which do not
conflict with Speckit's `.opencode/command/` (singular)
directory.

**Key findings**:

- OpenSpec skills go to `.opencode/skills/openspec-*/SKILL.md`
- OpenSpec commands go to `.opencode/commands/opsx-*.md`
- Speckit commands are in `.opencode/command/speckit.*.md`
- **No file-level collision** between the two systems
- Command prefixes are distinct: `opsx-*` vs `speckit.*`
- OpenSpec uses `openspec/` directory; Speckit uses
  `.specify/` directory -- no overlap

**OpenSpec CLI prerequisites**:

- Node.js >= 20.19.0
- Install: `npm install -g @fission-ai/openspec@latest`
- The `openspec` CLI is needed for `init`, `schema fork`,
  and runtime `status`/`instructions` queries from skills.

## R4: Boundary Guidelines Design

**Decision**: Document boundary guidelines as a decision
matrix with clear criteria and a default heuristic. The
guidelines are advisory (SHOULD), not enforced.

**Rationale**: The boundary between strategic and tactical
work is inherently fuzzy. Hard enforcement would create
friction without proportional benefit. Advisory guidelines
with a simple heuristic ("when in doubt, start with OpenSpec
and escalate if scope grows") are more practical.

**Criteria matrix**:

| Criterion | Speckit (Strategic) | OpenSpec (Tactical) |
|-----------|:------------------:|:-------------------:|
| User stories | >= 3 | < 3 |
| Cross-repo impact | Yes | No |
| Constitution changes | Always | Never |
| New hero architecture | Always | Never |
| New inter-hero artifact types | Always | Never |
| Bug fix | Never | Always |
| Single-repo maintenance | Never | Always |
| Refactoring (non-architectural) | Rarely | Usually |

**Default heuristic**: "When in doubt, start with OpenSpec.
If the scope grows beyond 3 stories or crosses repo
boundaries, escalate to Speckit by extracting the proposal
into a new numbered spec directory under `specs/`."

## R5: Current File Inventory

**Decision**: Documented the current canonical file inventory
for manifest generation.

**Findings**:

| Category | Directory | Files | Lines |
|----------|-----------|------:|------:|
| Templates | `.specify/templates/` | 6 | 588 |
| Scripts | `.specify/scripts/bash/` | 5 | 1,490 |
| Commands | `.opencode/command/` | 10 | 1,477 |
| Agents | `.opencode/agents/` | 1 | 133 |
| **Total** | | **22** | **3,688** |

Files not yet created (will be created by this spec):
- `.specify/config.yaml` (project-specific configuration)
- `openspec/` directory tree (OpenSpec integration)
- `cmd/unbound/main.go` (CLI entry point)
- `internal/scaffold/` (scaffold package with embedded assets)
- `.goreleaser.yaml` (release configuration)
