## Why

The Unbound Force constitution mandates spec-driven development,
coverage strategy, and branch protection — but enforcement today
is purely social. Agents and humans can skip phases, omit specs,
or push directly to main without automated resistance. Issue #514
identifies this gap: there is no headless, machine-readable
command that verifies constitution compliance at each workflow
phase.

The FullSend BYOA increments (Inc3–Inc5) need a
`validation_loop.script` that blocks write-capable harnesses
from proceeding when constitution checks fail. The same command
must be runnable as a local pre-commit or pre-PR hook.

## What Changes

Add a `uf gate` subcommand to the `unbound-force` CLI that
performs phase-specific constitution compliance checks and exits
non-zero when any check fails.

The command accepts `--phase <specify|plan|implement|review|pr>`
and `--format json` flags. Each phase runs a specific set of
checks derived from the constitution:

- **specify**: Verifies spec artifacts exist in the expected
  location (openspec or speckit).
- **plan**: Verifies plan artifact exists and spec is present.
- **implement**: Verifies spec + plan + tasks exist and are
  approved. Verifies coverage strategy is present
  (Constitution IV).
- **review**: Verifies implementation is complete per task
  checklist. Verifies review council has run.
- **pr**: Verifies review PASS exists. Verifies not on main
  branch (no direct commits).

## Capabilities

### New Capabilities
- `uf gate`: Headless constitution compliance gate command
- `--phase`: Select which workflow phase to check
- `--format json`: Machine-readable JSON output with structured
  check results, pass/fail status, and exit code semantics
- `--dir`: Target directory (defaults to cwd)
- Exit code 0: all checks pass. Exit code 1: one or more checks
  fail.

### Modified Capabilities
- None

### Removed Capabilities
- None

## Impact

- **cmd/unbound-force/**: New `gate.go` file with cobra
  subcommand, following the existing doctor/setup pattern
  (params struct, `runGate` function, `newGateCmd` constructor)
- **internal/gate/**: New package with check logic, organized
  by phase. Each check is a function returning a structured
  result.
- **FullSend harnesses**: Will wire `uf gate --phase <phase>
  --format json` as `validation_loop.script` (separate change)
- **Pre-commit/pre-PR hooks**: Users can add `uf gate` to
  their hook scripts (documentation change)

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This command produces self-describing JSON artifacts with
structured check results. It operates without runtime coupling
to other heroes — it reads filesystem artifacts (specs, plans,
tasks) and git state to make pass/fail determinations.

### II. Composability First

**Assessment**: PASS

The gate command is part of the `unbound-force` CLI and works
standalone. It does not require any hero to be deployed. It
checks for the presence of artifacts regardless of which hero
or human produced them.

### III. Observable Quality

**Assessment**: PASS

The `--format json` flag produces machine-parseable output with
per-check results, overall pass/fail, and phase metadata. The
JSON schema will be stable across minor versions.

### IV. Testability

**Assessment**: PASS

The gate command follows the existing CLI testability pattern:
`runGate(params)` with `io.Writer` injection. Each phase check
is a pure function that examines filesystem state and returns
a result struct. All checks are testable in isolation using
`t.TempDir()` with scaffolded artifact trees.

### V. Security by Default

**Assessment**: PASS

The gate command is read-only — it inspects filesystem and git
state but makes no mutations. Input validation applies to the
`--phase` and `--format` flags (enum validation). No external
network access, no secrets handling. Minimal process
execution limited to `git` for branch detection (see
design D5), using `exec.Command` with explicit argument
separation — no shell invocation.
