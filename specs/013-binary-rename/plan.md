# Implementation Plan: Binary Rename

**Branch**: `013-binary-rename` | **Date**: 2026-03-22 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/013-binary-rename/spec.md`

## Summary

Rename the `unbound` CLI binary to `unbound-force` with
a `uf` symlink alias to resolve a name collision with
the NLnet Labs Unbound DNS resolver. This is a
cross-cutting rename that touches the Go source
directory, Cobra root command, Makefile, GoReleaser
config, scaffold assets, doctor/setup hint strings,
living documentation, and cross-repo references (gaze,
website, homebrew-tap).

The approach is a directory rename (`cmd/unbound/` →
`cmd/unbound-force/`) so `go install` produces the
correct binary name, combined with a Makefile `install`
target that creates the `uf` symlink, and a GoReleaser
config that produces both names in Homebrew.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: `github.com/spf13/cobra`
(CLI), `github.com/charmbracelet/log` (logging),
`github.com/charmbracelet/lipgloss` (terminal styling)
**Storage**: N/A
**Testing**: Standard library `testing` package,
`go test -race -count=1`
**Target Platform**: macOS, Linux (CLI binary)
**Project Type**: CLI
**Performance Goals**: N/A (rename only, no behavior
changes)
**Constraints**: Breaking change -- all downstream
references must be updated. Completed specs are not
modified (historical records).
**Scale/Scope**: ~65 files in meta repo, ~11 files in
gaze, ~12 files in website, ~2 files in homebrew-tap.
Primarily string replacements with a few structural
changes (directory rename, Makefile, GoReleaser).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check
after Phase 1 design.*

### I. Autonomous Collaboration -- PASS

The rename does not affect inter-hero communication.
Artifacts are still produced and consumed via the same
paths and formats. The binary name is a CLI entry point,
not an inter-hero protocol.

### II. Composability First -- PASS

The rename improves composability by eliminating the
name collision. A developer can now install the Unbound
Force CLI alongside the NLnet Labs Unbound DNS resolver
without conflict. Both `unbound-force` and `uf` work
independently. No hero depends on the binary name for
its core function.

### III. Observable Quality -- PASS

The binary produces the same JSON and text output
regardless of whether invoked as `unbound-force` or
`uf`. The Cobra root command's `Use` field changes but
the output format, provenance metadata, and artifact
schemas are unaffected.

### IV. Testability -- PASS

All existing tests continue to work after updating
string assertions. The test file
(`cmd/unbound-force/main_test.go`, after Phase 1 rename)
tests the `runInit` function directly via params structs
-- it does not shell out to the binary by name.

**Coverage strategy**:
- **Unit tests (update)**: ~13 existing assertions
  across `main_test.go`, `scaffold_test.go`, and
  `doctor_test.go` reference `"unbound "` strings and
  must be updated to `"unbound-force "` or `"uf "`.
- **Regression tests (new)**: 3 new tests to permanently
  guard the rename requirements:
  `TestRootCmd_HelpOutput` (FR-004 alias in help),
  `TestScaffoldOutput_NoBareUnboundReferences`
  (FR-015/SC-003 scaffold content sweep),
  `TestDoctorHints_NoBareUnboundReferences` (FR-006
  hint string sweep).
- **Integration**: Build binary and verify `--help`
  output matches contract. `make install` produces both
  binaries.
- **Coverage target**: Maintain existing coverage
  percentage. No new code paths introduced -- this is
  a rename, not new functionality.

## Project Structure

### Documentation (this feature)

```text
specs/013-binary-rename/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── contracts/           # Phase 1 output
│   └── cli-schema.md
├── quickstart.md        # Phase 1 output
└── checklists/
    └── requirements.md  # Already created
```

### Source Code (repository root)

```text
cmd/unbound-force/          # Renamed from cmd/unbound/
├── main.go                 # Cobra root Use field updated
└── main_test.go            # Same tests, same directory

internal/doctor/
├── checks.go               # Hint strings updated
└── format.go               # Any "unbound" display text

internal/setup/
└── setup.go                # "unbound init" references

internal/scaffold/assets/   # Embedded .md files with CLI refs
├── opencode/agents/*.md    # "unbound init --divisor" etc.
└── ...

Makefile                    # New install target
.goreleaser.yaml            # Binary name + formula + symlink
AGENTS.md                   # Living doc references
README.md                   # Living doc references
```

**Structure Decision**: No new packages or directories.
This is a rename of the existing `cmd/unbound/`
directory and string replacements across existing files.
The Makefile gains one new target (`install`). No other
structural changes.

## Complexity Tracking

No constitution violations. No complexity justifications
needed.
