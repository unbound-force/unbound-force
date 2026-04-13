# Contract: Testing Strategy

**Branch**: `029-sandbox-cde-lifecycle` | **Date**: 2026-04-13

## Coverage Strategy

Per Constitution Principle IV (Testability), all
components MUST be testable in isolation without
external services.

### Unit Tests (primary)

All tests use injected dependencies — no real Podman,
no real Che, no real containers, no network access.

| Function | Test approach | Key assertions |
|----------|--------------|----------------|
| `ResolveBackend()` | Inject `LookPath`, `Getenv`, config | Correct backend selection per matrix |
| `PodmanBackend.Create()` | Inject `ExecCmd` | Volume create, container run, source copy, health check |
| `PodmanBackend.Start()` | Inject `ExecCmd` | Persistent resume vs. ephemeral fallback |
| `PodmanBackend.Stop()` | Inject `ExecCmd` | Persistent stop (preserve volume) vs. ephemeral rm |
| `PodmanBackend.Destroy()` | Inject `ExecCmd` | Container rm + volume rm, idempotent |
| `PodmanBackend.Status()` | Inject `ExecCmd` with mock JSON | Correct parsing of persistent workspace state |
| `CheBackend.Create()` | Inject `ExecCmd` (chectl) or `HTTPClient` (REST) | Devfile check, workspace creation, endpoint discovery |
| `CheBackend.Start()` | Inject `ExecCmd` or `HTTPClient` | Workspace start via chectl or REST |
| `CheBackend.Stop()` | Inject `ExecCmd` or `HTTPClient` | Workspace stop via chectl or REST |
| `CheBackend.Destroy()` | Inject `ExecCmd` or `HTTPClient` | Workspace delete via chectl or REST |
| `CheBackend.Status()` | Inject `ExecCmd` or `HTTPClient` | Correct parsing of Che workspace state |
| `CheBackend.Attach()` | Inject `ExecInteractive` | Correct endpoint URL construction |
| `projectName()` | Pure function | Sanitization, edge cases (special chars, empty) |
| `LoadConfig()` | Inject `ReadFile` | YAML parsing, defaults, env var override |
| `FormatWorkspaceStatus()` | Pure function | Output format for both backends |

### Backward Compatibility Tests

These tests verify that Spec 028 behavior is preserved:

| Test | Assertion |
|------|-----------|
| `TestStart_EphemeralMode` | `uf sandbox start` without prior `create` uses ephemeral mode |
| `TestStop_EphemeralMode` | `uf sandbox stop` in ephemeral mode removes container |
| `TestAttach_Unchanged` | `uf sandbox attach` works with both persistent and ephemeral |
| `TestExtract_Unchanged` | `uf sandbox extract` works with both persistent and ephemeral |
| `TestStatus_EphemeralFallback` | `uf sandbox status` shows Spec 028 format for ephemeral |

### Test Naming Convention

Per AGENTS.md: `TestXxx_Description`

**New tests**:
- `TestResolveBackend_AutoPodman`
- `TestResolveBackend_AutoChe`
- `TestResolveBackend_ExplicitPodman`
- `TestResolveBackend_ExplicitChe`
- `TestResolveBackend_CheNotConfigured`
- `TestResolveBackend_UnknownBackend`
- `TestPodmanCreate_HappyPath`
- `TestPodmanCreate_AlreadyExists`
- `TestPodmanCreate_VolumeCreateFails`
- `TestPodmanCreate_WithDemoPorts`
- `TestPodmanStart_PersistentResume`
- `TestPodmanStart_EphemeralFallback`
- `TestPodmanStop_PersistentPreservesVolume`
- `TestPodmanStop_EphemeralRemoves`
- `TestPodmanDestroy_HappyPath`
- `TestPodmanDestroy_NoWorkspace`
- `TestPodmanDestroy_RunningWorkspace`
- `TestPodmanStatus_PersistentRunning`
- `TestPodmanStatus_PersistentStopped`
- `TestCheCreate_WithChectl`
- `TestCheCreate_WithRestAPI`
- `TestCheCreate_NoDevfile`
- `TestCheCreate_Unreachable`
- `TestCheStart_WithChectl`
- `TestCheStop_WithChectl`
- `TestCheDestroy_WithChectl`
- `TestCheStatus_Running`
- `TestCheAttach_EndpointURL`
- `TestProjectName_Simple`
- `TestProjectName_SpecialChars`
- `TestProjectName_Empty`
- `TestLoadConfig_HappyPath`
- `TestLoadConfig_Missing`
- `TestLoadConfig_EnvOverride`
- `TestFormatWorkspaceStatus_Podman`
- `TestFormatWorkspaceStatus_Che`
- `TestFormatWorkspaceStatus_WithDemoEndpoints`

### Coverage Targets

- **internal/sandbox/**: ≥ 80% line coverage (overall)
  - `backend.go`: ≥ 90% (interface + resolver, pure logic)
  - `podman.go`: ≥ 80% (orchestration with error paths)
  - `che.go`: ≥ 75% (external API interaction, more
    error paths)
  - `workspace.go`: ≥ 90% (pure logic: naming, config)
  - `detect.go`: ≥ 90% (unchanged from Spec 028)
  - `config.go`: ≥ 90% (extended, still mostly pure)
  - `sandbox.go`: ≥ 75% (orchestration, backward compat)

### Test Isolation

- All tests use `t.TempDir()` for filesystem operations
- No shared mutable state between tests
- No network access (HTTP is injected)
- No real subprocess execution (ExecCmd is injected)
- No real Podman, Che, or container operations
- Config file tests use in-memory YAML via `ReadFile`
  injection

### Integration Testing (out of scope for v1)

Integration tests requiring real Podman containers or
Che instances are out of scope. The unit test strategy
with injected dependencies provides sufficient
confidence per the established codebase pattern.

Future integration test candidates:
- End-to-end `create` → `start` → `stop` → `start` →
  `destroy` with real Podman
- CDE workspace provisioning with a test Che instance
- Bidirectional git sync round-trip

### Regression Tests

- All existing Spec 028 tests MUST pass (FR-018)
- `TestScaffoldOutput_*` tests are NOT affected (no new
  scaffold assets)
- `go test -race -count=1 ./...` MUST pass
