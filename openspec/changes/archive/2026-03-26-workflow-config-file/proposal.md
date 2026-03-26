## Why

Spec 016 (Autonomous Define) requires the human to
explicitly pass `--define-mode=swarm` or use
`/workflow seed` every time they start a workflow.
There is no "set it once for the project" mechanism.
A team lead who wants all workflows to use autonomous
define must communicate this verbally to their team --
there is no project-level configuration.

The execution mode map is currently passed at workflow
creation time (per-workflow), not read from a config
file. This means the default (define=human) cannot be
changed at the project level without code changes.

## What Changes

Add a `.unbound-force/config.yaml` file that stores
default execution mode overrides and the spec review
flag. `uf init` scaffolds the file with commented-out
defaults. `Start()` reads the config and uses it as
the base, with CLI flags taking precedence.

Merge order:
1. `StageExecutionModeMap()` hardcoded defaults
2. `.unbound-force/config.yaml` project-level overrides
3. CLI flag per-workflow overrides (highest priority)

## Capabilities

### New Capabilities
- `workflow-config-file`: Project-level workflow
  configuration at `.unbound-force/config.yaml`
- `config-aware-start`: `Start()` reads config file
  and merges execution mode overrides
- `scaffold-config`: `uf init` creates the config file
  with commented-out defaults (user-owned, not
  overwritten on re-scaffold)

### Modified Capabilities
- `Start()`: Now reads project config before creating
  workflow

### Removed Capabilities
- None

## Impact

- `internal/orchestration/config.go` -- NEW: config
  reader
- `internal/orchestration/config_test.go` -- NEW: tests
- `internal/orchestration/engine.go` -- `Start()` loads
  and merges config
- `internal/orchestration/engine_test.go` -- config-aware
  Start tests
- `internal/scaffold/scaffold.go` -- create config file
  during init
- `internal/scaffold/scaffold_test.go` -- test config
  creation
- `cmd/unbound-force/main_test.go` -- file count update

## Constitution Alignment

### I. Autonomous Collaboration
**Assessment**: PASS -- the config file is an artifact
on disk, not runtime coupling.

### II. Composability First
**Assessment**: PASS -- the config file is optional.
Missing file = all defaults (human mode). No hero
depends on it.

### III. Observable Quality
**Assessment**: PASS -- the config is human-readable
YAML and machine-parseable.

### IV. Testability
**Assessment**: PASS -- `LoadWorkflowConfig()` accepts
a directory path, testable with `t.TempDir()`.
