# Implementation Plan: Divisor Council Refinement

**Branch**: `019-divisor-council-refinement` | **Date**: 2026-03-30 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/019-divisor-council-refinement/spec.md`

## Summary

Refine the Divisor Council review system by removing legacy
`reviewer-*.md` files from the scaffold, de-duplicating
cross-persona review responsibilities, harmonizing severity
definitions via a shared convention pack, qualifying all FR
references with spec numbers, integrating Hivemind learning
lookups, and adding `golangci-lint` + `govulncheck` to CI
and `/review-council` Phase 1a.

This is primarily a **Markdown-only change** affecting agent
files, convention packs, the review-council command, and the
CI workflow. The Go production code changes are limited to:
(1) removing 4 legacy `reviewer-*.md` files from scaffold
assets and updating file counts/paths in tests, (2) adding
legacy file detection + warning to `uf init`, (3) adding a
new `severity.md` convention pack to scaffold assets, and
(4) adding `golangci-lint`/`govulncheck` install steps to
`uf setup`.

## Technical Context

**Language/Version**: Go 1.24+ (scaffold engine, setup), Markdown (agents, packs, commands), YAML (CI workflow)
**Primary Dependencies**: `github.com/spf13/cobra` (CLI), `embed.FS` (asset embedding), `github.com/charmbracelet/log` (logging)
**Storage**: Filesystem only (Markdown files deployed to target directory, CI workflow YAML)
**Testing**: Standard library `testing` package; `go test -race -count=1 ./...`
**Target Platform**: macOS, Linux (CI: ubuntu-latest)
**Project Type**: CLI tool (meta-repo scaffold engine)
**Performance Goals**: N/A (file scaffolding, not runtime-critical)
**Constraints**: Backward compatible — existing repos with legacy files get a warning, not a deletion
**Scale/Scope**: 5 agent files rewritten, 4 legacy files removed from scaffold, 1 new pack, 1 CI workflow update, 1 command update, scaffold tests updated

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Autonomous Collaboration — PASS

Each Divisor persona remains an independent agent file that
can be invoked standalone. The shared severity convention
pack is consumed as a file artifact, not a runtime
dependency. Personas do not require synchronous interaction
with each other. The Hivemind learning lookup is optional
(graceful degradation when unavailable).

### II. Composability First — PASS

The Divisor agents remain independently deployable via
`uf init --divisor`. The severity pack follows the existing
convention pack ownership model (tool-owned, auto-updated).
The learning loop integration uses `hivemind_find` which is
already an optional MCP tool — no new hard dependencies.
Static analysis tools (`golangci-lint`, `govulncheck`) are
additive CI steps that don't block the review council when
absent.

### III. Observable Quality — PASS (directly improved)

This spec directly improves Observable Quality:
- Static analysis tools provide machine-verifiable evidence
  backing the Adversary's and SRE's quality claims (instead
  of relying solely on LLM-based inspection).
- Harmonized severity definitions ensure consistent,
  calibrated quality assessments across all 5 personas.
- The `/review-council` Phase 1a gate now includes
  `golangci-lint` (non-zero = gate failure), producing
  reproducible, machine-parseable lint output.

### IV. Testability — PASS

All Go code changes are testable in isolation:
- Scaffold asset removal is verified by existing
  `TestAssetPaths_MatchExpected` (updated counts).
- Legacy file detection is testable via `t.TempDir()` with
  pre-created `reviewer-*.md` files.
- The severity pack is a static Markdown file (no runtime
  logic to test).
- CI workflow changes are tested by CI itself.
- Agent file content is verified by existing drift detection
  tests.

## Project Structure

### Documentation (this feature)

```text
specs/019-divisor-council-refinement/
├── plan.md              # This file
├── research.md          # Phase 0: design decisions
├── data-model.md        # Phase 1: persona ownership, severity levels
├── quickstart.md        # Phase 1: verification steps
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
# Markdown files (agent personas, packs, commands)
.opencode/
├── agents/
│   ├── divisor-adversary.md    # Updated: de-dup, severity ref, FR qual, learning loop
│   ├── divisor-architect.md    # Updated: de-dup, severity ref, FR qual, learning loop
│   ├── divisor-guard.md        # Updated: de-dup, severity ref, FR qual, learning loop
│   ├── divisor-sre.md          # Updated: de-dup, severity ref, FR qual, learning loop
│   ├── divisor-testing.md      # Updated: de-dup, severity ref, FR qual, learning loop
│   ├── reviewer-adversary.md   # NOT DELETED (remains on disk, not scaffolded)
│   ├── reviewer-architect.md   # NOT DELETED (remains on disk, not scaffolded)
│   ├── reviewer-guard.md       # NOT DELETED (remains on disk, not scaffolded)
│   ├── reviewer-sre.md         # NOT DELETED (remains on disk, not scaffolded)
│   └── reviewer-testing.md     # NOT DELETED (remains on disk, not scaffolded)
├── command/
│   └── review-council.md       # Updated: severity ref, static analysis in Phase 1a
└── unbound/
    └── packs/
        └── severity.md         # NEW: shared severity definitions (tool-owned)

