## Context

Spec 016 added configurable execution modes to
`NewWorkflow()` via an `overrides` map parameter. But
there is no project-level config file to set defaults.
The human must pass `--define-mode=swarm` every time.

## Goals / Non-Goals

### Goals
- `.unbound-force/config.yaml` stores default execution
  mode overrides and spec review flag
- `Start()` reads it and merges with CLI overrides
- `uf init` scaffolds the file (user-owned)
- Missing file = all defaults (backward compatible)

### Non-Goals
- Storing other configuration (tool paths, Dewey
  settings, etc.)
- Making the config mandatory
- Adding a config editing CLI command

## Decisions

### Config File Format

```yaml
# .unbound-force/config.yaml
# Workflow configuration for Unbound Force hero lifecycle.
# CLI flags (--define-mode, --spec-review) override these.

# workflow:
#   execution_modes:
#     define: swarm
#   spec_review: false
```

Scaffolded with all values commented out. The team lead
uncomments what they want. Commenting is the mechanism
for "use defaults" -- no ambiguity.

### Config Reader

```go
// WorkflowConfig holds project-level workflow defaults.
type WorkflowConfig struct {
    Workflow struct {
        ExecutionModes map[string]string `yaml:"execution_modes"`
        SpecReview     bool              `yaml:"spec_review"`
    } `yaml:"workflow"`
}

// LoadWorkflowConfig reads .unbound-force/config.yaml.
// Returns zero-value config if file doesn't exist.
// Returns error if file exists but is malformed.
func LoadWorkflowConfig(dir string) (WorkflowConfig, error)
```

### Merge Order in Start()

```go
func (o *Orchestrator) Start(
    branch, backlogItemID string,
    overrides map[string]string,
    specReview bool,
) (*WorkflowResult, error) {
    // 1. Load project config (base overrides)
    cfg, err := LoadWorkflowConfig(o.WorkflowDir)
    if err != nil {
        log.Warn("workflow config error, using defaults", "err", err)
    }

    // 2. Merge: config overrides < CLI overrides
    merged := cfg.Workflow.ExecutionModes // project defaults
    for k, v := range overrides {         // CLI wins
        merged[k] = v
    }

    // 3. Spec review: CLI true wins over config
    review := cfg.Workflow.SpecReview || specReview

    // 4. Create workflow with merged overrides
    wf, err := o.NewWorkflow(branch, backlogItemID, merged, review)
    ...
}
```

### Scaffold Integration

The config file is **user-owned** -- `uf init` creates
it only if `.unbound-force/config.yaml` doesn't exist.
It is NOT overwritten on re-scaffold. This matches the
convention pack ownership model (user-owned custom
files are preserved).

The file is created by `initSubTools()` in
`scaffold.go`, after the `.unbound-force/` directory
exists (it's created by the orchestration engine or
by `uf setup`). If the directory doesn't exist yet,
the scaffold creates it.

### YAML Dependency

`gopkg.in/yaml.v3` is already an indirect dependency
(via `charmbracelet/log`). Promoting it to direct.

## Risks / Trade-offs

**Low risk**: The config file is optional. Missing =
all defaults. Malformed = warn and use defaults. The
merge logic is 10 lines of Go.

**User-owned means no auto-updates**: If we add new
config fields in the future, existing config files
won't have them. This is acceptable -- new fields
default to their zero values (false, nil map).
