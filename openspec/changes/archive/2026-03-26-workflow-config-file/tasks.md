## 1. Config Reader

- [x] 1.1 Create `internal/orchestration/config.go` with `WorkflowConfig` struct and `LoadWorkflowConfig(dir string) (WorkflowConfig, error)` function. Use `gopkg.in/yaml.v3` for parsing. Return zero-value config when file doesn't exist (no error). Return error when file exists but YAML is malformed. The config path is `filepath.Join(dir, "config.yaml")`.
- [x] 1.2 Promote `gopkg.in/yaml.v3` from indirect to direct dependency in `go.mod` by running `go mod tidy` after importing it.

## 2. Config Tests

- [x] 2.1 Write `TestLoadWorkflowConfig_FileExists` in `internal/orchestration/config_test.go`: create a temp dir with a valid `config.yaml` containing `workflow.execution_modes.define: swarm` and `workflow.spec_review: true`. Verify the returned config has the correct values.
- [x] 2.2 Write `TestLoadWorkflowConfig_FileMissing` in `internal/orchestration/config_test.go`: create an empty temp dir. Verify the returned config has zero values (nil map, false spec_review) and no error.
- [x] 2.3 Write `TestLoadWorkflowConfig_Malformed` in `internal/orchestration/config_test.go`: create a temp dir with invalid YAML content. Verify an error is returned.
- [x] 2.4 Write `TestLoadWorkflowConfig_CommentedOut` in `internal/orchestration/config_test.go`: create a temp dir with the scaffolded config (all values commented out). Verify the returned config has zero values (commented YAML = empty document).

## 3. Start() Integration

- [x] 3.1 Update `Start()` in `internal/orchestration/engine.go`: call `LoadWorkflowConfig(o.WorkflowDir)` before `NewWorkflow()`. Merge config execution modes as the base, then apply CLI overrides on top (CLI wins). Merge spec review with OR logic (config OR CLI). If config load returns error, log a warning and use empty defaults.
- [x] 3.2 Write `TestOrchestrator_Start_ReadsProjectConfig` in `internal/orchestration/engine_test.go`: create a config.yaml in the workflow dir with `define: swarm`, call `Start()` with no CLI overrides, verify the workflow's define stage has `execution_mode=swarm`.
- [x] 3.3 Write `TestOrchestrator_Start_CLIOverridesConfig` in `internal/orchestration/engine_test.go`: create a config.yaml with `define: swarm`, call `Start()` with `overrides={"define": "human"}`, verify CLI wins (define=human).
- [x] 3.4 Write `TestOrchestrator_Start_ConfigMissing_UsesDefaults` in `internal/orchestration/engine_test.go`: no config.yaml, call `Start()` with no overrides, verify all defaults (define=human).
- [x] 3.5 Write `TestOrchestrator_Start_ConfigMalformed_WarnsAndUsesDefaults` in `internal/orchestration/engine_test.go`: create malformed config.yaml, call `Start()`, verify workflow creates successfully with defaults (warning logged, no error).

## 4. Scaffold Integration

- [x] 4.1 Add config file creation to `initSubTools()` in `internal/scaffold/scaffold.go`: if `.unbound-force/config.yaml` doesn't exist, create it with the commented-out default content. Create `.unbound-force/` directory if needed. This is user-owned -- skip if file already exists.
- [x] 4.2 Write `TestInitSubTools_CreatesWorkflowConfig` in `internal/scaffold/scaffold_test.go`: verify `uf init` creates `.unbound-force/config.yaml` with commented content.
- [x] 4.3 Write `TestInitSubTools_PreservesExistingConfig` in `internal/scaffold/scaffold_test.go`: create a config.yaml with custom content, run scaffold, verify file is NOT overwritten.

## 5. Verify

- [x] 5.1 Run `go build ./...` to verify compilation
- [x] 5.2 Run `go test -race -count=1 ./internal/orchestration/...` to verify orchestration tests
- [x] 5.3 Run `go test -race -count=1 ./internal/scaffold/...` to verify scaffold tests
- [x] 5.4 Run `go test -race -count=1 ./...` to verify full test suite