# CI workflow
.github/workflows/
└── test.yml                    # Updated: add golangci-lint + govulncheck steps

# Go scaffold engine (asset changes)
internal/scaffold/
├── assets/opencode/agents/
│   ├── reviewer-adversary.md   # DELETED from assets
│   ├── reviewer-architect.md   # DELETED from assets
│   ├── reviewer-guard.md       # DELETED from assets
│   └── reviewer-sre.md         # DELETED from assets
├── assets/opencode/unbound/packs/
│   └── severity.md             # NEW asset
├── scaffold.go                 # Updated: legacy file detection, severity pack
└── scaffold_test.go            # Updated: file counts, asset paths, regression tests

# Setup (tool installation)
internal/setup/
└── setup.go                    # Updated: golangci-lint + govulncheck install steps
```

**Structure Decision**: This change modifies existing files
in the established project structure. No new packages or
directories are introduced (the `severity.md` pack goes
into the existing `.opencode/unbound/packs/` directory).
The scaffold asset directory loses 4 files and gains 1.

## Phase 0: Research

See [research.md](research.md) for design decisions on:
1. Agent boundary refactoring approach
2. Severity standard design
3. Hivemind integration pattern
4. golangci-lint configuration strategy
5. Legacy file detection mechanism

## Phase 1: Design

See [data-model.md](data-model.md) for:
- Persona ownership mapping (which persona owns which review dimension)
- Severity level definitions with domain-specific examples

See [quickstart.md](quickstart.md) for:
- Verification steps to confirm the implementation is correct

Contracts directory: **Skipped** — no external API or
inter-hero artifact schema changes. The severity pack is
an internal convention file, not a versioned schema.

## Coverage Strategy

This spec is primarily Markdown content changes. The Go
code changes are limited to scaffold asset management and
setup tool installation. Coverage approach:

- **Unit tests**: Scaffold asset path list
  (`TestAssetPaths_MatchExpected`), legacy file detection
  logic, `isDivisorAsset` updates, `isToolOwned` updates,
  severity pack deployment via `shouldDeployPack`.
- **Drift detection**: `TestEmbeddedAssets_MatchSource`
  ensures scaffold assets match canonical sources.
- **Regression tests**: New test verifying no `reviewer-*`
  files in scaffold output. Existing
  `TestScaffoldOutput_NoBareUnboundReferences` pattern.
- **Integration**: CI workflow changes tested by CI itself
  (golangci-lint and govulncheck run on every PR).
- **No e2e tests**: Agent file content is verified by drift
  detection. Severity definitions are verified by manual
  review and the review council itself.

Coverage target: Maintain existing 80% global threshold.
New Go code (legacy detection, severity pack handling) must
have ≥90% coverage.

## Constitution Re-Check (Post-Design)

*Re-evaluated after Phase 0 research and Phase 1 design.*

### I. Autonomous Collaboration — PASS (confirmed)

The research decision to encode ownership boundaries
directly in each agent file (rather than a shared registry)
strengthens autonomy — each agent is fully self-contained.
The Prior Learnings step uses `hivemind_find` as an
optional, non-blocking lookup. No new inter-agent runtime
coupling introduced.

### II. Composability First — PASS (confirmed)

The severity pack follows the existing convention pack
model: tool-owned, auto-deployed, language-agnostic. It
integrates with `shouldDeployPack` (always true, like
`default.md`). The `uf init --divisor` subset will include
the severity pack alongside existing packs. No new hard
dependencies.

### III. Observable Quality — PASS (confirmed, strengthened)

Post-design analysis confirms:
- `golangci-lint` non-zero exit = gate failure in both CI
  and `/review-council` Phase 1a (machine-verifiable).
- `govulncheck` output feeds directly into Adversary
  findings with specific CVE references (evidence-backed).
- Severity definitions in `severity.md` are structured with
  per-persona example tables (calibrated, reproducible).
- The auto-fix policy boundary (LOW/MEDIUM vs HIGH/CRITICAL)
  is now grounded in shared definitions rather than
  per-persona interpretation.

### IV. Testability — PASS (confirmed)

Post-design analysis confirms all new code paths are
testable:
- Legacy file detection: `filepath.Glob` on `t.TempDir()`
  with pre-created files.
- Severity pack scaffold: `shouldDeployPack("severity.md",
  lang)` returns true for all languages.
- `isToolOwned("severity.md")` returns true.
- `isDivisorAsset("severity.md")` returns true.
- No external services or network access required.

## Complexity Tracking

No constitution violations to justify. All changes align
with the four principles.
