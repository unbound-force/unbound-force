# Implementation Plan: The Divisor Architecture (PR Reviewer Council)

**Branch**: `005-the-divisor-architecture` | **Date**: 2026-03-19 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-the-divisor-architecture/spec.md`

## Summary

The Divisor is the PR Reviewer Council hero — an AI-driven code
review framework with dynamic persona discovery, convention packs
for language adaptation, and project-aware context loading. It is
distributed as part of the existing `unbound` binary (no separate
repo), with `unbound init` deploying Divisor agents alongside all
other scaffold files and `unbound init --divisor` deploying only
the Divisor subset. The implementation extends the existing scaffold
engine with: new `divisor-*.md` agent assets, a `/review-council`
command, convention pack files in `.opencode/divisor/packs/`, a
`--divisor` subset flag, and a `--lang` language detection flag.

## Technical Context

**Language/Version**: Go 1.24+ (CLI/scaffold engine), Markdown (agents, packs, commands)
**Primary Dependencies**: `github.com/spf13/cobra` (CLI), `embed.FS` (asset embedding), `github.com/charmbracelet/log` (logging)
**Storage**: Filesystem only (embedded assets deployed to target directory)
**Testing**: Standard library `testing` package, `-race -count=1`, drift detection tests
**Target Platform**: macOS, Linux (cross-compiled via GoReleaser)
**Project Type**: CLI tool (extending existing `unbound` binary)
**Performance Goals**: N/A (one-shot scaffold operation; agent execution is OpenCode-bound)
**Constraints**: No external network calls during scaffold; all assets compiled into binary
**Scale/Scope**: 12 new embedded files (5 agents, 1 command, 3 canonical convention packs, 3 custom convention pack stubs)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Autonomous Collaboration — PASS

The Divisor communicates through well-defined artifacts:
- Persona agents produce structured verdicts (findings with
  severity, category, file, line, description, recommendation).
- The `/review-council` command aggregates verdicts into a council
  decision — a self-describing Markdown report.
- Convention packs are standalone files consumed at review time.
- No runtime coupling between personas — each runs independently
  via the Task tool.
- JSON artifact envelope deferred to Spec 009 (SHOULD, not MUST).
  Markdown report is self-describing with provenance metadata.

### II. Composability First — PASS

- The Divisor is independently usable: `unbound init --divisor`
  deploys only Divisor files. No other hero required.
- Convention packs are the extension point — users add
  `custom_rules[]` without modifying the pack itself.
- Dynamic persona discovery means users can add/remove personas
  freely by adding/removing `divisor-*.md` files.
- When deployed alongside Gaze, The Divisor produces review
  findings that complement Gaze's static analysis (additive value).
- `unbound init` (full scaffold) includes Divisor alongside
  speckit/openspec — combining produces additive value without
  mandatory dependencies.

### III. Observable Quality — PASS

- Review reports include structured findings with severity,
  category, and file/line references — machine-parseable Markdown.
- Discovery summary lists invoked and absent personas.
- Iteration history tracks what was found and fixed per round.
- Provenance metadata: PR URL, review timestamp, convention pack
  used, persona versions.
- JSON artifact format deferred to Spec 009 but Markdown report
  meets the "minimum" bar for structured output.

### IV. Testability — PASS

- **Coverage strategy**: Unit tests for scaffold engine changes
  (Options, filtering, language detection), drift detection tests
  for new embedded assets, integration tests for `--divisor` and
  `--lang` flag behavior.
- **Isolation**: All tests use `t.TempDir()`. No shared state.
  Agent/command Markdown files are static — their "testing" is
  structural (drift detection) and behavioral (manual review
  council runs against sample PRs).
- **Coverage targets**: New Go code must maintain existing coverage
  levels. Drift detection catches embedded/source divergence.
- **Ratchets**: CI enforces `-race -count=1`. Drift detection tests
  fail the build if embedded assets don't match canonical sources.

**Gate Result**: ALL FOUR PRINCIPLES PASS. Proceed to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/005-the-divisor-architecture/
├── spec.md              # Feature specification (clarified)
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── scaffold-cli.md  # CLI contract for --divisor/--lang flags
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
# Divisor scaffold assets (new embedded files)
internal/scaffold/assets/
├── opencode/
│   ├── agents/
│   │   ├── divisor-guard.md         # NEW: Guard persona
│   │   ├── divisor-architect.md     # NEW: Architect persona
│   │   ├── divisor-adversary.md     # NEW: Adversary persona
│   │   ├── divisor-sre.md           # NEW: SRE persona
│   │   ├── divisor-testing.md       # NEW: Testing persona
│   │   ├── reviewer-guard.md        # EXISTING: kept for migration
│   │   ├── reviewer-architect.md    # EXISTING: kept for migration
│   │   ├── reviewer-adversary.md    # EXISTING: kept for migration
│   │   └── reviewer-sre.md          # EXISTING: kept for migration
│   ├── command/
│   │   └── review-council.md        # NEW: embedded (was non-embedded)
│   └── divisor/
│       └── packs/
│           ├── go.md                # NEW: Go convention pack (tool-owned)
│           ├── go-custom.md         # NEW: Go custom rules stub (user-owned)
│           ├── typescript.md        # NEW: TypeScript convention pack (tool-owned)
│           ├── typescript-custom.md # NEW: TypeScript custom rules stub (user-owned)
│           ├── default.md           # NEW: language-agnostic default (tool-owned)
│           └── default-custom.md    # NEW: default custom rules stub (user-owned)
└── ...existing assets unchanged...

# Scaffold engine changes
internal/scaffold/
├── scaffold.go          # MODIFIED: add Divisor/Lang to Options,
│                        #   filtering logic, language detection
└── scaffold_test.go     # MODIFIED: new asset paths, subset tests,
                         #   lang detection tests, drift detection

# CLI changes
cmd/unbound/
└── main.go              # MODIFIED: add --divisor and --lang flags

# Live canonical sources (new Divisor files)
.opencode/
├── agents/
│   ├── divisor-guard.md             # NEW canonical source
│   ├── divisor-architect.md         # NEW canonical source
│   ├── divisor-adversary.md         # NEW canonical source
│   ├── divisor-sre.md               # NEW canonical source
│   └── divisor-testing.md           # NEW canonical source
├── command/
│   └── review-council.md            # MODIFIED: scan divisor-*.md
└── divisor/
    └── packs/
        ├── go.md                    # NEW canonical source (tool-owned)
        ├── go-custom.md             # NEW canonical source (user-owned)
        ├── typescript.md            # NEW canonical source (tool-owned)
        ├── typescript-custom.md     # NEW canonical source (user-owned)
        ├── default.md               # NEW canonical source (tool-owned)
        └── default-custom.md        # NEW canonical source (user-owned)
```

**Structure Decision**: Extending the existing `internal/scaffold/`
package and `cmd/unbound/` CLI. No new Go packages needed. The
Divisor's implementation is primarily Markdown assets embedded via
`go:embed`, with ~200-400 lines of new Go code for filtering,
language detection, and CLI flags. Convention packs and agent files
live alongside existing scaffold assets in the same `embed.FS`.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
<!-- scaffolded by unbound vdev -->
