# Implementation Plan: Externalize Tool Initialization

**Branch**: `027-externalize-tool-init` | **Date**: 2026-04-11 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/027-externalize-tool-init/spec.md`

## Summary

Externalize Speckit, OpenSpec, and Gaze initialization from
embedded scaffold assets to CLI init command delegation. The
`uf` binary currently embeds 12 Speckit files and 1 OpenSpec
config file that become stale between releases. This change
removes those files from the embedded asset tree and instead
delegates to `specify init`, `openspec init --tools opencode`,
and `gaze init` during `uf init`. The `uf setup` command gains
2 new steps to install `uv` and `specify`. The custom OpenSpec
schema (`openspec/schemas/unbound-force/`) remains embedded
because it is project-specific and not created by any external
CLI.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: `github.com/spf13/cobra` (CLI),
  `embed.FS` (scaffold engine), `github.com/charmbracelet/log`
  (logging)
**Storage**: Filesystem only (embedded assets deployed to target
  directory, sub-tool CLI delegation)
**Testing**: Standard library `testing` package with `-race
  -count=1`
**Target Platform**: macOS (primary), Linux (secondary)
**Project Type**: CLI
**Performance Goals**: N/A (one-shot init command)
**Constraints**: All tool delegations MUST be gated by
  `LookPath` and directory existence checks. Failures MUST NOT
  block the rest of init (Constitution Principle II —
  Composability First).
**Scale/Scope**: 4 files modified (`scaffold.go`,
  `scaffold_test.go`, `setup.go`, `setup_test.go`), 13 embedded
  asset files removed, 2 new setup steps added

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after
Phase 1 design.*

### I. Autonomous Collaboration — PASS

Each tool delegation (`specify init`, `openspec init`,
`gaze init`) is a fire-and-forget subprocess call. No runtime
coupling between `uf` and the delegated tools. Each tool
produces its own files independently. If a tool is absent, the
delegation is skipped — `uf init` continues with the tools that
are available.

### II. Composability First — PASS

This change *strengthens* composability. Each tool owns its own
initialization files. `uf init` works with zero external tools
installed (deploys OpenCode agents, packs, and commands from
embedded assets). Each additional tool (`specify`, `openspec`,
`gaze`) adds value when present but is never required. The
`LookPath` + directory-existence gate pattern (established in
Spec 024 for Replicator) ensures graceful degradation.

### III. Observable Quality — PASS

All sub-tool delegation results are reported via `subToolResult`
structs in the `printSummary()` output, following the
established pattern for Dewey and Replicator init. The step
count in `uf setup` is updated to reflect the new steps.
Machine-parseable output is not affected (no JSON output
changes).

### IV. Testability — PASS

All external dependencies (`LookPath`, `ExecCmd`, `ReadFile`,
`WriteFile`) are already injectable on `scaffold.Options` and
`setup.Options`. New delegation code follows the same injection
pattern. Tests use injected stubs — no real subprocess
execution, no network access, no shared mutable state.

**Coverage strategy**: Unit tests for each new delegation
function. Update `expectedAssetPaths` to reflect removed
assets. Update `knownNonEmbeddedFiles` to reflect files now
created by external tools. Update setup step count assertions.
No integration tests needed — the delegation pattern is
well-established and tested in Specs 017 and 024.

## Project Structure

### Documentation (this feature)

```text
specs/027-externalize-tool-init/
├── plan.md              # This file
├── research.md          # Phase 0: design decisions
├── data-model.md        # Phase 1: data model changes
├── quickstart.md        # Phase 1: implementation guide
├── contracts/           # Phase 1: function contracts
│   ├── scaffold-changes.md
│   └── setup-changes.md
└── tasks.md             # Phase 2 (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── scaffold/
│   ├── scaffold.go          # initSubTools(): add specify, openspec, gaze delegation
│   ├── scaffold_test.go     # Update expectedAssetPaths, knownNonEmbeddedFiles, add delegation tests
│   └── assets/
│       ├── specify/         # REMOVE ENTIRELY (12 files → 0)
│       ├── openspec/
│       │   ├── config.yaml  # REMOVE (created by openspec init)
│       │   └── schemas/
│       │       └── unbound-force/  # KEEP (project-specific, not created by openspec init)
│       └── opencode/        # UNCHANGED
└── setup/
    ├── setup.go             # Add installUV(), installSpecify(), update step count 12→15
    └── setup_test.go        # Add tests for new install functions
```

**Structure Decision**: Existing single-project structure.
Changes are localized to `internal/scaffold/` and
`internal/setup/`. No new packages or directories.

## Complexity Tracking

> No constitution violations to justify.
