# Contract: Testing Strategy

**Branch**: `028-sandbox-command` | **Date**: 2026-04-12

## Coverage Strategy

Per Constitution Principle IV (Testability), all components
MUST be testable in isolation without external services.

### Unit Tests (primary)

All tests use injected dependencies — no real Podman, no
real containers, no network access.

| Function | Test approach | Key assertions |
|----------|--------------|----------------|
| `DetectPlatform()` | Inject `ExecCmd` returning mock `getenforce` output | Correct SELinux detection per platform |
| `buildRunArgs()` | Pure function, no injection needed | Correct flag construction for each mode/platform combo |
| `Start()` | Inject `LookPath`, `ExecCmd`, `HTTPGet` | Prerequisite checks, container start, health polling, attach |
| `Stop()` | Inject `ExecCmd` | Correct podman stop/rm sequence, idempotent on missing |
| `Attach()` | Inject `LookPath`, `ExecInteractive` | OpenCode check, correct attach URL |
| `Extract()` | Inject `ExecCmd`, `Stdin` | Patch generation, review display, confirmation prompt, git am on confirm |
| `Status()` | Inject `ExecCmd` returning mock inspect JSON | Correct parsing of container state |
| `isContainerRunning()` | Inject `ExecCmd` | True/false for running/stopped/missing |
| `waitForHealth()` | Inject `HTTPGet` with delayed success | Timeout behavior, retry logic |

### Test Naming Convention

Per AGENTS.md: `TestXxx_Description`

Examples:
- `TestStart_PodmanMissing`
- `TestStart_AlreadyRunning`
- `TestStart_DetachMode`
- `TestStart_IsolatedMount`
- `TestStart_DirectMount`
- `TestStart_SELinuxVolumeFlag`
- `TestStop_NoContainer`
- `TestStop_RunningContainer`
- `TestAttach_NoContainer`
- `TestExtract_NoChanges`
- `TestExtract_UserDeclines`
- `TestExtract_DirectModeWarning`
- `TestDetectPlatform_MacOSArm64`
- `TestDetectPlatform_FedoraSELinux`
- `TestDetectPlatform_FedoraNoSELinux`
- `TestBuildRunArgs_Isolated`
- `TestBuildRunArgs_Direct`
- `TestBuildRunArgs_SELinux`
- `TestBuildRunArgs_CustomImage`
- `TestWaitForHealth_ImmediateSuccess`
- `TestWaitForHealth_DelayedSuccess`
- `TestWaitForHealth_Timeout`
- `TestStatus_Running`
- `TestStatus_Stopped`
- `TestStatus_NoContainer`

### Coverage Targets

- **internal/sandbox/**: ≥ 80% line coverage
  - `detect.go`: ≥ 90% (pure logic, easy to test)
  - `config.go`: ≥ 90% (pure configuration building)
  - `sandbox.go`: ≥ 75% (orchestration with error paths)
- **cmd/unbound-force/sandbox.go**: Tested via existing
  Cobra command patterns (flag parsing, delegation)

### Test Isolation

- All tests use `t.TempDir()` for filesystem operations
- No shared mutable state between tests
- No network access (HTTP health check is injected)
- No real subprocess execution (ExecCmd is injected)
- No real Podman or container operations

### Integration Testing (out of scope for v1)

Integration tests that actually start containers are
out of scope for this spec. They would require Podman
installed in CI and would be slow. The unit test
strategy with injected dependencies provides sufficient
confidence per the established codebase pattern.

### Regression Tests

- `TestScaffoldOutput_*` tests in `scaffold_test.go` are
  NOT affected (no new scaffold assets)
- Existing `go test -race -count=1 ./...` MUST pass
  (FR-020)
