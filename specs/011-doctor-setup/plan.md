# Implementation Plan: Doctor and Setup Commands

**Branch**: `011-doctor-setup` | **Date**: 2026-03-21 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/011-doctor-setup/spec.md`

## Summary

Add two new commands to the `unbound` CLI: `doctor` (diagnose
environment health with environment-aware install hints) and
`setup` (automated tool chain installation through detected
version managers). Doctor checks 7 groups (Detected Environment,
Core Tools, Swarm Plugin, Scaffolded Files, Hero Availability,
MCP Server Config, Agent/Skill Integrity), embeds `swarm doctor`
output, and supports colored text or JSON output. Setup installs
OpenCode, Gaze, Swarm, configures `opencode.json`, initializes
`.hive/`, and runs `unbound init` -- all using the developer's
existing package managers (goenv, nvm, Homebrew, etc.).

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: `github.com/spf13/cobra` (CLI),
`github.com/charmbracelet/lipgloss` (colored output, promote
from indirect to direct), `gopkg.in/yaml.v3` (frontmatter
parsing)
**Storage**: N/A (reads filesystem and subprocess output,
writes only to `opencode.json` in setup)
**Testing**: Standard library `testing` package, `t.TempDir()`
for filesystem tests, injected `lookPath` and `execCmd`
functions for subprocess mocking
**Target Platform**: macOS (primary), Linux (supported)
**Project Type**: CLI (extends existing `unbound` binary)
**Performance Goals**: Doctor completes all checks in < 5
seconds on a typical developer machine (excluding network
for setup installs)
**Constraints**: 10-second timeout on `swarm doctor`
subprocess; setup MUST NOT use `sudo`; all output to stdout
(doctor) or stderr (progress messages in setup)
**Scale/Scope**: 8 binary checks, 7 check groups, ~50 files
validated in agent/skill integrity

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after
Phase 1 design.*

### I. Autonomous Collaboration — PASS

Doctor and setup operate independently. Doctor produces a
self-describing Report struct (JSON output includes all
metadata: check names, severities, install hints, versions).
No runtime coupling with any hero -- doctor only reads
existing artifacts and checks binary presence.

### II. Composability First — PASS

Both commands are independently useful. Doctor works without
setup (diagnosis without automation). Setup works without
doctor (automated install without prior diagnosis). Doctor
does not require any hero to be installed -- it reports
what's missing. Setup gracefully handles any combination of
missing tools.

### III. Observable Quality — PASS

Doctor produces JSON output (`--format=json`) as the
machine-parseable format alongside colored text. The Report
struct includes provenance (which manager detected each
tool, version, path). JSON output uses stable field names
suitable for tooling consumption.

### IV. Testability — PASS

All external dependencies (binary lookup, subprocess
execution, filesystem access) are injected via function
fields in the Options struct, following the existing pattern
in `heroes.go`. Tests use `t.TempDir()` for filesystem
isolation. No external services or network access required
for testing.

**Coverage strategy**: Unit tests for each check group
function, format tests for text and JSON output, integration
test for the full `Run()` pipeline with injected
dependencies. Target: 90%+ line coverage for
`internal/doctor/` and `internal/setup/`.

## Project Structure

### Documentation (this feature)

```text
specs/011-doctor-setup/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── cli-schema.md   # Command schema
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/unbound/
└── main.go              # Add newDoctorCmd(), newSetupCmd()

internal/
├── doctor/
│   ├── doctor.go        # Run(Options) (*Report, error)
│   ├── checks.go        # Per-group check functions
│   ├── environ.go       # DetectEnvironment() logic
│   ├── models.go        # Report, CheckGroup, CheckResult,
│   │                    # DetectedEnvironment, Severity types
│   ├── format.go        # FormatText(), FormatJSON()
│   └── doctor_test.go   # Unit tests
├── setup/
│   ├── setup.go         # Run(Options) error
│   └── setup_test.go    # Unit tests
└── orchestration/
    └── heroes.go        # (existing, reused unchanged)
```

**Structure Decision**: Two new packages under `internal/`:
`doctor` for diagnosis and `setup` for installation. This
follows the existing pattern where each domain has its own
package (`metrics/`, `impediment/`, `coaching/`, etc.).
The `doctor` package is larger because it contains the
environment detection, all check groups, and two formatters.
The `setup` package is smaller -- it reuses `doctor`'s
`DetectEnvironment()` and delegates to subprocess calls.

## Complexity Tracking

No constitution violations requiring justification.
