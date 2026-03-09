# Implementation Plan: Specification Framework

**Branch**: `003-specification-framework` | **Date**: 2026-03-08 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-specification-framework/spec.md`

## Summary

Establish a unified two-tier specification framework: Speckit
for strategic architectural work and OpenSpec for tactical
changes. This repo (`unbound-force/unbound-force`) serves as
the canonical source for all framework artifacts. Distribution
is via a Go binary (`unbound`) using Go's `embed.FS` scaffold
pattern -- identical to the architecture used by Gaze. All
distributable files are compiled into the binary at build time.
`unbound init` extracts them into the target repo. GoReleaser
handles cross-platform releases and Homebrew cask publishing.
A custom `unbound-force` OpenSpec schema (forked from
`spec-driven`) injects constitution alignment into every
proposal. Boundary guidelines are documented as advisory
criteria.

## Technical Context

**Language/Version**: Go 1.24+ (CLI binary), Markdown
(templates, commands, specs), YAML (schemas, configuration)
**Primary Dependencies**: Cobra (CLI framework),
`embed` (Go stdlib, file embedding), OpenSpec CLI
(`@fission-ai/openspec` v1.2.0+, npm), OpenCode (agent
runtime), GoReleaser v2 (release pipeline)
**Storage**: Filesystem only -- embedded Markdown/YAML/Bash
files extracted to target repo directories
**Testing**: Go tests (`go test ./...`), including drift
detection test (embedded assets match canonical source files)
**Target Platform**: macOS darwin/amd64 + darwin/arm64,
Linux linux/amd64 + linux/arm64 (static binaries, CGO_ENABLED=0)
**Project Type**: CLI tool / framework distribution
**Performance Goals**: Scaffold in under 5 seconds (SC-002)
**Constraints**: Self-contained binary with zero runtime
dependencies; Node.js required only for OpenSpec CLI
**Scale/Scope**: 32 canonical files (22 Speckit + 6 OpenSpec
+ 4 agents) embedded in binary, distributed to N consumer
repositories (currently 3: Gaze, Website, unbound-force)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after
Phase 1 design.*

### Pre-Design Assessment

#### I. Autonomous Collaboration -- PASS

The specification framework produces well-defined artifacts
(specs, plans, tasks, proposals) that are self-describing
Markdown files with metadata. Both Speckit and OpenSpec
create artifacts that can be consumed by any hero without
synchronous interaction. The framework does not introduce
runtime coupling -- all communication is through files
published to well-known locations (`specs/`, `openspec/`).

#### II. Composability First -- PASS

Speckit functions independently without OpenSpec (FR-012,
edge case: "Speckit MUST function independently"). OpenSpec
is additive and optional. Each tier delivers value alone.
The framework works in any Git repository, not just Unbound
Force hero repos (FR-012). Extension points
(`.specify/config.yaml`, `openspec/config.yaml`) allow
customization without modification.

#### III. Observable Quality -- PASS

The scaffold produces observable outputs: every scaffolded
file includes a version marker comment
(`<!-- scaffolded by unbound v1.0.0 -->`), providing
provenance tracking. The scaffold result reports created,
skipped, updated, and overwritten files. A drift detection
test ensures embedded assets match canonical source files.
The custom OpenSpec schema enforces constitution alignment
as traceable, reviewable evidence in proposals.

### Post-Design Re-Assessment

All three principles remain PASS after design phase.
Additional confirmation:

- **Autonomous Collaboration**: The Go binary is a
  self-contained distribution vehicle. No synchronous
  coordination or network access needed at scaffold time.
- **Composability First**: The two tiers (Speckit, OpenSpec)
  are independently deployable.   The `unbound` binary works
  standalone via `go install` or Homebrew -- no other heroes
  required. The boundary is advisory, not a hard coupling.
- **Observable Quality**: Version markers in every scaffolded
  file provide provenance. The drift detection test ensures
  embedded assets stay in sync. Scaffold results are
  categorized (created/skipped/updated/overwritten).

## Project Structure

### Documentation (this feature)

```text
specs/003-specification-framework/
+-- spec.md               # Feature specification
+-- plan.md               # This file
+-- research.md           # Phase 0: distribution, schema,
|                         #   integration, boundary research
+-- data-model.md         # Phase 1: entity definitions
+-- quickstart.md         # Phase 1: setup guide
+-- contracts/
|   +-- installer-cli.md  # Installer CLI contract
|   +-- openspec-schema.md # Custom schema contract
+-- checklists/
|   +-- requirements.md   # Spec quality checklist
+-- tasks.md              # Phase 2: task list (via /speckit.tasks)
```

### Source Code (repository root)

```text
# NEW: Go CLI binary (scaffold tool)
cmd/unbound/
+-- main.go                       # Cobra CLI entry point
go.mod                            # Go module definition
go.sum                            # Dependency checksums
.goreleaser.yaml                  # Release configuration

# NEW: Scaffold package (follows Gaze pattern)
internal/scaffold/
+-- scaffold.go                   # Core scaffold logic
+-- scaffold_test.go              # Tests + drift detection
+-- assets/                       # Embedded via go:embed
    +-- specify/
    |   +-- templates/            # 6 Speckit templates
    |   +-- scripts/bash/         # 5 Speckit scripts
    +-- opencode/
    |   +-- command/              # 10 OpenCode commands
    |   +-- agents/               # 4 agent files
    +-- openspec/
        +-- schemas/
        |   +-- unbound-force/    # Custom OpenSpec schema
        +-- config.yaml           # Default OpenSpec config

# Canonical source files (used by devs AND embedded)
.specify/
+-- memory/
|   +-- constitution.md           # Org constitution v1.0.0
+-- templates/                    # 6 Speckit templates
+-- scripts/bash/                 # 5 Speckit scripts
+-- config.yaml                   # NEW: project configuration

.opencode/
+-- command/                      # 10 commands (9 speckit + 1)
+-- agents/                       # 4 agent files

openspec/                         # NEW: OpenSpec directory
+-- specs/                        # Living behavior contracts
+-- changes/                      # Active tactical changes
+-- schemas/
|   +-- unbound-force/            # Custom schema
+-- config.yaml                   # OpenSpec configuration

scripts/
+-- validate-hero-contract.sh     # (already exists)
```

**Structure Decision**: Go CLI binary with embedded assets,
following the Gaze project's scaffold pattern. Canonical
source files live at their standard locations in the repo
root (`.specify/`, `.opencode/`). Copies of these files are
maintained under `internal/scaffold/assets/` for embedding.
A drift detection test ensures the two copies stay in sync.

**Drift detection**: `TestEmbeddedAssetsMatchSource` (from
Gaze pattern) verifies that every file under
`internal/scaffold/assets/` is byte-identical to the
corresponding canonical file. If drift is detected, the
test prints a `cp` command to fix it.

## Complexity Tracking

No constitution violations to justify. All three principles
pass cleanly.
