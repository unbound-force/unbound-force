# Implementation Plan: Replicator Migration

**Branch**: `024-replicator-migration` | **Date**: 2026-04-06 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/024-replicator-migration/spec.md`

## Summary

Replace the Node.js Swarm plugin (`opencode-swarm-plugin`) with
Replicator (a standalone Go binary) across three CLI subsystems:
`uf setup`, `uf init`, and `uf doctor`. This eliminates the bun
prerequisite, reduces setup steps from 15 to 12, migrates
`opencode.json` from a `plugin` array to an `mcp.replicator`
entry, and updates all install hints and health checks. The
migration follows established patterns already proven in the
codebase: `installGaze()` for Homebrew installation,
`checkDewey()` for MCP binary health checks, Dewey MCP entry
for `opencode.json` configuration, and `dewey init` delegation
for sub-tool initialization.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: `github.com/spf13/cobra` (CLI), `embed.FS` (scaffold), `github.com/charmbracelet/log` (logging), `encoding/json` (opencode.json manipulation)
**Storage**: Filesystem only (`opencode.json`, `.hive/`)
**Testing**: Standard library `testing` package, `t.TempDir()` for isolation, `-race -count=1`
**Target Platform**: macOS (primary), Linux
**Project Type**: CLI (meta-repository tooling)
**Performance Goals**: N/A (CLI tool, sub-second operations)
**Constraints**: All external dependencies injected for testability; no global state
**Scale/Scope**: 6 production files modified, ~50 test assertions updated/added

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Autonomous Collaboration — PASS

Replicator is a standalone Go binary that communicates via MCP
protocol (stdio). It produces and consumes artifacts through
well-defined file formats (`.hive/` directory, MCP messages).
No runtime coupling with other heroes — Replicator is discovered
via `exec.LookPath` and invoked as a subprocess, same as Dewey.

### II. Composability First — PASS

Replicator is independently installable via Homebrew. When absent,
`uf init` skips the MCP entry and `uf doctor` reports informational
status (not failure). This matches the existing Dewey pattern where
absence is gracefully handled. No hero requires Replicator as a
hard prerequisite.

### III. Observable Quality — PASS

`uf doctor` produces machine-parseable JSON output (`--format=json`)
with the Replicator check group following the same `CheckGroup` /
`CheckResult` structure as all other groups. All check results
include severity, message, and install hints.

### IV. Testability — PASS

All new functions follow the established dependency injection
pattern: `Options.LookPath`, `Options.ExecCmd`,
`Options.ExecCmdTimeout`, `Options.ReadFile`. No real subprocess
calls in tests. Coverage strategy: unit tests for each new/modified
function with injected mocks, following the existing test patterns
in `setup_test.go` and `doctor_test.go`.

**Coverage Strategy**:
- **Unit tests**: Every new function (`installReplicator`,
  `runReplicatorSetup`, `checkReplicator`, updated
  `configureOpencodeJSON`, updated `installOpenSpec`) tested
  with injected `LookPath`/`ExecCmd` mocks
- **Integration tests**: `TestRun_*` tests in `setup_test.go`
  and `doctor_test.go` verify end-to-end flow with mock
  dependencies
- **Regression tests**: Existing `TestScaffoldOutput_*` tests
  verify no stale references remain
- **Target**: Maintain existing coverage level (no regressions)

## Project Structure

### Documentation (this feature)

```text
specs/024-replicator-migration/
├── plan.md              # This file
├── research.md          # Phase 0: pattern analysis
├── data-model.md        # Phase 1: data model changes
├── quickstart.md        # Phase 1: implementation quickstart
├── contracts/           # Phase 1: behavioral contracts
│   ├── setup-replicator.md
│   ├── init-replicator.md
│   └── doctor-replicator.md
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── setup/
│   ├── setup.go         # Remove installSwarmPlugin/ensureBun/runSwarmSetup/initializeHive
│   │                    # Add installReplicator/runReplicatorSetup
│   │                    # Update installOpenSpec (npm-only)
│   │                    # Update step count 15→12
│   └── setup_test.go    # Update/add tests for new functions
├── scaffold/
│   ├── scaffold.go      # Replace plugin array logic with MCP entry
│   │                    # Add replicator init delegation
│   │                    # Add legacy plugin migration
│   └── scaffold_test.go # Update opencode.json tests
├── doctor/
│   ├── checks.go        # Replace checkSwarmPlugin() with checkReplicator()
│   ├── doctor.go        # Update check group list
│   ├── environ.go       # Update install hints (remove bun/npm swarm refs)
│   └── doctor_test.go   # Update/add tests for new checks
opencode.json            # Replace plugin array with mcp.replicator
```

**Structure Decision**: This is a modification-only change to
existing packages. No new packages or directories are created.
The existing layered architecture (`cmd/` → `internal/`) is
preserved. All changes follow established patterns within each
package.

## Constitution Re-Check (Post-Design)

*Re-evaluated after Phase 1 design artifacts were produced.*

### I. Autonomous Collaboration — PASS (unchanged)

Design confirms: Replicator communicates via MCP protocol
(stdio), discovered via `exec.LookPath`, invoked as subprocess.
No runtime coupling. Artifact format (`.hive/`) is self-describing.

### II. Composability First — PASS (unchanged)

Design confirms: `configureOpencodeJSON()` skips Replicator
entry when binary absent. `checkReplicator()` returns Warn (not
Fail) when missing. `initSubTools()` captures errors as warnings.
Legacy plugin migration runs independently of Replicator presence.

### III. Observable Quality — PASS (unchanged)

Design confirms: `checkReplicator()` returns `CheckGroup` with
`CheckResult` entries, serializable to JSON via `--format=json`.
All results include severity, message, and install hints.

### IV. Testability — PASS (unchanged)

Design confirms: All new functions use injected dependencies
(`LookPath`, `ExecCmd`, `ExecCmdTimeout`, `ReadFile`). Coverage
strategy defined: unit tests per function, integration tests per
subsystem, regression tests for stale references. No external
services or network calls in tests.

**Post-design verdict**: All four principles PASS. No violations.
No complexity tracking needed.

## Complexity Tracking

> No constitution violations. All changes follow established patterns.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| (none)    | —          | —                                   |
